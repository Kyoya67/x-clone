#!/usr/bin/env bash
# コマンド失敗・未定義変数・パイプ途中の失敗で停止する。
set -euo pipefail

# 1. 転送を中継するrunning状態のNATインスタンスをタグから取得する。
# AWS_PROFILE / AWS_REGIONは、このスクリプトを実行したシェルから引き継ぐ。
nat_instance_id=$(aws ec2 describe-instances \
  --filters "Name=tag:Name,Values=nat-instance" "Name=instance-state-name,Values=running" \
  --query "Reservations[0].Instances[0].InstanceId" \
  --output text)

# 接続先を取得できなければ、転送を開始せず停止する。
if [[ -z "$nat_instance_id" || "$nat_instance_id" == None ]]; then
  echo 'Could not resolve running nat-instance' >&2
  exit 1
fi

# 2. ローカルPCにSession Managerプラグインがあるか確認する。
if ! command -v session-manager-plugin >/dev/null 2>&1; then
  echo 'Install the Session Manager plugin before starting the tunnel' >&2
  exit 1
fi

# 3. ローカル側の待受ポートを設定する。未指定なら15432を使用する。
local_port=${DB_LOCAL_PORT:-15432}

# 1〜65535の整数であることを確認する。10#は数値を10進数として扱う指定。
if ! [[ "$local_port" =~ ^[0-9]{1,5}$ ]] || (( 10#$local_port < 1 || 10#$local_port > 65535 )); then
  echo 'DB_LOCAL_PORT must be between 1 and 65535' >&2
  exit 1
fi

# 4. AWS CLIで転送先RDSのホスト名を取得する。インスタンス名の既定値はapp-db。
rds_host=$(aws rds describe-db-instances --db-instance-identifier "${RDS_INSTANCE_ID:-app-db}" \
  --query 'DBInstances[0].Endpoint.Address' --output text)

# 接続先を取得できなければ、転送を開始せず停止する。
if [[ -z "$rds_host" || "$rds_host" == None ]]; then
  echo 'Could not resolve RDS endpoint' >&2
  exit 1
fi

# 5. SSMのポート転送を開始する。
# ローカルPCの待受ポート → SSM → NATインスタンス → RDSの5432番、という経路を作る。
# DBへのログインやSQL実行は行わない。利用中はターミナルを開いたままにする。
echo "Forwarding localhost:$local_port to RDS via SSM. Leave this terminal open; Ctrl+C to stop."

# execでシェルをAWS CLIに置き換え、セッション終了まで待つ。終了操作はCtrl+C。
exec aws ssm start-session \
  --target "$nat_instance_id" \
  --document-name AWS-StartPortForwardingSessionToRemoteHost \
  --parameters "host=$rds_host,portNumber=5432,localPortNumber=$local_port"
