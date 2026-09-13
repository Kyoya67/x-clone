# AWSインフラ構成

## 1. 環境構築手順

stg環境のTerraform実行、DBユーザー登録、マイグレーション、API起動確認までの流れ。

| 順序 | 作業 | 実行場所 | コマンド |
| ---- | ---- | -------- | -------- |
| 1 | dbadmin用の環境変数を用意 | infrastructure/stg | .envを作成 |
| 2 | AWSリソースを作成・更新 | infrastructure/stg | make apply |
| 3 | API・マイグレーション用イメージをECRへpush | backend | make api-ecr-push<br>make migration-ecr-push |
| 4 | Terraform側のイメージタグを更新して反映 | infrastructure/stg | make apply |
| 5 | SSMポート転送を開始 | backend | make db-tunnel |
| 6 | DBユーザーを登録 | backend | make db-user |
| 7 | RDSマイグレーションを単発実行 | backend | AWS_PROFILE=x-clone-terraform-stg AWS_REGION=ap-northeast-1 make migrate-rds-up |
| 8 | APIのECSサービスを起動・更新 | infrastructure/stg | make apply |
| 9 | APIタスクの起動・ヘルスチェックを確認 | AWSコンソール / AWS CLI | ECSサービス、タスク、CloudWatch Logsを確認 |

### .env

infrastructure/stg/.envで、Terraform実行時だけ使う値を管理する。

`````bash
TF_VAR_dbadmin_password=任意のRDS管理者パスワード
TF_VAR_dbadmin_password_version=1
`````

この値はRDSのdbadminパスワードと、Secrets Managerのdb/dbadminへ同時に渡す。Secretの中身をTerraform data sourceで読み込むとTerraform Stateに残る可能性があるため、Secretから読み戻してRDSへ渡す構成にはしない。

### DBユーザー登録

DBユーザー登録は、SSMポート転送を開いた状態で実行する。

`````bash
cd backend
make db-tunnel
`````

別ターミナルで実行する。

`````bash
cd backend
make db-user
`````

db-userはdbadminとしてRDSへ接続し、app_userとmigration_userを作成・更新する。

### RDSマイグレーション

RDSマイグレーションはECSの単発タスクdb-migratorで実行する。SSMポート転送は使わない。

`````bash
cd backend
AWS_PROFILE=x-clone-terraform-stg AWS_REGION=ap-northeast-1 make migrate-rds-up
`````

状態確認だけ行う場合。

`````bash
cd backend
AWS_PROFILE=x-clone-terraform-stg AWS_REGION=ap-northeast-1 make migrate-rds-status
`````

## 2. VPC・サブネットとリソース配置

VPC（10.0.0.0/16）の中に、次の4つのサブネットがある。

| VPC内のサブネット | CIDR          | 現在配置されているリソース                        |
| ----------------- | ------------- | ------------------------------------------------- |
| ├ public-1a       | 10.0.0.0/18   | NATインスタンス（EC2）。外向き通信とSSM接続の中継 |
| ├ public-1c       | 10.0.64.0/18  | なし                                              |
| ├ private-1a      | 10.0.128.0/18 | APIのECSタスク、RDS（app-db）                     |
| └ private-1c      | 10.0.192.0/18 | なし                                              |

## 3. リソースとセキュリティーグループの対応・許可する通信

| AWSリソース       | セキュリティグループ名 | 用途                      | インバウンドルール                                                                                              | アウトバウンドルール                  |
| ----------------- | ---------------------- | ------------------------- | --------------------------------------------------------------------------------------------------------------- | ------------------------------------- |
| EC2：nat-instance | nat-instance           | 外向き通信・SSM接続の中継 | 10.0.128.0/18・10.0.192.0/18から全プロトコル許可<br>EC2 Instance ConnectのAWS管理プレフィックスリストからTCP 22 | 0.0.0.0/0へ全プロトコル許可           |
| ECS：api          | api                    | APIサーバー               | nat-instance SGからTCP 8080（手動API確認）                                                                      | db SGへTCP 5432<br>0.0.0.0/0へTCP 443 |
| ECS：db-migrator  | db-migrator            | DBマイグレーション        | なし                                                                                                            | db SGへTCP 5432<br>0.0.0.0/0へTCP 443 |
| RDS：app-db       | db                     | PostgreSQL                | api・db-migrator・nat-instance SGからTCP 5432                                                                   | なし（許可済み接続への応答は可能）    |

## 4. IAMユーザー・ロール・ポリシー

| 利用者・AWSリソース | IAMユーザー／ロール名                    | 用途                                     | アタッチする権限ポリシー                     | 許可する操作・対象                                                                                   |
| ------------------- | ---------------------------------------- | ---------------------------------------- | -------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| 開発者PC            | x-clone-terraform-stg（IAMユーザー）     | Terraform・AWS CLI・DB管理コマンドの実行 | AdministratorAccess（AWS管理）               | 全AWS操作・全リソース（ポリシー上の許可。SCP等の制約は未確認）                                       |
| ECS：api            | api-task-execution（実行ロール）         | ECS基盤によるコンテナ起動・ログ送信      | api-task-execution（カスタマー管理）         | ECR apiのイメージ取得<br>/ecs/backendへのログ送信<br>db/app_userのSecret取得                         |
| ECS：api            | api-task（タスクロール）                 | APIプログラムのAWS操作用                 | なし                                         | AWS操作権限なし。DBの読み書きはapp_userのSQL権限                                                     |
| ECS：db-migrator    | db-migrator-task-execution（実行ロール） | ECS基盤によるコンテナ起動・ログ送信      | db-migrator-task-execution（カスタマー管理） | ECR db-migratorのイメージ取得<br>/ecs/backend-migrationへのログ送信<br>db/migration_userのSecret取得 |
| ECS：db-migrator    | db-migrator-task（タスクロール）         | マイグレーションプログラムのAWS操作用    | なし                                         | AWS操作権限なし。DB変更はmigration_userのSQL権限                                                     |
| EC2：nat-instance   | nat-ssm（ロール）                        | SSM Agentの管理・通信                    | AmazonSSMManagedInstanceCore（AWS管理）      | SSMへの情報登録・管理用通信。DB操作・Secret取得の権限なし                                            |

## 5. DBユーザー・Secret・実行場所の関係

詳細な初期設定手順、Secret管理方針、ECSタスクとの関係は [DESIGN.md](./DESIGN.md) にまとめる。

| DBユーザー | Secret | 使う場所 | 用途 |
| ---------- | ------ | -------- | ---- |
| dbadmin | db/dbadmin | 開発者PCで実行するDBユーザー初期設定コマンド | 初期設定・ユーザー管理 |
| migration_user | db/migration_user | ECSタスク：db-migrator | テーブル作成・変更、migration.schema_migrationsの管理 |
| app_user | db/app_user | ECSサービスで常時起動するAPIタスク | アプリデータの読み書き |

## 6. 残りの対応

- 作業用IAMユーザーはAdministratorAccess。用途別の最小権限化は未実施。
- NATは単一障害点。APIも通常1タスク、RDSもSingle-AZであり、全体として冗長構成ではない。
- stg/aws.tfに残るdbadmin_references_ready=falseと「applyを禁止する」コメントは、現在のRDSリソースで停止条件として使われていない。
