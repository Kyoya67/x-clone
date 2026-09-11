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

Secret名は「db/dbadmin」「db/app_user」「db/migration_user」。

3つともusername・passwordのJSON形式で保存する。既存のapp_user・migration_userのURLは、make db-userの再実行で同じパスワードのJSONへ変換する。APIはDB_USER・DB_PASSWORDをSecretから取得し、接続先はECSの環境変数で指定する。

Secretの形式変更に合わせて、新しいAPIイメージのビルド・pushとタスク定義の適用が必要。旧APIイメージのまま新しい環境変数設定で起動しない。

管理コマンドからのSecret保存には、所有者だけが読み書きできる一時ファイル（0600）を使う。通常の成功・失敗時には削除するが、強制終了時は残る可能性がある。

dbadmin用SecretとRDS自動管理の解除設定は追加済み。パスワードは実行時に渡し、Stateに残さず同じ値をSecretとRDSへ設定する。更新時は両方で共通のパスワード版番号を増やす。失敗時の再実行には同じパスワードを使う。

infrastructure/stg/.env.exampleを.envへコピーし、2つのTF_VAR値を設定する。同ディレクトリのmake plan・make applyが.envを読み込む。.envはGit管理対象外だが、ローカルには平文で保存される。terraformコマンドを直接実行する場合は自動では読み込まない。

ユーザーによるRDS変更のapplyは完了。管理コマンドはdb/dbadminを使用する。APIタスクはdb/app_user、マイグレーションタスクはdb/migration_userを参照する設定。

既存Secretの名前変更は置き換えになる。旧Secretに値がある場合は、削除前に新Secretへ同じ値を移してから参照を切り替える。Terraformは値をコピーしないため、そのままapplyしない。

- マイグレーションのGoコードはmigration_userのみ許可する。タスク実行ロールもdb/migration_userのみ取得を許可する。
- db-userコマンドはapp_userとmigration_userを作成する。認証情報は別々のSecrets Managerへ保存する。migration_user用の保存先をTerraformで作成してから実行する。実DBへの適用は未確認。
- マイグレーション用イメージを再ビルド・pushし、stg/aws.tfのタグ更新とapply後に単発実行で確認する。今回のコード変更はAWSへ未反映。
- make db-userは既存のusers・posts・follows・schema_migrationsの所有権をmigration_userへ変更する。未作成のテーブルはスキップする。app_userの読み書き権限は維持する。実RDSへの適用は未確認。
- CI/CDからの起動は未実装。
