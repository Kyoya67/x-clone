# Frontend

X風のタイムライン画面を提供するフロントエンドです。

## ディレクトリ構成

```text
frontend/
├── src/
│   ├── components/      # 再利用可能なUIコンポーネント
│   ├── data/            # モックデータ
│   ├── pages/           # ページ単位のコンポーネント
│   ├── state/           # Contextなどの共有状態
│   ├── test/            # テストセットアップ
│   ├── App.tsx          # ルーティング定義
│   ├── main.tsx         # エントリーポイント
│   └── styles.css       # スタイル
├── index.html
├── package.json
├── package-lock.json
├── tsconfig.json
├── vite.config.ts
├── vitest.config.ts
└── .prettierrc
```

## 使用技術

- React
- TypeScript
- Vite

## 必要な環境

- Node.js 20.12.0以上
- npm（Node.jsに付属）

Node.jsのバージョン確認：

```bash
node --version
npm --version
```

## 環境構築

frontendディレクトリで以下を実行します。

```bash
npm ci
```

`npm ci`は`package-lock.json`に記録された依存関係を利用するため、他の開発者と同じ環境を再現しやすい方法です。

## 開発サーバーの起動

```bash
npm run dev
```

起動後、ターミナルに表示されたURLをブラウザで開いてください。

## ビルド

```bash
npm run build
```

本番用ファイルがdistディレクトリに生成されます。

## テスト

```bash
npm run test
```

ファイル変更を監視しながらテストする場合：

```bash
npm run test:watch
```

## コード整形

```bash
npm run format
npm run format:check
```

## CI

### 実行条件

```yaml
pull_request:
  branches:
    - develop
  paths:
    - 'frontend/**'
    - '.github/workflows/frontend-ci.yml'
```

`develop`向けPull Requestの作成・更新時に、`frontend`ディレクトリまたは`frontend-ci.yml`を変更している場合、CIが実行されます。

```yaml
push:
  branches:
    - develop
  paths:
    - 'frontend/**'
    - '.github/workflows/frontend-ci.yml'
```

push先のブランチが`develop`で、かつ`frontend`ディレクトリまたは`frontend-ci.yml`を変更した場合、CIが実行されます。

### 実行内容

```text
npm ci
  ↓
npm run format:check
  ↓
npm run test
  ↓
npm run build
```

### ローカルでの確認

CIと同じ内容をローカルで確認する場合は、`frontend`ディレクトリで次を実行してください。

```bash
npm ci
npm run format:check
npm run test
npm run build
```

## 現在の実装範囲

- X風のタイムライン画面
- 投稿フォーム
- 投稿のローカル追加
- 文字数カウント
- いいね操作
- React Routerによる画面遷移
- フォロー・フォロー解除とフォロー中タブ
- レスポンシブ表示

現在はモックデータを使用しています。バックエンドAPIとの接続は今後実装します。
