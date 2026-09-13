# Infrastructure

AWS上にアプリケーション基盤を構築するTerraformコードです。

## ディレクトリ構成

`````text
infrastructure/
├── README.md              # インフラの環境構築・運用手順
├── ARCHITECTURE.md        # AWSリソース構成、SG、IAM、DBユーザー設計
├── modules/               # 環境間で再利用するTerraform module
│   ├── vpc/
│   ├── subnet/
│   ├── route_table/
│   ├── security_group/
│   ├── ec2/
│   ├── ecr/
│   ├── ecs/
│   ├── ecs_service/
│   ├── ecs_task_definition/
│   ├── rds/
│   ├── secrets_manager/
│   └── cloudwatch_logs/
├── stg/                   # stg環境のTerraform root module
│   ├── aws.tf
│   ├── backend.tf
│   ├── variable.tf
│   ├── version.tf
│   ├── Makefile
│   └── .env.example
└── prd/                   # prd環境のTerraform root module
    ├── aws.tf
    ├── backend.tf
    ├── variable.tf
    ├── version.tf
    └── import.tf
`````

環境ごとにTerraform root moduleを分ける。`stg/`と`prd/`で`modules/`を再利用する。

## 使用技術

- Terraform
- AWS
- Amazon ECS
- Amazon RDS for PostgreSQL
- Amazon ECR
- AWS Secrets Manager
- AWS Systems Manager Session Manager

## 必要な環境

- Terraform 1.14.1以上
- AWS CLI
- Docker
- jq

バージョン確認：

`````bash
terraform version
aws --version
docker --version
jq --version
`````

## 全体の構築・反映順序

| 順序 | 作業 | 主な対象 |
| ---- | ---- | -------- |
| 1 | TerraformでAWSリソースを作成する | VPC、Subnet、SG、IAM、ECR、ECS、RDS、Secrets Manager |
| 2 | DockerイメージをECRへpushする | API、db-migrator |
| 3 | TerraformでECSタスク定義・サービスへ反映する | ECSタスク定義、ECSサービス |
| 4 | DBユーザーを登録する | dbadmin、app_user、migration_user、Secrets Manager |
| 5 | RDSマイグレーションを実行する | db-migrator ECS単発タスク、RDS |
| 6 | API起動を確認する | ECSサービス、APIタスク、CloudWatch Logs |

## TerraformによるAWSリソース構築

Terraformの実行は`infrastructure/stg`ディレクトリで行います。

`````bash
cd infrastructure/stg
`````

`.env`を作成し、dbadmin用のパスワードを設定します。

`````bash
cp .env.example .env
`````

`````bash
TF_VAR_dbadmin_password=任意のRDS管理者パスワード
TF_VAR_dbadmin_password_version=1
`````

`.env`はTerraform実行時だけ読み込まれます。RDSのdbadminパスワードと、Secrets Managerの`db/dbadmin`へ同じ値を渡します。

初回のみTerraformを初期化します。

`````bash
terraform init
`````

Terraform plan：

`````bash
make plan
`````

Terraform apply：

`````bash
make apply
`````

## ECRへのDockerイメージpush

API用イメージとマイグレーション用イメージは、`backend`ディレクトリからpushします。

`````bash
cd backend
make api-ecr-push
make migration-ecr-push
`````

## ECSへのイメージ反映

初回構築時は、ECRへpush済みの実在するイメージタグを各環境の`aws.tf`へ反映し、ECSタスク定義とECSサービスを作成します。

`````bash
cd infrastructure/stg
make apply
`````

`````bash
cd infrastructure/prd
AWS_PROFILE=x-clone-terraform-prd make apply
`````

初回以降のイメージタグ更新はGitHub Actions CDが行います。Terraform側ではECSタスク定義の`container_definitions`を`ignore_changes`にしているため、CDが反映したcommit SHAのイメージタグを、後続のTerraform applyで古いタグへ戻さない構成です。

そのため、初回構築時だけは次の順序にします。

1. TerraformでECRなどの土台を作成する
2. APIとdb-migratorのDockerイメージをECRへpushする
3. push済みの実在タグを`aws.tf`のECSタスク定義imageへ入れる
4. Terraform applyでECSタスク定義・ECSサービスを作成する
5. 以降のイメージタグ更新はCDに任せる

APIはECSサービスで常時起動します。マイグレーションはECSサービスではなく、必要なときだけ単発タスクとして起動します。

## DBユーザー登録

DBユーザー登録では、開発者PCからRDSへ接続するためにSSMポート転送を使います。

1つ目のターミナルでSSMポート転送を開始します。

`````bash
cd backend
make db-tunnel
`````

2つ目のターミナルでDBユーザー登録を実行します。

`````bash
cd backend
make db-user
`````

`make db-user`は、dbadminでRDSへ接続し、`app_user`と`migration_user`を作成・更新します。

## RDSマイグレーション

RDSマイグレーションは、ECSの単発タスク`db-migrator`で実行します。SSMポート転送は使いません。

`````bash
cd backend
AWS_PROFILE=x-clone-terraform-stg AWS_REGION=ap-northeast-1 make migrate-rds-up
`````

状態確認だけ行う場合：

`````bash
cd backend
AWS_PROFILE=x-clone-terraform-stg AWS_REGION=ap-northeast-1 make migrate-rds-status
`````

## API起動確認

ECSサービスがAPIタスクを起動していることを確認します。

- ECSクラスター：`x-clone`
- ECSサービス：`api`
- ECSタスク定義：`api`

起動後、ECSタスクのヘルスチェックとCloudWatch Logsを確認します。

## 設計資料

AWSリソース構成、セキュリティグループ、IAM、DBユーザー、Secret管理方針は[ARCHITECTURE.md](./ARCHITECTURE.md)を参照してください。
