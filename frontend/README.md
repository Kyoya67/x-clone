# Frontend

X風のタイムライン画面を提供するフロントエンドです。

## 使用技術

- React
- TypeScript
- Vite

## 必要な環境

- Node.js 20以上を推奨
- npm

## 環境構築

frontendディレクトリで以下を実行します。

`````
npm install
`````

## 開発サーバーの起動

`````
npm run dev
`````

起動後、ターミナルに表示されたURLをブラウザで開いてください。

## ビルド

`````
npm run build
`````

本番用ファイルがdistディレクトリに生成されます。

## 現在の実装範囲

- X風のタイムライン画面
- 投稿フォーム
- 投稿のローカル追加
- 文字数カウント
- いいね操作
- レスポンシブ表示

現在はモックデータを使用しています。バックエンドAPIとの接続は今後実装します。

## ディレクトリ構成

`````
frontend/
├── src/
│   ├── App.tsx          # アプリケーション本体
│   ├── main.tsx         # エントリーポイント
│   ├── styles.css       # スタイル
│   └── vite-env.d.ts    # Viteの型定義
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.ts
`````
