#!/usr/bin/env bash
set -euo pipefail

# AWS_PROFILE / AWS_REGION are inherited from the operator's shell.
: "${NAT_INSTANCE_ID:?Set NAT_INSTANCE_ID to the NAT EC2 instance ID}"
if ! command -v session-manager-plugin >/dev/null 2>&1; then
  echo 'Install the Session Manager plugin before starting the tunnel' >&2
  exit 1
fi
local_port=${DB_LOCAL_PORT:-15432}
if ! [[ "$local_port" =~ ^[0-9]{1,5}$ ]] || (( 10#$local_port < 1 || 10#$local_port > 65535 )); then
  echo 'DB_LOCAL_PORT must be between 1 and 65535' >&2
  exit 1
fi

rds_host=$(aws rds describe-db-instances --db-instance-identifier "${RDS_INSTANCE_ID:-app-db}" \
  --query 'DBInstances[0].Endpoint.Address' --output text)
if [[ -z "$rds_host" || "$rds_host" == None ]]; then
  echo 'Could not resolve RDS endpoint' >&2
  exit 1
fi

echo "Forwarding localhost:$local_port to RDS via SSM. Leave this terminal open; Ctrl+C to stop."
exec aws ssm start-session \
  --target "$NAT_INSTANCE_ID" \
  --document-name AWS-StartPortForwardingSessionToRemoteHost \
  --parameters "host=$rds_host,portNumber=5432,localPortNumber=$local_port"
