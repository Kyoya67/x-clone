# サブネット構成

| 名前 | CIDR | 用途 | デフォルトルート |
| --- | --- | --- | --- |
| public-1a | 10.0.0.0/24 | NATインスタンス、将来のALB | Internet Gateway |
| public-1c | 10.0.1.0/24 | 将来のALB | Internet Gateway |
| private-1a | 10.0.20.0/24 | ECS・RDSの配置候補 | NATインスタンス |
| private-1c | 10.0.21.0/24 | ECS・RDSの配置候補 | NATインスタンス |

RDSはSingle-AZを維持し、2つのprivateサブネットをDBサブネットグループへ指定する。API・マイグレーション・RDSのSGは引き続き分離する。DBの送信許可は追加しない。

## 既存環境からの移行

- 旧db-privateのサブネット・ルートテーブル・関連付けはsubnet_moves.tfで引き継ぐ。RDS本体とDBサブネットグループのIDは変更しない。
- 共有privateへNATの経路を追加し、APIサービスの配置先を変更する。
- 旧app-privateの2サブネット、ルートテーブル、関連付け、NAT受信ルールを削除する。
- 単発マイグレーションはTier=privateで配置先を検索する。apply前は変更後のスクリプトを実行しない。

apply前に単発タスクが旧サブネットで動作していないことを確認する。切り替え中は旧タスクのNAT通信が途切れる可能性がある。ENIが残って削除が失敗した場合は、所有するタスクを確認し、終了・ENI解放を待ってplanを確認する。ENIを手動で強制削除しない。

planでRDS本体・DBサブネットグループ・保持する4サブネットの再作成がないことを確認する。2026-09-11のplanは2追加・7変更・7削除。変更には既存のrds.force_sslのapply_method差分も含む。

apply後はAPIのHEALTHY、make migrate-rds-statusの成功、SSM経由のDB接続、サブネットが4つであることを確認する。実環境への適用・移行後の確認は未実施。
