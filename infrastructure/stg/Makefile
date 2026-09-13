SHELL := /bin/bash
TERRAFORM ?= terraform

.PHONY: plan apply

# .envの値を環境変数としてTerraformへ渡す。値はコマンド表示に出さない。
plan apply:
	@set -euo pipefail; \
	test -f .env || { echo '.env is required; copy .env.example and set the values'; exit 1; }; \
	set -a; \
	source ./.env; \
	set +a; \
	test -n "$${TF_VAR_dbadmin_password:-}" || { echo 'TF_VAR_dbadmin_password is required'; exit 1; }; \
	test -n "$${TF_VAR_dbadmin_password_version:-}" || { echo 'TF_VAR_dbadmin_password_version is required'; exit 1; }; \
	$(TERRAFORM) $@
