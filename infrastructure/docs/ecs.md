# ECSタスク定義

## 今回の実装範囲

Fargate用タスク定義と、IAMロール・CloudWatch Logs・アプリ用Secrets Managerの保存先を追加する。ECSクラスター・サービス・ALBはまだ作成しない。タスク定義の登録だけではコンテナは起動しない。

| 項目 | 設定 |
| --- | --- |
| タスクファミリー / コンテナ名 | backend |
| 起動方式 | Fargate、awsvpc、Linux/X86_64 |
| CPU / メモリ | 0.25 vCPU / 512 MiB。stgの初期値で、負荷確認後に調整 |
| イメージ | ECR backend:78694d。更新時はstg/aws.tfのタグを変更 |
| ポート | 8080 |
| 実行ユーザー | 65532:65532（非root） |
| ファイルシステム | 読み取り専用 |
| ログ | /ecs/backend、14日保持 |
| DB接続 | DATABASE_SSL_MODE=verify-full、DATABASE_URLはSecretから起動時に注入 |

リソース別に`modules/ecs_task_definition`・`modules/iam`・`modules/cloudwatch_logs`・`modules/secrets_manager`へ分割し、`stg/aws.tf`で組み立てる。

## ロールと認証情報

実行権限は`aws_iam_policy`で独立したカスタマー管理ポリシーとして作成し、`aws_iam_role_policy_attachment`で実行ロールへ紐付ける。他のロールでも`execution_policy_arn`を使って共有できる。ただし対象のECR・ロググループ・シークレットに権限を限定しているため、同じリソースへのアクセスが必要なロールにのみ付与する。共有ポリシーの変更は紐付いた全ロールに影響する。

- 実行ロール：ECS基盤がECRイメージ取得、ログ送信、`backend/database-url`の値の取得に使用する。ECR認証トークン以外は対象リソースのARNに権限を限定する。
- タスクロール：Goアプリが使用する。現在AWS APIを呼ばないため権限ポリシーを付けない。
- RDS管理者シークレット：今回のタスクからは参照も取得もできない。通常のアプリでdbadminは使用しない。

Secrets Managerは保存先のみをTerraformで作成する。値はTerraformに渡さず、アプリ用DBユーザーを作成後にAWSコンソールなどからプレーンテキストとして次の形式で登録する（JSONではない）。

`````text
postgres://<アプリ用ユーザー>:<URLエンコード済みパスワード>@<RDSのホスト名>:5432/app?sslrootcert=/app/certs/rds-ca-bundle.pem
`````

実際の認証情報をGit・タスク定義のenvironment・ドキュメントに記載しない。値未登録ではECSがSecretを取得できず起動に失敗する。値を変更しても既存タスクの環境変数は更新されないため、タスクの再起動が必要。

## 起動前の残作業

1. アプリ用DBユーザー・権限・接続URLを用意する。
2. ECSクラスターとマイグレーション用の実行手段を追加し、テーブルを作成する。
3. サービスでapp-privateサブネットとbackend-sgを指定する。パブリックIPは付けず、AWSサービスへはNAT経由で到達させる。
4. ALB・ターゲットグループ・backend-sgの受信ルールを追加し、サービスを起動する。ALBのヘルスチェックはGET /healthを使用する予定。
5. ログ、TLS接続、投稿・フォロー・タイムラインを実環境で確認する。

scratchにはシェルやcurlがないため、それらを使うコンテナhealthCheckは定義していない。ECSサービス側のALBヘルスチェックは今後設定する。ECRのライフサイクルは実行中タスクの参照を考慮しないので、指定タグが削除されると新しいタスクが起動できなくなる点にも注意する。

## 検証

リポジトリルートから実行する。applyは差分を確認してから行う。

`````bash
AWS_PROFILE=x-clone-terraform-stg terraform -chdir=infrastructure/stg init
AWS_PROFILE=x-clone-terraform-stg terraform -chdir=infrastructure/stg validate
AWS_PROFILE=x-clone-terraform-stg terraform -chdir=infrastructure/stg plan
`````

参考：[AWSのタスク実行ロール](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_execution_IAM_role.html)、[Secrets Managerからの注入](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/specifying-sensitive-data-tutorial.html)。
