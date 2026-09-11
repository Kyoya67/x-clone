#!/usr/bin/env bash
# コマンド失敗・未定義変数・パイプ途中の失敗で停止する。
set -euo pipefail

# 1. 実行内容を確認する。引数なしは状態確認、upはマイグレーション適用。
action=${1:-status}
case "$action" in status|up) ;; *) echo 'Action must be status or up' >&2; exit 1 ;; esac

# AWSのJSONレスポンスを処理するため、jqが必要。
command -v jq >/dev/null || { echo 'jq is required' >&2; exit 1; }

# 2. 対象環境を設定する。環境変数がなければ右側の値を使用する。
# AWS_PROFILE/AWS_REGIONは実行元から引き継ぐ。CIでは引き受けたIAMロールも利用可能。
cluster=${ECS_CLUSTER:-x-clone}
definition=${ECS_MIGRATION_TASK_DEFINITION:-db-migrator}
environment=${DEPLOY_ENV:-stg}

# 3. プロジェクト・環境・用途のタグで、アプリ用privateサブネットを取得する。
subnets=$(aws ec2 describe-subnets \
  --filters 'Name=tag:Project,Values=x-clone' "Name=tag:Env,Values=$environment" 'Name=tag:Tier,Values=private' \
  --output json --no-cli-pager)

# サブネットがない場合や、複数のVPCにまたがる場合は誤接続を避けるため停止する。
jq -e '(.Subnets | length) > 0 and ([.Subnets[].VpcId] | unique | length) == 1' <<< "$subnets" >/dev/null || {
  echo 'Expected private subnets in exactly one project VPC' >&2; exit 1;
}

# 4. 同じVPCにあるマイグレーション専用SGを取得する。
vpc=$(jq -r '.Subnets[0].VpcId' <<< "$subnets")
groups=$(aws ec2 describe-security-groups \
  --filters "Name=vpc-id,Values=$vpc" 'Name=group-name,Values=migration' 'Name=tag:Project,Values=x-clone' "Name=tag:Env,Values=$environment" \
  --output json --no-cli-pager)

# SGを1つに特定できなければ停止する。
jq -e '(.SecurityGroups | length) == 1' <<< "$groups" >/dev/null || {
  echo 'Expected exactly one migration security group' >&2; exit 1;
}

# 5. タスクに渡すネットワーク設定をJSONで組み立てる。公開IPは付与しない。
network=$(jq -cn --argjson subnets "$subnets" --argjson groups "$groups" \
  '{awsvpcConfiguration:{subnets:[$subnets.Subnets[].SubnetId],securityGroups:[$groups.SecurityGroups[].GroupId],assignPublicIp:"DISABLED"}}')

# 6. タスク定義をリビジョン付きARNに確定する。
# この後に新リビジョンが登録されても、今回の起動対象が変わらないようにする。
definition=$(aws ecs describe-task-definition --task-definition "$definition" \
  --query 'taskDefinition.taskDefinitionArn' --output text --no-cli-pager)

# マイグレーション以外のタスク定義を誤って起動しないように確認する。
[[ "$definition" == arn:*:task-definition/db-migrator:* ]] || { echo 'Unexpected migration task definition' >&2; exit 1; }

# 7. コンテナ内のGoコマンドへ、--action status または --action upを渡す。
overrides=$(jq -cn --arg action "$action" '{containerOverrides:[{name:"migration",command:["--action",$action]}]}')

# 8. Fargateの単発タスクを1つ起動する。この時点ではSQLの完了は待たない。
echo "Starting $definition on $cluster (action=$action)"
result=$(aws ecs run-task --cluster "$cluster" --task-definition "$definition" \
  --launch-type FARGATE --platform-version 1.4.0 --count 1 \
  --network-configuration "$network" --overrides "$overrides" \
  --output json --no-cli-pager)

# 起動APIの失敗や、想定外のタスク件数を確認する。
if ! jq -e '((.failures // []) | length) == 0 and (.tasks | length) == 1' <<< "$result" >/dev/null; then
  jq '{failures,taskArns:[.tasks[]?.taskArn]}' <<< "$result" >&2
  exit 1
fi

# 9. 起動したタスクのARNを表示し、そのタスクが停止するまで待つ。
task=$(jq -er '.tasks[0].taskArn' <<< "$result")
echo "Task: $task"

# 待機に失敗してもタスクは動いている可能性があるため、自動で再実行しない。
if ! aws ecs wait tasks-stopped --cluster "$cluster" --tasks "$task"; then
  echo "Wait failed. Task may still be running: $task. Inspect before retrying; no new task was started automatically." >&2
  exit 1
fi

# 10. 停止理由とコンテナの終了コードを取得・表示する。
result=$(aws ecs describe-tasks --cluster "$cluster" --tasks "$task" --output json --no-cli-pager)
jq '{failures,tasks:[.tasks[]?|{taskArn,lastStatus,stopCode,stoppedReason,containers:[.containers[]?|{name,exitCode,reason}]}]}' <<< "$result"

# 11. 停止しただけでは成功にしない。取得失敗がなく、対象コンテナが終了コード0で終わったか確認する。
# 起動に失敗したタスクは終了コードがない場合もあるため、それも失敗として扱う。
jq -e '((.failures // []) | length) == 0 and (.tasks | length) == 1
  and .tasks[0].lastStatus == "STOPPED" and .tasks[0].stopCode == "EssentialContainerExited"
  and ([.tasks[0].containers[] | select(.name == "migration")] | length) == 1
  and ([.tasks[0].containers[] | select(.name == "migration")][0].exitCode == 0)' <<< "$result" >/dev/null || {
  echo 'Migration task failed. Check /ecs/backend-migration in CloudWatch Logs.' >&2; exit 1;
}

echo 'Migration task succeeded.'
