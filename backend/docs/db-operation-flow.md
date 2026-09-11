# DB初期設定の流れと各ファイルの役割

## 最初に区別すること

| 作業 | Goが動く場所 | RDSへの接続 | AWS CLIの役割 |
| --- | --- | --- | --- |
| マイグレーション | ECSタスクのコンテナ内 | ECSから直接接続。SSM不要 | ローカルからECSタスクを起動し、終了を確認 |
| app_userの作成・権限付与 | ローカルPC | SSMポート転送経由 | RDS情報・管理者Secretの取得と、アプリ用Secretの保存 |

どちらもSQLを実行するのはRDS上のPostgreSQL。SQLを送信するのはGoのDBドライバであり、AWS CLIではない。初期設定時はどちらもdbadminで接続する。作成するapp_userは通常のバックエンドが使うDBユーザーで、アプリのログインユーザーではない。

## 1. マイグレーション：ECSで実行

先にTerraformでリソースを用意し、マイグレーション用イメージをECRへpushしておく。

`````text
【ローカルPC】
make -C backend migrate-rds-up
  → backend/Makefile
  → scripts/ecs-migrate.sh
  → AWS CLIのecs run-taskで単発タスクを起動
      │
      ▼
【AWS / ECS】
ECS基盤がECRからイメージを取得
  → 専用のタスク実行ロールで管理者Secretを取得
  → DB_USER / DB_PASSWORDをコンテナへ注入
  → /app/migrate-rds を起動（cmd/migrate-rdsをビルドした実行ファイル）
      → internal/dbadminのConnectionURL・OpenDatabaseを利用
      → golang-migrateがmigrations/*.up.sqlを読み込む
      → GoのDBドライバでRDSへ直接SQLを送信
      → 終了

【ローカルPC】
scripts/ecs-migrate.shが停止を待ち、終了コード0を確認
`````

SSMを接続する操作も、NATへログインしてコマンドを打つ操作も不要。NATはECR等への外向き通信には使うが、ECSからRDSへのDB接続はVPC内で直接行う。

ECS内でGoソースやshファイルを実行するのではなく、Dockerfileでビルド済みのGo実行ファイルを動かす。ecs-migrate.sh自体はローカルに残る。将来はこの起動処理をCI/CDから呼び出せる。

詳細：[ECSマイグレーションの準備・実行コマンド](rds-migrations.md)

## 2. DBユーザー追加：ローカルからSSM経由で実行

マイグレーションが成功してテーブルができた後、別々のターミナルで次の2段階を実行する。AWS_PROFILE・AWS_REGION・NAT_INSTANCE_ID等の準備は[SSM接続手順](../../infrastructure/docs/db-tunnel.md)を参照。

### ターミナルA：通信経路を開く

`````text
【ローカルPC】
make -C backend db-tunnel
  → backend/Makefile
  → scripts/db-tunnel.sh
  → AWS CLIのssm start-session
  → localhost:15432 → SSM → NAT上のSSM Agent → RDS:5432
`````

この段階ではDBユーザーを作成しない。転送セッションを開いたままにしておく。NAT上でGoコマンドを実行するわけではない。

### ターミナルB：Goコマンドでユーザーを設定する

`````text
【ローカルPC】
make -C backend db-user DB_TUNNEL=127.0.0.1:15432
  → backend/MakefileがRDS用CAを一時ファイルへ取得
  → go run ./cmd/db-user
      │
      ├── internal/dbadmin/admin.go
      │     → AWS CLIでRDS接続先・dbadminのSecretを取得
      │     → 接続URLを作成
      │
      ├── internal/dbadmin/tunnel.go
      │     → Goの接続先をlocalhost:15432へ向ける
      │     → SSM経由でRDSへ接続（RDSの証明書検証は維持）
      │
      ├── cmd/db-user/main.go
      │     → CREATE ROLE / GRANTなどのSQLをDBドライバで送信
      │     → RDS内にapp_userを作成し、テーブルの操作権限を付与
      │
      └── internal/dbadmin/admin.go
            → AWS CLIでapp_userの接続情報をSecrets Managerへ保存
`````

図は担当処理の大枠を示す。実装では既存ユーザー・Secretの確認やトランザクション制御も行う。AWS CLIによるAWS APIへの通信はSSMトンネルを通さず、DB通信だけがトンネルを通る。

詳細：[DBユーザー設定・再実行時の挙動](db-user.md)

## ファイル対応表

パスはbackend/からの相対パス。

| ファイル | どこから使われるか | 担当 |
| --- | --- | --- |
| Makefile | ローカルでmakeを実行 | shやGoコマンドの呼び出し。イメージのビルド・pushもここから開始 |
| scripts/ecs-migrate.sh | Makefile → ローカルのBash | ECS起動、停止待ち、終了コード確認 |
| scripts/db-tunnel.sh | Makefile → ローカルのBash | SSMの通信経路を開く |
| cmd/migrate-rds/main.go | Dockerfileでビルド → ECS内で起動 | 注入された環境変数でDB接続し、マイグレーション・状態確認 |
| cmd/db-user/main.go | Makefile → ローカルのgo run | app_user作成・権限付与・接続情報の保存を進行 |
| internal/dbadmin/admin.go | 上記Goコマンドから関数を呼ぶ | AWS操作・管理者接続・接続URL作成。ECS側が使うのはConnectionURLのみ |
| internal/dbadmin/tunnel.go | 上記Goコマンドから関数を呼ぶ | DB接続を作る。db-userはSSM経由、migrate-rdsは直接接続 |
| migrations/*.up.sql | ECS内のgolang-migrateが読む | テーブル・インデックス・外部キー等の作成 |
| Dockerfileのmigrationステージ | ローカルのイメージビルド時 | 実行ファイル・SQL・RDS用CAをイメージへ格納 |

internal/のGoファイルは単独で起動するコマンドではない。cmd/から関数として呼ばれ、その呼び出し元と同じ場所で動く。admin.goはAWS CLIを呼ぶが、scripts/のshファイルは呼ばない。tunnel.goもSSMセッション自体は作らず、既に開いている転送経路を利用する。

## この資料の範囲

現在のコードの呼び出し関係を整理した資料で、実AWS上の成功を示すものではない。今回はドキュメントのみを変更し、削除されたテストは復元していない。
