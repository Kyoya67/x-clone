# RDS PostgreSQL

## 構成と選択理由

`modules/rds`を`stg/aws.tf`から呼び出す。ユーザーの動作確認では、RDSへの接続とマイグレーションの適用（version=4、dirty=false）が完了している。

| 項目 | stgの設定 | 理由・注意点 |
| --- | --- | --- |
| エンジン | PostgreSQL 16.15 | ローカルのPostgreSQL 16とメジャーを合わせる |
| インスタンス | db.t4g.micro / Single-AZ | stgのコストを抑える。AZ障害時の自動フェイルオーバーはない |
| ストレージ | 暗号化gp3、20 GiB、最大100 GiBへ自動拡張 | データ保護と容量不足への備え。拡張分は課金され、縮小できない |
| ネットワーク | 既存のDB専用privateサブネット2つ、非公開 | DBサブネットグループは2 AZに跨がるが、Single-AZなのでDBが2台作られるわけではない |
| SG | backend-sgと管理用NATのSGから5432を許可 | NAT経由はSSMポート転送用。インターネットからの直接接続は許可しない |
| TLS | rds.force_ssl=1 / rds-ca-rsa2048-g1 | 平文接続をDB側で拒否する。クライアント側のverify-fullも必要 |
| 認証 | dbadmin、RDS管理のSecrets Managerシークレット | パスワードをコードやTerraform Stateに渡さない |
| バックアップ | 7日保持、削除保護あり | 誤削除防止。削除時は保護解除と最終スナップショットが必要 |

RDSとSecrets Managerはapply後に料金が発生する。バックアップ時間はUTC 18:00–19:00、メンテナンスはUTC日曜19:00–20:00（日本時間では月曜04:00–05:00）。マイナーバージョンの自動更新を有効にしているため、今後の更新後にTerraformの指定バージョンも確認する。

## 接続の関係

`````text
ECSタスク（API用・マイグレーション用のSGをそれぞれ付与）
  └─ TCP 5432 / TLS → RDS（db-sg、DB専用privateサブネット）

backend-sgのTCP 443送信
  └─ NAT → ECR・CloudWatch Logs・Secrets Manager
`````

443の許可はECSのイメージ取得などに必要な通信のためで、Goアプリに外部API呼び出しを追加したわけではない。ECS側が行う通信とGoのCA証明書設定は別である。backend-sgの受信ルールは、ALB構築時に追加する。DB側には送信ルールを追加していないが、SGはステートフルなので許可した接続への応答は可能。

## 確認と適用

リポジトリルートから実行する。

`````bash
AWS_PROFILE=x-clone-terraform-stg terraform -chdir=infrastructure/stg init
AWS_PROFILE=x-clone-terraform-stg terraform -chdir=infrastructure/stg plan
# 差分と費用を確認した後に実行する
AWS_PROFILE=x-clone-terraform-stg terraform -chdir=infrastructure/stg apply
`````

モジュールのoutputは、今後ECSへ渡すRDSホスト名と管理者シークレットARNのみ。ルートのoutputs.tfは追加していない。シークレット値のdata sourceは使用せず、Terraformがパスワードを読み取らない構成とする。RDS管理シークレットは自動ローテーションされるため、固定値をアプリへコピーして運用しない。

## ECS起動前に残っている作業

DBユーザーとマイグレーションの方針・現在の実装は[こちら](../../backend/docs/db-operation-flow.md)を参照。

- アプリ用認証情報のSecrets Manager管理・ローテーション方針とECSへの注入を実装する。ECSが起動時に注入する値は、シークレット更新だけでは既存タスクに反映されないため、更新時の再起動も設計する。
- `DATABASE_SSL_MODE=verify-full`と、RDSのホスト名・`sslrootcert=/app/certs/rds-ca-bundle.pem`を含む接続URLで動作確認する。[コンテナの証明書設定](../../backend/docs/container.md)を参照。
- prdではMulti-AZ、インスタンスサイズ、監視・バックアップ要件を再検討する。同一アカウント・リージョンで併設する場合は識別子app-dbを環境ごとに区別する。

削除時の最終スナップショット名はapp-db-final。同名のスナップショットが既にある場合は、削除実行前に重複しない名前へ変更する。

## 参考資料

- [AWS：RDS PostgreSQLのTLS設定](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/PostgreSQL.Concepts.General.SSL.html)
- [Terraform：RDS管理のマスターパスワード](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/db_instance#manage_master_user_password)
