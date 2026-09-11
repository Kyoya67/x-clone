# NATインスタンスの採用理由と動作確認

## 採用理由

stgでは、Private Subnetのアプリから外部APIやパッケージ配布元へアクセスするため、Public SubnetのEC2をNATインスタンスとして使う。小規模な検証環境のため、NAT Gatewayの時間課金・データ処理課金を避け、固定費を抑えることを優先した。現在のインスタンスタイプはt3.micro。

NATインスタンスにもEC2・EBS・Public IPv4などの料金が発生する。常にNAT Gatewayより安いと保証するものではなく、利用量に応じて費用を確認する。

NAT GatewayはAWSが管理するが、NATインスタンスではOS更新、転送設定、監視、障害復旧を自分たちで行う。現在は1台のため、停止や配置先AZの障害でPrivate Subnetからの外部通信が止まる。prdの冗長化・運用方針は別途検討する。

参考: [AWSのNATインスタンス説明](https://docs.aws.amazon.com/vpc/latest/userguide/VPC_NAT_Instance.html)

## ブラウザからの接続

EC2コンソールの「EC2 Instance Connect」で、Public IPv4とユーザー名ec2-userを指定する。NAT用Security Groupは、AWS管理プレフィックスリスト `com.amazonaws.ap-northeast-1.ec2-instance-connect` からTCP 22への接続を許可する。設定は `infrastructure/modules/security_group/main.tf` で管理する。

## 確認コマンド

NATインスタンスのターミナルで実行する。

`````bash
# cloud-initによる初回起動設定が終了するまで待つ
sudo cloud-init status --wait

# LinuxのIPv4パケット転送が有効か確認する（1なら有効）
sysctl net.ipv4.ip_forward

# 送信元IPを書き換えるMASQUERADEルールとパケット数を表示する
# -t nat: NATテーブル、-L: 一覧、-n: 名前解決なし、-v: 詳細表示
sudo iptables -t nat -L POSTROUTING -n -v

# 中継パケットを扱うFORWARDチェーンの許可・拒否ルールを確認する
sudo iptables -L FORWARD -n -v

# OS起動時にiptablesの保存済みルールを読み込む設定か確認する
sudo systemctl is-enabled iptables

# iptablesサービスが有効な状態にあるか確認する
sudo systemctl is-active iptables

# NATインスタンス自身から外部HTTPSへ接続する（ヘッダーのみ取得）
curl -I --connect-timeout 10 https://aws.amazon.com
`````

## 確認結果（2026-09-10）

ユーザーがEC2コンソールのターミナルで実行した結果を記録する。HTTPレスポンスは検証に必要なステータスのみ抜粋した。

| 確認項目 | 実行結果 | 分かったこと |
| --- | --- | --- |
| cloud-init | status: done | 初回起動設定が完了 |
| IPv4転送 | net.ipv4.ip_forward = 1 | IP転送が有効 |
| POSTROUTING | MASQUERADE / out: ens5 / 329 packets / 23327 bytes | 送信元IP変換ルールが存在し、一致したパケットがある |
| FORWARD | policy ACCEPT / 追加ルールなし | このチェーンでは転送を拒否していない |
| iptables自動起動 | enabled | サービスの自動起動が設定済み |
| iptables稼働状態 | active | サービスがactive状態 |
| 外部HTTPS | HTTP/2 200 | NATインスタンス自身から外部に接続できる |

これらはOS側の設定と、NATインスタンス自身の外部通信を確認した結果である。MASQUERADEのカウンターにはインスタンス自身が送った通信も含まれ得るため、Private Subnetからの転送成功までは証明しない。再起動後の動作も未確認。

## 次の確認

Private Subnet内のEC2またはECSから、上記のcurlを実行して成功することを確認する。同時にNATインスタンスのMASQUERADEカウンター増加を確認する。これにより、Private Subnet → NATインスタンス → Internet Gateway → インターネットの経路を検証する。

初期設定に問題がある場合は次でログを確認する。

`````bash
# cloud-initが実行した起動スクリプトの出力を確認する
sudo tail -n 100 /var/log/cloud-init-output.log
`````
