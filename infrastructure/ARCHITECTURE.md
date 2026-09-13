# AWSインフラ構成

## 現在のAWSリソース構成

### VPC・サブネットとリソース配置

VPC（10.0.0.0/16）の中に、次の4つのサブネットがある。

| VPC内のサブネット | CIDR          | 現在配置されているリソース                        |
| ----------------- | ------------- | ------------------------------------------------- |
| ├ public-1a       | 10.0.0.0/18   | NATインスタンス（EC2）。外向き通信とSSM接続の中継 |
| ├ public-1c       | 10.0.64.0/18  | なし                                              |
| ├ private-1a      | 10.0.128.0/18 | APIのECSタスク、RDS（app-db）                     |
| └ private-1c      | 10.0.192.0/18 | なし                                              |

### リソースとセキュリティーグループの対応・許可する通信

| AWSリソース       | セキュリティグループ名 | 用途                      | インバウンドルール                                                                                              | アウトバウンドルール                  |
| ----------------- | ---------------------- | ------------------------- | --------------------------------------------------------------------------------------------------------------- | ------------------------------------- |
| EC2：nat-instance | nat-instance           | 外向き通信・SSM接続の中継 | 10.0.128.0/18・10.0.192.0/18から全プロトコル許可<br>EC2 Instance ConnectのAWS管理プレフィックスリストからTCP 22 | 0.0.0.0/0へ全プロトコル許可           |
| ECS：api          | api                    | APIサーバー               | nat-instance SGからTCP 8080（手動API確認）                                                                      | db SGへTCP 5432<br>0.0.0.0/0へTCP 443 |
| ECS：db-migrator  | db-migrator            | DBマイグレーション        | なし                                                                                                            | db SGへTCP 5432<br>0.0.0.0/0へTCP 443 |
| RDS：app-db       | db                     | PostgreSQL                | api・db-migrator・nat-instance SGからTCP 5432                                                                   | なし（許可済み接続への応答は可能）    |

### IAMユーザー・ロール・ポリシー

| 利用者・AWSリソース | IAMユーザー／ロール名                    | 用途                                     | アタッチする権限ポリシー                     | 許可する操作・対象                                                                                   |
| ------------------- | ---------------------------------------- | ---------------------------------------- | -------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| 開発者PC            | x-clone-terraform-stg（IAMユーザー）     | Terraform・AWS CLI・DB管理コマンドの実行 | AdministratorAccess（AWS管理）               | 全AWS操作・全リソース（ポリシー上の許可。SCP等の制約は未確認）                                       |
| ECS：api            | api-task-execution（実行ロール）         | ECS基盤によるコンテナ起動・ログ送信      | api-task-execution（カスタマー管理）         | ECR apiのイメージ取得<br>/ecs/backendへのログ送信<br>db/app_userのSecret取得                         |
| ECS：api            | api-task（タスクロール）                 | APIプログラムのAWS操作用                 | なし                                         | AWS操作権限なし。DBの読み書きはapp_userのSQL権限                                                     |
| ECS：db-migrator    | db-migrator-task-execution（実行ロール） | ECS基盤によるコンテナ起動・ログ送信      | db-migrator-task-execution（カスタマー管理） | ECR db-migratorのイメージ取得<br>/ecs/backend-migrationへのログ送信<br>db/migration_userのSecret取得 |
| ECS：db-migrator    | db-migrator-task（タスクロール）         | マイグレーションプログラムのAWS操作用    | なし                                         | AWS操作権限なし。DB変更はmigration_userのSQL権限                                                     |
| EC2：nat-instance   | nat-ssm（ロール）                        | SSM Agentの管理・通信                    | AmazonSSMManagedInstanceCore（AWS管理）      | SSMへの情報登録・管理用通信。DB操作・Secret取得の権限なし                                            |

### DBユーザー・Secret・実行場所の関係

| DBユーザー     | Secret            | 使う場所                                     | 用途                                                  |
| -------------- | ----------------- | -------------------------------------------- | ----------------------------------------------------- |
| dbadmin        | db/dbadmin        | 開発者PCで実行するDBユーザー初期設定コマンド | 初期設定・ユーザー管理                                |
| migration_user | db/migration_user | ECSタスク：db-migrator                       | テーブル作成・変更、migration.schema_migrationsの管理 |
| app_user       | db/app_user       | ECSサービスで常時起動するAPIタスク           | アプリデータの読み書き                                |

## DBユーザー・Secret・実行場所の設計

DBユーザー登録の初期設定では、次のバックエンド側ファイルを使う。

| ファイル                     | 役割                                                                             |
| ---------------------------- | -------------------------------------------------------------------------------- |
| backend/scripts/db-tunnel.sh | 開発者PCの127.0.0.1:15432から、SSMとnat-instance経由でRDSへ接続できる経路を作る  |
| backend/cmd/db-user/main.go  | dbadminでRDSへ接続し、app_user・migration_userの作成、Secret保存、権限設定を行う |

### DBユーザー：dbadmin（初期設定・ユーザー管理用）

- 開発者PCでbackend/cmd/db-userを実行するときに使う。API・マイグレーションのECSタスクでは使わない。
- migration_userとapp_userを作成する。
- 各ユーザーへ必要な権限を付与する。

### DBユーザー：migration_user（マイグレーション用）

- 単発実行するECSタスクdb-migratorが、DBへ接続するときに使う。
- 現在は開発者がタスクを起動する。将来はCI/CDから同じタスクを起動し、デプロイ時のテーブル作成・変更に使う。
- CI/CDがタスクを起動する権限はIAM、タスク内からDBを変更する権限はmigration_userが担う。
- appデータベースに接続する（CONNECT）。
- publicスキーマを利用し、テーブルを作成する（USAGE・CREATE）。
- migration_userで実行するマイグレーションにより、public配下のアプリテーブルを作成・変更する。
- マイグレーション管理用のschema_migrationsはmigrationスキーマで管理する。

### DBユーザー：app_user（API用）

- ECSサービスで常時動かすAPIタスクapiが、DBへ接続するときに使う。
- アプリ利用者が画面から投稿・フォロー・タイムライン表示を操作すると、APIがapp_userとしてDBを読み書きする。
- appデータベースに接続する（CONNECT）。
- publicスキーマを利用する（USAGE）。
- 存在するusers・posts・followsのデータを取得・追加・更新・削除する（SELECT・INSERT・UPDATE・DELETE）。
- テーブルの作成・構造変更・削除は行わない。

## DBユーザー登録の初期設定

```mermaid
sequenceDiagram
  actor DEV as 開発者
  participant ENV as infrastructure/stg/.env
  participant CMD as backend/cmd/db-user<br/>バックエンドのGo管理コード（開発者PCで実行）
  participant TF as Terraform（開発者PC）
  participant TUN as backend/scripts/db-tunnel.sh
  participant SM as Secrets Manager
  participant RDS as RDS
  DEV->>ENV: 1. dbadminのパスワードを設定
  DEV->>TF: 2. .envを読み込み、apply
  TF->>SM: db/dbadminを作成し、固定名dbadminとパスワードを保存
  TF->>SM: db/app_user・db/migration_userのSecret保存先を作成（値はまだ空）
  TF->>RDS: dbadminと同じパスワードを設定
  DEV->>TUN: 3. SSMポート転送を開始<br/>backend/scripts/db-tunnel.sh
  TUN->>RDS: 開発者PCの127.0.0.1:15432から<br/>nat-instance経由でRDSへ転送
  DEV->>CMD: 4. DBユーザー作成コマンドを実行<br/>backend/cmd/db-user/main.go
  CMD->>SM: IAMユーザーの認証で<br/>db/dbadminを取得
  SM-->>CMD: dbadminのユーザー名・パスワード
  CMD->>CMD: 2ユーザーのパスワードを生成（再実行時は保存済みの値を再利用）
  CMD->>RDS: 5. SSM転送経由でdbadminとしてTLS接続<br/>app_user・migration_userを作成、権限を設定
  CMD->>SM: 6. 各ユーザー名・パスワードを<br/>db/app_user・db/migration_userへ保存
```

実行するファイルと順序：

| 順序 | 実行するもの                    | 役割                                                                                                  |
| ---- | ------------------------------- | ----------------------------------------------------------------------------------------------------- |
| 1    | infrastructure/stg/.env         | dbadmin用のパスワードをローカルで管理する                                                             |
| 2    | infrastructure/stg の Terraform | db/dbadmin・db/app_user・db/migration_userのSecret、RDS、IAMなどを作成する                            |
| 3    | backend/scripts/db-tunnel.sh    | SSMポート転送で、開発者PCの127.0.0.1:15432からRDSへ到達できる経路を作る                               |
| 4    | backend/cmd/db-user/main.go     | dbadminでRDSへ接続し、app_user・migration_userの作成、migrationスキーマ、default privilegesを設定する |

.envで管理する値：

| 値                              | 用途                                              | 理由                                                      |
| ------------------------------- | ------------------------------------------------- | --------------------------------------------------------- |
| TF_VAR_dbadmin_password         | RDSのdbadminパスワード、db/dbadmin Secretの値     | Terraform apply時だけ渡す。コミットしない                 |
| TF_VAR_dbadmin_password_version | dbadminパスワードを意図的に更新するときの世代番号 | password_wo / secret_string_wo の更新タイミングを明示する |

dbadminのSecret値をTerraformのdata sourceで読み込んでRDSに渡す構成にはしない。Secretの中身がTerraform Stateに残る可能性があるため、RDSとSecrets Managerへ同じ値を.envから同時に渡す。

SSMポート転送は、開発者PCからRDSへ一時的に到達するための経路。db-user実行時だけ使い、APIタスクやマイグレーションタスクでは使わない。

## マイグレーション（現在の手動起動と将来のCI/CD）

現在も将来も、db-migratorは処理が終わると終了する単発のECSタスク。将来は起動元を開発者PCからCI/CDへ切り替える。CI/CDは未実装。

```mermaid
flowchart TD
  DEV["現在：開発者が手動で起動"]
  CI["将来：CI/CDが自動で起動"]
  ECS["ECSタスク定義"]
  SM["Secrets Manager<br/>DBの認証情報を保管"]
  TASK["ECSタスク＝実体<br/>マイグレーションを実行"]
  DB["RDS<br/>テーブルを作成・変更"]
  DONE["処理完了後、タスクは終了"]
  DEV -->|"① 起動を要求<br/>起動元のIAM権限"| ECS
  CI -.->|"将来は起動元を差し替え"| ECS
  ECS -->|"② 認証情報を取得<br/>タスク実行ロールの権限"| SM
  SM -->|"DBの認証情報"| ECS
  ECS -->|"③ 環境変数へ渡す<br/>タスクを起動"| TASK
  TASK -->|"④ DBへ接続・SQL実行<br/>マイグレーション用<br/>DBユーザーの権限"| DB
  TASK -->|"⑤"| DONE
```

CI/CDはDBへ直接接続しない。起動元を切り替えても、Secretの取得はECS基盤、DBへの接続・SQL実行はタスク内のGoプログラムが担当する。マイグレーションタスクからRDSへの接続にSSMポート転送は使わない。

## APIからのデータ読み書き

```mermaid
flowchart LR
  USER["利用者のブラウザ<br/>投稿・フォロー・一覧表示"]
  SVC["ECSサービス<br/>APIタスクを常に1つ維持"]
  API["ECS：APIタスク<br/>リクエストを処理"]
  DB["RDS<br/>投稿・ユーザー・フォロー<br/>のデータを保存"]
  AS["Secrets Manager<br/>db/app_user"]
  SVC -->|"① タスク定義に従って起動・維持"| API
  USER -->|"② HTTPリクエスト"| API
  API -->|"③ app_userでSQL実行<br/>TLS・パスワード認証"| DB
  DB -->|"④ データ・処理結果"| API
  API -->|"⑤ HTTPレスポンス"| USER
  AS -.->|"起動時のみ：ECS基盤が<br/>実行ロールで取得し<br/>環境変数へ注入"| API
```

app_userを使うのはAPI。ECSサービスがAPIタスクを常時1つ維持する。利用者はHTTPでAPIを操作し、RDSへ直接接続しない。Secretの取得はタスク起動時に行い、HTTPリクエストのたびには取得しない。
