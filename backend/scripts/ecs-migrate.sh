#!/usr/bin/env bash
set -euo pipefail

# AWS_PROFILE/AWS_REGION are inherited; CI may use an assumed role instead.
action=${1:-status}
case "$action" in status|up) ;; *) echo 'Action must be status or up' >&2; exit 1 ;; esac
command -v jq >/dev/null || { echo 'jq is required' >&2; exit 1; }
cluster=${ECS_CLUSTER:-app}
definition=${ECS_MIGRATION_TASK_DEFINITION:-backend-migration}
environment=${DEPLOY_ENV:-stg}

# Resolve only this project's network, and reject ambiguous results.
subnets=$(aws ec2 describe-subnets \
  --filters 'Name=tag:Project,Values=x-clone' "Name=tag:Env,Values=$environment" 'Name=tag:Tier,Values=app-private' \
  --output json --no-cli-pager)
jq -e '(.Subnets | length) > 0 and ([.Subnets[].VpcId] | unique | length) == 1' <<< "$subnets" >/dev/null || {
  echo 'Expected app-private subnets in exactly one project VPC' >&2; exit 1;
}
vpc=$(jq -r '.Subnets[0].VpcId' <<< "$subnets")
groups=$(aws ec2 describe-security-groups \
  --filters "Name=vpc-id,Values=$vpc" 'Name=group-name,Values=migration' 'Name=tag:Project,Values=x-clone' "Name=tag:Env,Values=$environment" \
  --output json --no-cli-pager)
jq -e '(.SecurityGroups | length) == 1' <<< "$groups" >/dev/null || {
  echo 'Expected exactly one migration security group' >&2; exit 1;
}
network=$(jq -cn --argjson subnets "$subnets" --argjson groups "$groups" \
  '{awsvpcConfiguration:{subnets:[$subnets.Subnets[].SubnetId],securityGroups:[$groups.SecurityGroups[].GroupId],assignPublicIp:"DISABLED"}}')

# Resolve the revision once so the run is not affected by a later registration.
definition=$(aws ecs describe-task-definition --task-definition "$definition" \
  --query 'taskDefinition.taskDefinitionArn' --output text --no-cli-pager)
[[ "$definition" == arn:*:task-definition/backend-migration:* ]] || { echo 'Unexpected migration task definition' >&2; exit 1; }
overrides=$(jq -cn --arg action "$action" '{containerOverrides:[{name:"migration",command:["--action",$action]}]}')
echo "Starting $definition on $cluster (action=$action)"
result=$(aws ecs run-task --cluster "$cluster" --task-definition "$definition" \
  --launch-type FARGATE --platform-version 1.4.0 --count 1 \
  --network-configuration "$network" --overrides "$overrides" \
  --output json --no-cli-pager)
if ! jq -e '((.failures // []) | length) == 0 and (.tasks | length) == 1' <<< "$result" >/dev/null; then
  jq '{failures,taskArns:[.tasks[]?.taskArn]}' <<< "$result" >&2
  exit 1
fi
task=$(jq -er '.tasks[0].taskArn' <<< "$result")
echo "Task: $task"
if ! aws ecs wait tasks-stopped --cluster "$cluster" --tasks "$task"; then
  echo "Wait failed. Task may still be running: $task. Inspect before retrying; no new task was started automatically." >&2
  exit 1
fi
result=$(aws ecs describe-tasks --cluster "$cluster" --tasks "$task" --output json --no-cli-pager)
jq '{failures,tasks:[.tasks[]?|{taskArn,lastStatus,stopCode,stoppedReason,containers:[.containers[]?|{name,exitCode,reason}]}]}' <<< "$result"
# STOPPED alone is not success: startup failures may have no exit code.
jq -e '((.failures // []) | length) == 0 and (.tasks | length) == 1
  and .tasks[0].lastStatus == "STOPPED" and .tasks[0].stopCode == "EssentialContainerExited"
  and ([.tasks[0].containers[] | select(.name == "migration")] | length) == 1
  and ([.tasks[0].containers[] | select(.name == "migration")][0].exitCode == 0)' <<< "$result" >/dev/null || {
  echo 'Migration task failed. Check /ecs/backend-migration in CloudWatch Logs.' >&2; exit 1;
}
echo 'Migration task succeeded.'
