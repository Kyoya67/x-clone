# DBユーザーとマイグレーション

## 目指す流れ

1. dbadminで migration_user と app_user を作る。
2. migration_userで最初のマイグレーションを実行する。
3. 作ったテーブルの読み書き権限をapp_userへ付与する。

ユーザー作成自体には、テーブルは必要ない。

## ユーザーの役割

- dbadmin：初期設定でユーザー・権限を用意する。
- migration_user：テーブルを作成・変更する。CI/CDのマイグレーションでも使用する。
- app_user：アプリからデータを読み書きする。

## ECSとCI/CD

開発者が単発のECSタスクを起動し、タスク内からmigration_userでDBへ接続する構成を目指す。将来は同じタスクをCI/CDから起動する。

CI/CDにはタスクを起動するAWSのIAM権限を与える。DB接続には引き続きmigration_userを使い、その認証情報はSecrets Managerからタスクへ渡す。CI/CD自体にDBのパスワードを渡す必要はない。

## 現在の実装との差

- 現在のマイグレーションタスクはdbadminで接続する。
- db-userコマンドが作るのはapp_userのみ。migration_userの作成・接続への切り替えは未実装。
- 既存テーブルはdbadminで作成済みのため、切り替え時に所有権も整理する。
- CI/CDからの起動は未実装。
