# バックエンドのコンテナ

リポジトリルートで実行する。DockerfileはGo 1.25.5でテスト・vet・ビルドを行い、実行用イメージには静的バイナリ、RDS用CA証明書、Swagger用OpenAPI YAMLを含める。非rootユーザーで起動する。

## RDS用CA証明書

### 外部APIを使わなくても必要な理由

バックエンドからRDSへ送るSQL・認証情報・取得データもネットワークを流れるため、クラウドではTLSで暗号化し、接続先が本物のDBか検証する方針とする。SGは接続元・ポートを制限するものであり、通信の暗号化や証明書検証の代わりにはならない。

RDS側のサーバー証明書はAWSが用意する。バックエンド側に置くのは、それを検証するための公開されたCA証明書であり、DBの秘密鍵や自作・購入したサーバー証明書ではない。[AWSのRDS PostgreSQL TLS設定](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/PostgreSQL.Concepts.General.SSL.html)

### 配置と接続設定

外部HTTPS APIは利用しないため、実行用イメージには一般的なCA一覧をコピーせず、[AWS公式の東京リージョン用RDS CAバンドル](https://docs.aws.amazon.com/ja_jp/AmazonRDS/latest/UserGuide/UsingWithRDS.SSL.html)をビルド時に取得し、`/app/certs/rds-ca-bundle.pem`へ配置する。

配置だけでは証明書検証は有効にならない。RDSへ直接接続する際は、`DATABASE_SSL_MODE=verify-full`と、`DATABASE_URL`のクエリパラメータ`sslrootcert=/app/certs/rds-ca-bundle.pem`を設定する。ローカルDBの`disable`設定は変更しない。RDSへの実接続は未検証。

CAバンドルの更新時は取得ステップのキャッシュを使わず再ビルドし、再デプロイする（例：`docker buildx build --no-cache --platform linux/amd64 --load -t backend:local backend`）。

## ALBの証明書との違い

ALBでHTTPSを受け付ける際は、ACMで管理するサーバー証明書をHTTPSリスナーに設定する方針とする。この証明書はブラウザなどへ提示するもので、バックエンドのDockerイメージには配置しない。[AWSのALB証明書設定](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/https-listener-certificates.html)

| 通信 | サーバー証明書を提示する側 | 証明書を検証する側 | 検証用CAの場所 |
| --- | --- | --- | --- |
| ブラウザ → ALB（HTTPS） | ALB（ACMの証明書） | ブラウザ | ブラウザ・OSの信頼ストア |
| Goバックエンド → RDS（PostgreSQL over TLS） | RDS（AWSが用意する証明書） | Goバックエンド | `/app/certs/rds-ca-bundle.pem` |

ALBで受けたTLS接続と、バックエンドからRDSへのTLS接続は別の接続である。ALBの証明書を設定しても、DB通信が自動で暗号化・検証されるわけではない。また、ALBからバックエンドへの転送区間も別途HTTP/HTTPSを選ぶ必要があり、ALBでのHTTPS受付だけで全区間の暗号化が保証されるわけではない。

### 証明書検証のシーケンス

以下は接続設定後の想定フロー。TLSハンドシェイクの細部は省略している。CA証明書は検証側がローカルで読み取るもので、認証局へ毎回問い合わせるわけではない。

#### ブラウザからALBへのHTTPS接続

`````mermaid
sequenceDiagram
    participant Browser as ブラウザ
    participant ALB as ALB（ACM証明書を設定）
    participant Go as Goバックエンド
    Browser->>ALB: TLS接続を開始
    ALB-->>Browser: サーバー証明書を提示
    Note over Browser: ブラウザ・OSの信頼済みCAを使い<br/>発行元・有効期限・ホスト名を検証
    alt 検証成功
        Browser->>ALB: HTTPSでAPIリクエスト
        ALB->>Go: リクエストを転送（別接続）
        Note over ALB,Go: この区間のHTTP/HTTPSは別途設定
        Go-->>ALB: レスポンス
        ALB-->>Browser: HTTPSでレスポンス
    else 検証失敗
        Note over Browser,ALB: 接続を中止し、証明書エラー
    end
`````

#### GoバックエンドからRDSへのTLS接続

`````mermaid
sequenceDiagram
    participant Go as Goバックエンド
    participant RDS as RDS PostgreSQL
    Note over Go: verify-fullとsslrootcertを指定<br/>コンテナ内のRDS用CAファイルを読み込む
    Go->>RDS: TLS接続を開始
    RDS-->>Go: DBサーバー証明書を提示
    Note over Go: RDS用CAを使い<br/>発行元・有効期限・ホスト名を検証
    alt 検証成功
        Go->>RDS: 暗号化された接続でDB認証
        RDS-->>Go: 認証成功（認証情報が正しい場合）
        Go->>RDS: SQLを送信
        RDS-->>Go: 実行結果
    else 検証失敗
        Note over Go,RDS: 接続を中止（SQLは送らない）
    end
`````

### 現在の実装範囲

- RDS用CAのイメージへの配置、Dockerビルド、ビルド内のGoテスト・vetは確認済み。
- RDSへの`verify-full`接続と、ALBへのACM証明書設定は今後実装・検証する。
- ローカルDBでは引き続き`DATABASE_SSL_MODE=disable`を使用する。
- 将来外部HTTPS APIを利用する場合は、その接続先を検証できる一般的なCA一覧の追加も検討する。

## ビルドと起動確認

`````bash
# ECSの実行アーキテクチャもX86_64に合わせる
docker buildx build --platform linux/amd64 --load -t backend:local backend

# ローカルDBを起動（初回のマイグレーション・seedはbackendの手順に従う）
docker compose -f backend/docker-compose.yml up -d postgres

# ローカル開発用の接続情報。ホストの8080番と競合しないよう18080番を使う
docker run --rm --name backend-check --platform linux/amd64 \
  --network backend_default -p 127.0.0.1:18080:8080 \
  -e DATABASE_URL=postgres://app:app@postgres:5432/app \
  -e DATABASE_SSL_MODE=disable backend:local
`````

別ターミナルから確認する。

`````bash
curl --fail http://localhost:18080/health
curl --fail -o /dev/null -w "%{http_code}\n" http://localhost:18080/docs
curl --fail http://localhost:18080/openapi.yaml
`````

APIは起動時にDB接続を確認するため、DBへ接続できなければ終了する。本番の接続先・認証情報・SSL設定は実行時に注入し、イメージに含めない。マイグレーションやseedはコンテナ起動時に自動実行しない。

## ECRへのpush

変更をコミットした後、リポジトリルートで実行する。AWS CLI・Docker・makeと、`x-clone-terraform-stg`プロフィールの設定が必要。

`````bash
# API用イメージ
make -C backend api-ecr-push

# マイグレーション用イメージ
make -C backend migration-ecr-push
`````

`backend`ディレクトリ内なら`-C backend`は不要。[Makefile](../Makefile)では入口を分け、共通の`_ecr-push`処理へリポジトリとDockerfileのビルド対象を渡す。以前の`ecr-push`は`api-ecr-push`へ置き換えた。

| コマンド | Dockerfileのステージ | ECRリポジトリ |
| --- | --- | --- |
| api-ecr-push | app | backend |
| migration-ecr-push | migration | backend-migration |

共通処理は以下を順番に行い、どこかで失敗した場合は後続処理を停止する。

1. 現在のGitコミットハッシュの先頭6桁をタグにする。
2. AWS CLIでECRへログインする（トークンは標準入力でDockerへ渡す）。
3. 現在のコードから`linux/amd64`向けイメージをビルドする。Dockerfile内のテスト・vetも実行する。
4. `089244387218.dkr.ecr.ap-northeast-1.amazonaws.com/<リポジトリ>:<タグ>`へpushする。

AWS接続先はMakefileの`ECR_PROFILE`・`ECR_REGION`・`ECR_REGISTRY`で指定している。`ECR_REPOSITORY`・`DOCKER_TARGET`は各入口で明示する。共通処理を直接実行して対象が未指定なら停止する。未コミットの変更もビルド対象になるため、タグとソースを対応させるには実行前にコミットすること。

ECRはIMMUTABLEのため同じタグを上書きできない。ライフサイクルポリシーはpush日時が新しい3イメージを保持する。通常は変更をコミットしてからビルドする。初回の`60e889`は、当該コミットのバックエンドコードに未コミットのDockerfileを加えてビルドしたイメージ。
