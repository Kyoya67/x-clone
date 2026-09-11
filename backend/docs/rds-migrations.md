# RDSマイグレーション（ECS単発タスク）

SSMとの違いと各Go・shファイルの関係は[DB初期設定の流れ](db-operation-flow.md)を参照。

## 方針

マイグレーションは初回からECS/Fargateの単発タスクで実行する。ローカルやCIからAWS CLIのecs run-taskで起動する。SSM・SSH・ECS Execは不要。DBユーザー作成とTablePlus接続には引き続きSSMを使う。

`````text
ローカルPC / 将来のGitHub Actions
  → ECS RunTask
  → app-privateサブネットのmigrationコンテナ
  → RDSへ直接TLS接続
  → SQL適用後に終了
  → 呼び出し元が終了コード0を確認
`````

既存migrations/001〜004とgolang-migrate v4.19.1を使用し、public.schema_migrationsで管理する。シード投入・app_user作成は行わない。通常のAPIサーバー起動時にもマイグレーションを実行しない。

## リソースと権限

| 対象 | 設定 |
| --- | --- |
| ECSクラスター | app。サービスはまだ作らず、単発タスクだけ起動 |
| タスク定義 | backend-migration。0.25 vCPU / 512 MiB / Linux amd64 |
| コンテナ | migration。scratch・非root・読み取り専用。SQLとRDS用CAを含む |
| ECR | backend-migration。API用とは別。タグはコミットハッシュ先頭6桁 |
| DBユーザー | 初期構築はdbadmin。通常APIのapp_userとは分離 |
| タスク実行ロール | migration-task-execution。専用ECR・ロググループ・RDS管理者Secretに限定 |
| タスクロール | migration-task。コンテナ本体はAWS APIを呼ばないため権限ポリシーなし |
| ネットワーク | app-private → RDS:5432。ECR・Secrets Manager・ログ送信はNAT経由HTTPS。公開IP・受信ルールなし |
| ログ | /ecs/backend-migration（14日保持） |

ECS基盤がRDS管理者Secretのusername/passwordを環境変数DB_USER/DB_PASSWORDとして注入する。TerraformにはSecretのARNだけを渡し、値は取得・保存しない。コンテナのGoコマンドは環境変数から接続情報を組み立て、RDSのホスト名とCAでverify-full検証する。AWS CLIやGoのAWS SDKはイメージ内に不要。Secret値は標準出力や引数へ出さない。ただしコンテナ内では環境変数としてアクセスできるため、タスク定義変更・実行権限は管理者に限定する。

DB管理者Secretを読める権限は専用ロールだけに追加し、通常バックエンドのロールは変更しない。将来は所有権・拡張機能の権限を整理した上で、専用のマイグレーションDBユーザーへ切り替える。

## 初回の準備

以下はリポジトリルートで実行する。まず変更をコミットし、イメージ内容とタグを一致させる。タグはビルドからタスク登録まで同じ値を使う。

`````bash
export AWS_PROFILE=x-clone-terraform-stg
export AWS_REGION=ap-northeast-1
export TF_VAR_migration_image_tag="$(git rev-parse HEAD | cut -c1-6)"

terraform -chdir=infrastructure/stg init
terraform -chdir=infrastructure/stg plan
# 差分を確認してから適用。まだマイグレーションは動かない。
terraform -chdir=infrastructure/stg apply

# 同じコミットから専用イメージをビルド・pushする。
make -C backend migration-ecr-push
`````

Terraformはクラスター・タスク定義などを作るだけで、イメージをpushせずタスクも起動しない。最初はECRが存在しないため、リソース作成→push→タスク実行の順になる。既存のAPI用ECRとイメージ指定は変えない。専用ECRも最新3件保持で、使用中イメージを判定しないため古いリビジョンの再実行時はイメージの存在を確認する。

## 状態確認・適用

実行環境にはAWS CLIとjqが必要。ローカルでは上記AWS_PROFILE/AWS_REGIONを設定する。将来のCIではOIDC等の一時認証情報を使う。

`````bash
# 単発ECSタスクで適用状況を確認（テーブルは変更しない）。
make -C backend migrate-rds-status

# 単発ECSタスクで未適用SQLをすべて適用する。
make -C backend migrate-rds-up

# 現在のSQLではCloudWatch Logsでversion=4 dirty=falseを確認する。
aws logs tail /ecs/backend-migration --since 10m
`````

スクリプトはProject=x-clone・Env=stg・Tier=app-privateのタグからサブネットを取得し、同じVPC内のmigration SGを使用する。複数VPCやSGが見つかった場合は停止する。DEPLOY_ENV、ECS_CLUSTER、ECS_MIGRATION_TASK_DEFINITIONで対象環境・クラスター・タスク定義リビジョンを指定できる。prd用の環境は未構築。

デフォルトのタスクはstatusのみ。upを実行する時はコンテナコマンドだけを上書きする。起動APIのfailures、停止状態、stopCode、migrationコンテナの終了コードを確認する。STOPPEDになっただけでは成功としない。待機のタイムアウトや中断ではECSタスクが動き続けている可能性があるので、表示されたTask ARNを調査してから再実行する。自動再起動・強制停止は行わない。

実行者にはec2:DescribeSubnets・DescribeSecurityGroups、ecs:DescribeTaskDefinition・RunTask・DescribeTasks、およびmigration-task-execution/migration-taskの2ロールへのiam:PassRoleが必要。PassRoleの渡し先はecs-tasks.amazonaws.comに限定する。これらは今回IAMユーザーへ自動付与しない。ログ参照には別途CloudWatch Logsの読み取り権限が必要。

## 適用後

マイグレーション成功後、[SSM接続](../../infrastructure/docs/db-tunnel.md)を開始して次を実行する。これはDBユーザーとアプリ用Secretを作成・更新する別の操作。

`````bash
make -C backend db-user DB_TUNNEL=127.0.0.1:15432
`````

app_userには作成済みのusers・posts・followsへの読み書き権限を付与する。必要な初期データ投入とアプリ起動確認は別途行う。

## 失敗時と今後のCI/CD

適用済みのSQLは再実行しない。dirty=trueや外部キー違反等で失敗した場合は実DBとログを調査する。down・forceはこのコマンドに追加していない。ドライバの詳細には機密情報が含まれる可能性があるため固定メッセージへ置き換える。SQLは1ファイル60秒、マイグレーションロック取得待ちは15秒を上限とする。

CI/CDでも同じタスクを起動し、成功時だけアプリのデプロイへ進む。並列デプロイを防止し、稼働中の旧アプリとも互換性があるDB変更を行う。CI/CDワークフロー自体は今回未実装。

## 検証状況

Goでは成功・適用済み・dirty拒否・エラー詳細の非表示・SQL読み込み・環境変数とTLS設定をテストする。起動スクリプトはAWS CLIをモックに置き換え、成功・起動失敗・終了コード異常/欠落・待機失敗をテストする（jqがない環境ではスキップ）。

既存SQLと実行処理は使い捨てPostgreSQL 16でversion=4 dirty=falseと再実行成功を確認済み。実AWSへのapply・ECR push・ECSタスク起動・RDS適用は未実施。

2026-09-10: Go全体のテスト・vet、起動スクリプトのモックテスト、専用イメージのlinux/amd64ビルド、ネットワーク無効・読み取り専用コンテナでの認証情報未設定時の終了コード1を確認。Terraform validate成功。planは13追加・1変更・0削除で、1変更は既存のRDSパラメータ適用方法の差分（SSM/ECS対応とは無関係）。

参考：[ECS単発タスク](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/standalone-tasks.html)、[Secretの環境変数注入](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/secrets-envvar-secrets-manager.html)
