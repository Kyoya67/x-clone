# Infrastructure

AWS上にstg環境のアプリケーション基盤を構築するTerraformコードです。

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

## 環境構築

Terraformの実行は`infrastructure/stg`ディレクトリで行います。

`````bash
cd infrastructure/stg
`````

`.env`を作成し、dbadmin用のパスワードを設定します。

`````bash
TF_VAR_dbadmin_password=任意のRDS管理者パスワード
TF_VAR_dbadmin_password_version=1
`````

`.env`はTerraform実行時だけ読み込まれます。RDSのdbadminパスワードと、Secrets Managerの`db/dbadmin`へ同じ値を渡します。

Terraform plan：

`````bash
make plan
`````

Terraform apply：

`````bash
make apply
`````

## ECRへのイメージpush

API用イメージとマイグレーション用イメージは、`backend`ディレクトリからpushします。

`````bash
cd backend
make api-ecr-push
make migration-ecr-push
`````

push後、`infrastructure/stg/aws.tf`のイメージタグを更新してTerraformを再実行します。

`````bash
cd infrastructure/stg
make apply
`````

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

AWSリソース構成、セキュリティグループ、IAM、DBユーザー、Secret管理方針は[DESIGN.md](./DESIGN.md)を参照してください。
