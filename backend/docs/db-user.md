# RDSのアプリ用DBユーザー

## 目的と実装範囲

`cmd/db-user`は、管理者dbadminを使ってPostgreSQL内に`app_user`を作り、接続URLを既存のSecrets Manager `backend/database-url`へ登録する管理者向けGoコマンド。アプリのログインユーザーとは別であり、通常のHTTPサーバー起動時には実行しない。

現在はコマンドと単体テストの実装まで。実RDSへのユーザー作成・Secret値登録はまだ実行していない。ローカルからは[NATインスタンス経由のSSMポート転送](../../infrastructure/docs/db-tunnel.md)を使用する。利用前にSSM用IAM・SG設定のapplyとセッション開始が必要。

## 処理

1. AWS CLI経由でRDSのホスト名と管理者シークレットARNを取得する。
2. Secrets Managerからdbadminの認証情報をメモリ上へ読み取り、RDSへverify-fullで接続する。
3. 初回は暗号学的乱数からパスワードを生成してapp_userを作成する。
4. appデータベースへのCONNECT、publicスキーマへのUSAGE、存在するusers・posts・followsへのSELECT・INSERT・UPDATE・DELETEを付与する。
5. アプリ用接続URLをSecrets Managerへ登録する。パスワードを引数・一時ファイル・通常ログ・Terraform Stateへ出さない。

管理者権限、CREATE DATABASE、CREATE ROLE、テーブル所有権、schema_migrationsへの権限は付与しない。PostgreSQLのPUBLIC経由のデフォルト権限は変更しないため、既存DBの全面的な権限監査を行うコマンドではない。現在はUUIDを利用するためシーケンス権限は不要。

マイグレーション前はユーザーだけを作成できる。テーブル作成後に必ず再実行して権限を付ける。将来のテーブルへ自動付与はせず、必要な対象をコードに追加する。既存ユーザーのパスワードは変更しない。保存済みのURLが別DB・別ユーザーを指す場合や、ユーザーだけ存在してSecretがない場合は停止する。

DBとSecrets Managerは同一トランザクションにはできないため、DBの変更をコミットする前にSecretへ保存する。保存が失敗した場合はDB変更をロールバックする。保存後のDBコミットが失敗した場合は再実行で保存済みパスワードを再利用する。他の運用によるパスワード変更・同時Secret更新には対応せず、値が不整合の場合は手動調査する。実行中に他の作業で認証情報を変更しないこと。

## 実行条件

- Go、AWS CLI、curl、make、Bashがある環境。証明書の取得先へのHTTPS接続も必要。
- RDSの5432番へ到達できるVPC接続、または上記SSMポート転送。実行のためにDBを公開したりSGを全開放しない。
- 実行者のIAM権限：rds:DescribeDBInstances、管理者SecretへのGetSecretValue、アプリ用SecretへのDescribeSecret・GetSecretValue・PutSecretValue。
- 通常のECSタスクロールにはこれらの管理権限を追加しない。専用の管理環境を用意する。

接続環境を用意した後、リポジトリルートで実行する。Makefileが東京リージョンのRDS用CAを[AWS公式配布元](https://truststore.pki.rds.amazonaws.com/ap-northeast-1/ap-northeast-1-bundle.pem)から一時ファイルへ取得し、そのパスをGoコマンドへ渡す。手動でのダウンロード・パス指定は不要。取得に失敗した場合はGoコマンドを実行せず停止し、正常終了・エラー終了時ともに一時ファイルを削除する。

`````bash
AWS_PROFILE=x-clone-terraform-stg AWS_REGION=ap-northeast-1 \
  make -C backend db-user
`````

AWSプロフィールを持たないVPC内の専用管理環境では、適切なIAMロールの認証情報を使いAWS_PROFILEは指定しない。現行のscratch版バックエンドイメージにはGoコマンド・AWS CLI・この管理ツールは含めていないため、そこでこのmakeコマンドを実行することはできない。

クライアント側のエラーは認証情報が漏れない固定メッセージに置き換える。DB側の監査・SQLログへのアクセス権も別途管理する。実行後はマイグレーションと権限の再付与、必要な初期データ投入を行い、ECS起動・機能確認へ進む。

## テスト

`````bash
cd backend
go test ./cmd/db-user
go test ./...
go vet ./...
`````

URLエンコード、TLS設定、乱数パスワード、既存値の再利用、対象不一致の拒否、ユーザー作成・対象テーブルへの権限付与、テーブル未作成時の処理、エラー詳細の非表示をテストする。AWS CLIとの連携・VPC経由の実DB接続は未検証。
