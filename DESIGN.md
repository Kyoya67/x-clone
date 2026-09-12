# Design Doc

## 1. 開発方針

```text
課題全体の完成
├── Milestone 1: 初期実装・ローカル開発基盤
│   ├── issue: フロントエンドの初期構成と画面を実装する
│   ├── issue: Goバックエンドの初期構成を作成する
│   ├── issue: データベース環境を構築する
│   ├── issue: バックエンドAPIの基本構成を整備する
│   ├── issue: 投稿機能をバックエンドから実装する
│   ├── issue: フォロー機能をフロントエンドからデータベースまで実装する
│   └── issue: タイムライン機能を実装する
└── Milestone 2: Terraform・CI/CDを含むデプロイ基盤の構築
    ├── issue: AWS stg環境を構築し、ECSでAPI起動とマイグレーションを確認する
    ├── issue: AWSリソースの役割・依存関係・命名を整理する
    ├── issue: DBユーザー管理・マイグレーション実行基盤を整理する
    ├── issue: フロントエンドをstg環境へデプロイし、API・RDSとの疎通を確認する
    └── issue: ECRへのイメージpushとECSへの反映を自動化する
```

Milestoneは関連するIssueを段階ごとにまとめ、Issueは個別の作業を管理する。

Issueごとに`develop`ブランチから作業ブランチを作成し、作業完了後に`develop`へのPull Requestを作成する。各Issueの変更を`develop`に統合し、必要な単位で`main`へ反映する。

## 2. AIの活用方針

### 使用したツール

- OpenAI Codex
- 作業ルール：[AGENTS.md](../AGENTS.md)

### Codexに任せたこと

- コードの初期実装
- テストコードの作成案
- エラー原因の調査
- 実装方針の比較案作成

### 自分で判断・検証したこと

- 採用する技術とアーキテクチャ
- APIとデータモデルの設計
- Codexが生成したコードのレビュー
- テスト実行と動作確認
- Design Docの内容

### 活用時に意識したこと

Codexの提案をそのまま採用せず、実装内容を確認し、テストや動作確認を行ったうえで採用可否を判断した。

## 3. 要件の整理

### 機能要件

本課題では、Twitter APIクライアントではなく、サービスとしてのTwitterクローンを開発する。

#### 必須要件（課題文で指定）

- ツイートする
- 他ユーザーをフォローする
- タイムラインを見る

#### 追加・補足要件

- おすすめユーザーは、フォロー機能の動作確認用として固定のモックデータを表示する
- おすすめ対象の選定ロジックやランキングは、今回の必須要件の対象外とする

### 非機能要件

#### 必須要件（課題文で指定）

- WebフロントエンドはTypeScriptとReactを利用する
- バックエンドはGo言語を利用する
- 3〜6名程度のチームでアジリティ高く開発でき、新しいメンバーも参加しやすい構成にする
- 各開発者のラップトップ上で普段の開発を行えるようにする
- AWSなどのパブリッククラウドにデプロイし、顧客へ提供できることを考慮する
- ユーザー体験、将来の拡張性、セキュリティ、エラーハンドリングなどの横断的な関心事を考慮する

#### 追加・補足要件

- 現段階では、必須要件の実装と検証を優先する
- 必須要件に含まれないおすすめユーザーの選定ロジックは、必要性を判断したうえで将来の拡張課題とする

## 4. 技術選定とトレードオフ

### フロントエンドの技術構成

#### UIライブラリ・フレームワーク

| 採用 | 選択肢  | 役割                  | 判断理由                                                                   |
| ---- | ------- | --------------------- | -------------------------------------------------------------------------- |
| ⭕️   | React   | UIライブラリ          | 課題の指定技術。Go APIと分離したSPAを構築する                              |
| —    | Next.js | Reactのフレームワーク | SSR、Server Components、Next.js側のAPIなどが今回不要。別サーバー層も増える |

#### ビルドツール・バンドラー

| 採用 | 選択肢    | 役割                                           | 判断理由                                                  |
| ---- | --------- | ---------------------------------------------- | --------------------------------------------------------- |
| ⭕️   | Vite      | 開発サーバーと本番ビルドを提供するビルドツール | 設定がシンプルで起動・変更反映が速く、React SPAに必要十分 |
| —    | webpack   | バンドラーと開発基盤                           | 実績は豊富だが、今回のSPAでは設定や周辺ツールが過剰       |
| —    | Rspack    | webpack互換を意識した高速なバンドラー          | webpack互換性や大規模ビルドが現時点で不要                 |
| —    | Turbopack | Next.jsなどで利用される高速なバンドラー        | 今回はNext.jsを採用しないため対象外                       |

#### JavaScript・TypeScript変換

| 採用 | 選択肢            | 役割                             | 判断理由                                         |
| ---- | ----------------- | -------------------------------- | ------------------------------------------------ |
| ⭕️   | SWC / esbuildなど | TypeScriptや最新JavaScriptの変換 | Viteの設定に任せ、個別の変換基盤を追加しない     |
| —    | Babel             | JavaScript・TypeScript変換       | 柔軟で実績はあるが、今回の構成では追加設定が不要 |

#### SPAのルーティング

| 採用 | 選択肢             | 役割                       | 判断理由                                                     |
| ---- | ------------------ | -------------------------- | ------------------------------------------------------------ |
| ⭕️   | React Router v7    | SPAのURLと画面を対応付ける | React + Viteに導入しやすく、現時点ではDeclarative Modeで十分 |
| —    | Next.js App Router | Next.jsのルーティング      | Next.jsを採用しないため今回は利用しない                      |

#### テスト基盤

| 採用 | 選択肢 | メリット                                         | トレードオフ・判断理由                                                   |
| ---- | ------ | ------------------------------------------------ | ------------------------------------------------------------------------ |
| ⭕️   | Vitest | Viteと設定やモジュール変換の考え方を共有しやすい | 今回のVite構成に合わせやすいため採用                                     |
| —    | Jest   | Reactを含むエコシステムで実績と情報量が多い      | Viteとは別に変換設定などを整える必要があり、今回はVitestより設定が増える |

### フロントエンドの配信方式：S3＋CloudFrontとAmplify Hosting

採用方式は未決定。[Issue #27](https://github.com/Kaminashi-Inc/ENG-1103_Kyoya67/issues/27)で、stg.x-clone.kyo8.devへの公開とAPIへの疎通を実装する。

#### 比較の前提

- React＋ViteのSPAをビルドした静的ファイルを配信する。フロント用の常駐サーバーは不要。
- どちらを選んでも、Go APIは既存のECS、DBは既存のRDSを使う。
- AmplifyはHostingのみを比較対象とし、認証・DBなどのAmplifyバックエンド機能への置き換えは行わない。Viteの成果物distを配信できる。[Vite配信の公式手順](https://docs.amplify.aws/gen1/javascript/deploy-and-host/frameworks/deploy-vite-site/)

#### 比較

| 観点 | S3＋CloudFront | Amplify Hosting |
| --- | --- | --- |
| 役割 | S3で静的ファイルを保管し、CloudFrontで配信する | ビルド・デプロイ・配信をまとめて管理する |
| 構築・運用の負担 | バケットのアクセス制御、CDN、キャッシュ、デプロイ手順を自分で組み立てる | 設定対象をアプリ・ブランチ・ビルド設定に集約しやすい |
| デプロイ・CI/CD | CIでビルドし、S3へ配置。キャッシュ更新や旧成果物の保持・戻し方も設計する | Git連携による自動ビルド・デプロイを利用できる。[公式概要](https://docs.aws.amazon.com/amplify/latest/userguide/welcome.html) |
| Terraform管理 | S3・CloudFront・IAM・DNSなどを個別に管理できる | アプリ・ブランチ・ドメイン等を管理できる。配信基盤の内部リソースを個別管理する方式ではない。[AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/6.34.0/docs/resources/amplify_app.html) |
| 独自ドメイン・HTTPS | DNSとCloudFrontの証明書を設定する。ACM証明書はus-east-1に用意する。[証明書要件](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/cnames-and-https-requirements.html) | 独自ドメインを関連付け、Amplify管理の証明書を利用できる。[独自ドメイン](https://docs.aws.amazon.com/amplify/latest/userguide/custom-domains.html) |
| 配信元の保護 | S3を非公開にし、OAC経由でCloudFrontからのみ取得させる構成にする。[オリジン設定](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/DownloadDistS3AndCustomOrigins.html) | 静的配信用のS3アクセス制御を自分で組み立てる必要はない。アプリやAPIの利用者認証は別途必要 |
| APIへの接続 | CloudFrontで静的ファイルとAPIの転送先を分ける、またはブラウザから別APIドメインへ接続する | HTTPSのAPIへのリバースプロキシrewrite、またはブラウザから別APIドメインへ接続する。[rewrite設定](https://docs.aws.amazon.com/amplify/latest/userguide/redirect-rewrite-examples.html#reverse-proxy-rewrite) |
| SPAの再読み込み | 画面URLをindex.htmlへ振り向ける処理を設計する | SPA向けrewriteを設定する。APIへの転送ルールとは分ける |
| 費用 | S3の保存・リクエスト、CloudFrontの利用プラン・通信量、CIの費用などを合算 | ビルド時間、成果物の保存、配信量などが課金対象 |
| 今回との相性（評価） | 配信・API経路・権限を細かく設計し、Terraformで追いたい場合に向く | フロント配信の管理を減らし、アプリの公開・動作確認を早く進めたい場合に向く |

費用はアクセス量・ビルド頻度・無料枠・利用プランで変わるため、現時点ではどちらが安いか断定しない。比較時は同じ条件で見積もる。既存のECS・RDSや、追加するALB等の費用は別途必要。[S3料金](https://aws.amazon.com/s3/pricing/)、[CloudFront料金](https://aws.amazon.com/cloudfront/pricing/)、[Amplify料金](https://aws.amazon.com/amplify/pricing/)

#### 今回、どちらでも必要な対応

- ECS APIへの入口を用意する。フロントを公開するだけでは、プライベートサブネットのAPIへ接続できない。
- 現在のフロントは/apiを呼び、Viteの開発用プロキシがプレフィックスを除去している。公開環境にも同等の転送・パス変換を用意するか、APIのURL設計を変更する。
- 別オリジンのAPIを呼ぶ場合はCORSを設定する。同一オリジンにする場合はAPI転送とSPAのrewriteを分け、APIエラーをindex.htmlに置き換えない。
- 投稿・フォローの書き込みと利用者別の応答を考慮し、APIのHTTPメソッド・ヘッダー転送・キャッシュ方針を確認する。
- Amplify採用時は、このリポジトリのfrontendをアプリルートとしてビルド設定する。[モノレポ設定](https://docs.aws.amazon.com/amplify/latest/userguide/monorepo-configuration.html)

#### 判断基準

- 配信基盤の制御・学習を優先するならS3＋CloudFront。
- フロントのデプロイ運用を簡単にするならAmplify Hosting。
- Terraformを使うことだけを理由にAmplifyを除外しない。Amplifyを採用しても、ECSのデプロイやDBマイグレーションのCI/CDは別途設計する。

### バックエンド

| 採用 | 選択肢                   | 役割             | 判断理由                                                                                                            |
| ---- | ------------------------ | ---------------- | ------------------------------------------------------------------------------------------------------------------- |
| ⭕️   | Go                       | APIサーバー      | 課題の指定技術。コンパイル時の型検査と標準ライブラリのHTTP・context機能を活用し、単一バイナリとしてデプロイできる。 |
| ⭕️   | `net/http` + gorilla/mux | HTTPルーティング | HTTPの標準的なインターフェースを維持しつつ、HTTPメソッド・パスパラメータごとのルート定義を簡潔に記述できる。        |
| —    | Go標準ライブラリのみ     | HTTPルーティング | 依存を減らせる一方、現時点で必要なパスパラメータの取得やルート定義が冗長になりやすい。                              |
| —    | chi                      | HTTPルーティング | 軽量でGoらしい選択肢だが、gorilla/muxで必要な機能を満たしており、移行する理由がない。                               |

### データベース

| 採用 | 選択肢     | 役割     | 判断理由                                                                                              |
| ---- | ---------- | -------- | ----------------------------------------------------------------------------------------------------- |
| ⭕️   | PostgreSQL | アプリDB | 外部キー、制約、トランザクションを厳密に扱いやすい。JOIN、集約、JSONBなどの機能も豊富で拡張しやすい。 |
| —    | MySQL      | アプリDB | 一般的なCRUD中心のサービスでは十分な選択肢だが、今回はPostgreSQLの型・クエリ・拡張性を優先した。      |

### インフラ基盤

| 採用 | 選択肢          | 役割         | 判断理由                                                                                                                     |
| ---- | --------------- | ------------ | ---------------------------------------------------------------------------------------------------------------------------- |
| ⭕️   | AWS             | クラウド基盤 | 課題でパブリッククラウドへのデプロイを考慮する必要があり、ECS、RDS、ECR、Secrets Managerなどを一つのクラウド内で構成できる。 |
| —    | Google Cloud    | クラウド基盤 | Cloud RunやCloud SQLなどで同等構成は可能だが、今回はAWSの学習・検証を優先する。                                              |
| —    | Azure           | クラウド基盤 | コンテナ実行やDBのマネージドサービスは利用できるが、今回の検証対象から外す。                                                 |
| —    | Heroku / Render | PaaS         | 構築は簡単だが、VPC、IAM、セキュリティグループ、RDS相当の構成管理を学習・検証しづらい。                                      |

### IaC

| 採用 | 選択肢             | 役割     | 判断理由                                                                                                                    |
| ---- | ------------------ | -------- | --------------------------------------------------------------------------------------------------------------------------- |
| ⭕️   | Terraform          | IaC      | AWSリソースを宣言的に管理でき、クラウドやサービスをまたいだ構成にも対応しやすい。Stateにより差分確認と再作成もしやすい。    |
| —    | AWS CloudFormation | IaC      | AWS純正で統合度は高いが、記述量が増えやすく、今回はTerraformの汎用性を優先した。                                            |
| —    | AWS CDK            | IaC      | プログラミング言語で抽象化できる一方、生成されるCloudFormationの理解も必要になるため、今回はTerraformでリソースを明示する。 |
| —    | Pulumi             | IaC      | GoやTypeScriptで書けるが、今回の目的ではTerraformの情報量と学習効果を優先した。                                             |
| —    | 手動構築           | 構築方法 | 初期検証は速いが、再現性が低く、変更履歴やレビューが残りにくいため採用しない。                                              |

AWSリソース構成、セキュリティグループ、IAM、DBユーザー、Secret管理方針の詳細は[infrastructure/ARCHITECTURE.md](infrastructure/ARCHITECTURE.md)にまとめる。

## 5. システム構成とデータモデル

### バックエンドの構成

Goバックエンドは、HTTPサーバーの起動、ルーティング、HTTPリクエスト処理、業務ロジック、データアクセスの責務を分ける。`cmd/api`、`internal/controllers`、`internal/services`、`internal/repositories`、`internal/models`を使用し、投稿・フォロー・タイムラインを実装する。

```text
cmd/api/main.go
    │ DB初期化・サーバー起動
    ▼
internal/routers/router.go
    │ service・controllerの生成、ルート登録
    ▼
internal/controllers/
    │ HTTPリクエスト・レスポンス処理
    ▼
internal/services/
    │ 業務ロジック
    ▼
internal/repositories/ ── DBアクセス
    │
internal/models/ ─────── データモデル
```

`main.go`ではDBを初期化して`routers.NewRouter(db)`へ渡す形を維持する。controllerを個別に`NewRouter`の引数へ追加せず、必要なcontrollerやserviceの生成は`router.go`に集約することで、アプリケーションの組み立てとルート定義を追いやすくする。health endpointのようにDBやserviceを必要としない処理は、controller単体で実装する。

### データベースの構成

ユーザー、投稿、フォロー関係はそれぞれ1つのテーブルで管理する。ユーザーごとにテーブルを作成するのではなく、`follows`テーブルの各行で「誰が誰をフォローしたか」を表す。

```mermaid
erDiagram
    users ||--o{ posts : "author_id"
    users ||--o{ follows : "follower_id"
    users ||--o{ follows : "followee_id"

    users {
        UUID id PK
        VARCHAR handle UK
        VARCHAR display_name
        TEXT bio
        TIMESTAMPTZ created_at
    }

    posts {
        UUID id PK
        UUID author_id FK
        VARCHAR content
        TIMESTAMPTZ created_at
    }

    follows {
        UUID follower_id PK, FK
        UUID followee_id PK, FK
        TIMESTAMPTZ created_at
    }
```

`follows`は、同じ`users`テーブルを2つの役割で参照する。たとえば田中が佐藤をフォローすると、`follower_id`は田中のID、`followee_id`は佐藤のIDとなる。`(follower_id, followee_id)`を複合主キーにすることで、同じユーザーを重複してフォローできない。また、`follower_id <> followee_id`の制約により、自分自身のフォローを防ぐ。

タイムライン取得時は、`posts.author_id`と`users.id`を結合して投稿者情報を取得する。`following`タイムラインでは、さらに`follows.followee_id`と投稿者IDを結合し、`follows.follower_id`が現在のユーザーである投稿だけを残す。現時点の`for-you`は推薦機能ではなく、全投稿を新しい順で表示する。

## 6. 今後の拡張性や運用を見据えた懸念点
