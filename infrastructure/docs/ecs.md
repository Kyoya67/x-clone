# ECSタスク定義

## 今回の実装範囲

クラスターx-cloneのサービスapiで、FargateのAPIタスクを通常1つ維持する。ALB・外部公開は次のPRとし、ECSのコンテナhealthCheckでHTTP応答を自動確認する。

更新中は旧タスクを残して新タスクを起動するため、一時的に最大2タスクとなる。起動失敗時はデプロイサーキットブレーカーで停止・ロールバックする（初回は戻せる成功済みデプロイがない）。Terraformはサービスの安定化を待つ。[ECSサービス設定](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecs_service)

| 項目 | 設定 |
| --- | --- |
| ECSクラスター | x-clone |
| タスクファミリー / コンテナ名 | api / api |
| 起動方式 | Fargate、awsvpc、Linux/X86_64 |
| CPU / メモリ | 0.25 vCPU / 512 MiB。stgの初期値で、負荷確認後に調整 |
| イメージ | ECR api。更新時はstg/aws.tfのタグを変更 |
| ポート | 8080 |
| 実行ユーザー | 65532:65532（非root） |
| ファイルシステム | 読み取り専用 |
| ログ | /ecs/backend、14日保持 |
| DB接続 | DB_HOST・DB_PORTは環境変数、DB_USER・DB_PASSWORDはSecretのJSONキーから注入。TLSはverify-full |

リソース別に`modules/ecs_task_definition`・`modules/iam`・`modules/cloudwatch_logs`・`modules/secrets_manager`へ分割し、`stg/aws.tf`で組み立てる。

## ロールと認証情報

実行権限は`aws_iam_policy`で独立したカスタマー管理ポリシーとして作成し、`aws_iam_role_policy_attachment`で実行ロールへ紐付ける。他のロールでも`execution_policy_arn`を使って共有できる。ただし対象のECR・ロググループ・シークレットに権限を限定しているため、同じリソースへのアクセスが必要なロールにのみ付与する。共有ポリシーの変更は紐付いた全ロールに影響する。

- 実行ロール：ECS基盤がECRイメージ取得、ログ送信、`db/app_user`の値の取得に使用する。ECR認証トークン以外は対象リソースのARNに権限を限定する。
- タスクロール：Goアプリが使用する。現在AWS APIを呼ばないため権限ポリシーを付けない。
- RDS管理者シークレット：今回のタスクからは参照も取得もできない。通常のアプリでdbadminは使用しない。

app_user用Secretの保存先はTerraformで作成し、管理コマンドがusername・passwordのJSONを保存する。接続先はSecretに含めない。

実際の認証情報をGit・タスク定義のenvironment・ドキュメントに記載しない。値未登録ではECSがSecretを取得できず起動に失敗する。値を変更しても既存タスクの環境変数は更新されないため、タスクの再起動が必要。

## 起動とhealth確認

DBユーザーとマイグレーションの方針・現在の実装は[こちら](../../backend/docs/db-operation-flow.md)を参照。

1. healthcheck追加をコミットし、APIイメージを再build・pushする。TerraformのAPIタグをそのタグへ更新する。古いイメージには確認コマンドがないため、先にapplyしない。db/app_userの認証情報も必要。
2. Terraformをinit・plan・applyし、サービスapiのタスクがRUNNING、コンテナのヘルスステータスがHEALTHYになることを確認する。
3. ECSコンソールでタスクのプライベートIPを確認する。
4. 手動でも確認する場合は、NATインスタンスへSSMで接続し、次を実行する（IPは実際のタスクに置き換える）。

`````bash
curl --fail --show-error --max-time 10 http://<タスクのプライベートIP>:8080/health
`````

200と`{"status":"ok"}`を確認する。Goは起動時のdb.Ping成功後にHTTPサーバーを開始するため、起動時のRDS接続成功も確認できる。ただしhealth自身はDBへ問い合わせず、その時点の接続や読み書き権限は保証しない。実環境でのサービス起動・HTTP確認は未実施。

サービスはapp-privateサブネット・backend-sgを使用し、パブリックIPを付けない。8080の受信はNATホストのSGからのみ許可し、ECR等への通信はNATを経由する。

scratchにはシェルやcurlがないため、Go製の/app/healthcheckをCMDで直接実行する。コンテナ内の127.0.0.1:8080/healthが200なら終了コード0、それ以外・通信失敗は1を返す。30秒間隔、タイムアウト5秒（HTTP側3秒）、連続失敗3回、起動猶予30秒。UNHEALTHYになったタスクはECSサービスが置き換える。[AWSのコンテナヘルスチェック](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_HealthCheck.html)

ECRのライフサイクルは実行中タスクの参照を考慮しないので、指定タグが削除されると新しいタスクが起動できなくなる点にも注意する。

## 検証

リポジトリルートから実行する。applyは差分を確認してから行う。

`````bash
AWS_PROFILE=x-clone-terraform-stg terraform -chdir=infrastructure/stg init
AWS_PROFILE=x-clone-terraform-stg terraform -chdir=infrastructure/stg validate
AWS_PROFILE=x-clone-terraform-stg terraform -chdir=infrastructure/stg plan
`````

参考：[AWSのタスク実行ロール](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_execution_IAM_role.html)、[Secrets Managerからの注入](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/specifying-sensitive-data-tutorial.html)。
