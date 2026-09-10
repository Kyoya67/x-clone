# ローカルからRDSへのSSMポートフォワーディング

既存NATインスタンスをSSMの転送先として兼用し、privateなRDSへ接続する。DB管理接続ではSSH鍵・ローカルIPの許可・22番への接続を使用しない。ブラウザ版EC2 Instance Connect用の既存SSHルールだけは維持する。

`````text
ローカルのGoコマンド / TablePlus
  → localhost:15432
  → Session Manager
  → NAT上のSSM Agent
  → RDS:5432
`````

GoはTCP接続先だけをローカルへ変更し、RDSのホスト名・CAによるverify-full検証は維持する。Secretへ保存するURLもRDSのホスト名のまま。TablePlusでもTLSのCA・ホスト名検証を別途設定する必要がある。

## 1. Terraformを適用する

`````bash
export AWS_PROFILE=x-clone-terraform-stg
export AWS_REGION=ap-northeast-1
terraform -chdir=infrastructure/stg plan
# 差分を確認してから実行
terraform -chdir=infrastructure/stg apply
`````

NATにSSM用IAMロール・インスタンスプロファイルを付与する。ロールにはAWS管理ポリシーAmazonSSMManagedInstanceCoreをattachmentで紐付け、DB認証情報へのアクセス権は与えない。ローカルSSH用ルールとoperator_ssh_cidrは廃止。以前terraform.tfvarsへ記載した場合はその項目も削除する。NATのSGからRDS:5432への許可は維持する。

AL2023のSSM Agentを利用する。既存EC2のuser_dataは変更せず、NATの停止・再起動を伴う変更を避ける。適用後にSSM登録を確認する。

`````bash
aws ssm describe-instance-information \
  --filters Key=InstanceIds,Values=i-03653ea01d0f54ffa \
  --query 'InstanceInformationList[].{Status:PingStatus,AgentVersion:AgentVersion}'
`````

Onlineになること、リモートホスト転送に必要なAgent 3.1.1374.0以上であることを確認する。登録されない場合は既存のブラウザ接続から確認し、停止していれば起動する。

`````bash
sudo systemctl status amazon-ssm-agent
# 停止・無効の場合
sudo systemctl enable --now amazon-ssm-agent
`````

AgentからSSM/ssmmessagesのHTTPSエンドポイントへの到達性が必要。現在のNATはpublicサブネットにあり、公開IP・IGW・既存の送信許可を利用する。RDSへの接続許可と、SSMへ接続するIAM権限は別。

## 2. 実行者とPCの準備

ローカルにはAWS CLIとSession Managerプラグインが必要。このMacではプラグインの存在を確認済み。[インストール手順](https://docs.aws.amazon.com/ja_jp/systems-manager/latest/userguide/install-plugin-macos-overview.html)

実行者には次の権限が必要。今回のTerraformは実行者のIAMユーザーの権限を変更しない。

- rds:DescribeDBInstances（RDSホスト名取得）
- ssm:StartSession（対象NATのEC2 ARNとAWS-StartPortForwardingSessionToRemoteHostドキュメントARN）
- ssm:TerminateSession / ssm:ResumeSession（自身のセッション）
- ssm:DescribeInstanceInformation（Agent確認）
- DBユーザー作成時は別途、管理者Secret取得・アプリSecret登録の権限。[db-userの実行条件](../../backend/docs/db-user.md)を参照。

## 3. ターミナルAで転送を開始する

リポジトリルートで実行する。

`````bash
export AWS_PROFILE=x-clone-terraform-stg
export AWS_REGION=ap-northeast-1
export NAT_INSTANCE_ID=i-03653ea01d0f54ffa
make -C backend db-tunnel
`````

EC2再作成時はNAT_INSTANCE_IDを更新する。DB_LOCAL_PORTで待受ポートを変更できる（既定15432）。転送開始後はターミナルを開いたままにし、終了時はCtrl+C。転送セッションが開いただけではDB認証成功とは限らない。

## 4. ターミナルBでDBユーザーを作成する

この操作はDBユーザー・権限とSecret値を変更する。

`````bash
AWS_PROFILE=x-clone-terraform-stg AWS_REGION=ap-northeast-1 \
  make -C backend db-user DB_TUNNEL=127.0.0.1:15432
`````

CAはMakefileが自動取得する。マイグレーション後にも再実行してテーブル権限を付与する。GoやDB認証情報をNATへ配置しない。

## 運用と検証状況

SSMセッション開始はIAMで制御する。ポート転送内のSQLはSession Managerのセッション内容ログには記録されないため、必要ならDB側で監査する。標準のリモートホスト転送ドキュメントは転送先をパラメータ指定できるため、接続権限は信頼できる管理者に限定し、転送先のSGも制限する。NATとの兼用はstgのコスト優先であり、将来は専用管理ホストへの分離を検討する。

SSM用設定・スクリプトは実装済み。apply・実SSM接続・実RDSのユーザー作成は未実施。

参考：[AWSのリモートホスト転送](https://docs.aws.amazon.com/ja_jp/systems-manager/latest/userguide/session-manager-working-with-sessions-start.html#sessions-remote-port-forwarding)
