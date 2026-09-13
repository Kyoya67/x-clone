# Goテスト

## internal/controllers

Controllerの依存先はFakeへ差し替え、HTTPリクエストに対するレスポンスを検証する。

`post.go`では、次の観点をテストしている。

- 正常なリクエストで`201 Created`と投稿データを返す
- Serviceが返した内部エラーを`500 Internal Server Error`へ変換する
- DBなどの内部エラーの詳細をレスポンスへ含めない

`follow.go`では、次の観点をテストしている。

- フォロー・フォロー解除で`204 No Content`を返す
- URLの`userId`をServiceへ渡す
- UUID形式でない`userId`を拒否する
- Serviceの入力エラーを`400 Bad Request`へ変換する
- Serviceの内部エラーを`500 Internal Server Error`へ変換し、詳細をレスポンスへ含めない
- ログインユーザーがフォローしているユーザーIDを`200 OK`で返す
- フォロー一覧のJSON書き込み失敗時は`500 Internal Server Error`を返す

`health.go`では、次の観点をテストしている。

- ヘルスチェックで`200 OK`と`{"status":"ok"}`を返す

`swagger.go`では、次の観点をテストしている。

- Swagger UIのHTMLを返す
- OpenAPI仕様ファイルを返す
- OpenAPI仕様ファイルが存在しない場合は`500 Internal Server Error`を返す
- 複数の探索パスからOpenAPI仕様ファイルを読み取る

`timeline.go`では、次の観点をテストしている。

- タイムライン取得結果をJSONで返す
- `feed`未指定時は`for_you`として扱う
- Serviceの内部エラーを`500 Internal Server Error`へ変換し、詳細をレスポンスへ含めない
- タイムラインのJSON書き込み失敗時は`500 Internal Server Error`を返す

### カバレッジ結果

| ファイル  | 関数                  | カバレッジ |
| --------- | --------------------- | ---------: |
| follow.go | NewFollowController   |     100.0% |
|           | Follow                |     100.0% |
|           | Unfollow              |     100.0% |
|           | ListFollowing         |     100.0% |
|           | handleFollowAction    |     100.0% |
|           | followeeIDFromRequest |     100.0% |
| health.go | NewHealthController   |     100.0% |
|           | Health                |     100.0% |
| post.go   | NewPostController     |     100.0% |
|           | Create                |     100.0% |
| swagger.go | SwaggerUI            |     100.0% |
|           | OpenAPISpec           |     100.0% |
|           | readOpenAPISpec       |     100.0% |
| timeline.go | NewTimelineController |   100.0% |
|           | List                  |     100.0% |
| 合計      | -                     |     100.0% |

計測コマンド：

```bash
cd backend
go test ./internal/controllers -coverprofile=/tmp/controller-cover.out
go tool cover -func=/tmp/controller-cover.out
```

## internal/services

Serviceの依存先はFakeへ差し替え、業務ロジックとrepository呼び出しを検証する。

`post.go`では、次の観点をテストしている。

- 投稿内容の前後の空白を除去してrepositoryへ渡す
- 1文字未満の投稿を拒否する
- バリデーションエラー時にrepositoryを呼び出さない
- フォロー中ユーザーIDの取得をrepositoryへ委譲する

`follow.go`では、次の観点をテストしている。

- フォロー・フォロー解除をrepositoryへ委譲する
- 自分自身のフォロー・フォロー解除を拒否する
- バリデーションエラー時にrepositoryを呼び出さない

### カバレッジ結果

| ファイル  | 関数                   | カバレッジ |
| --------- | ---------------------- | ---------: |
| follow.go | NewFollowService       |     100.0% |
|           | Follow                 |     100.0% |
|           | Unfollow               |     100.0% |
|           | ListFolloweeIDs        |     100.0% |
|           | validateFollowRelation |     100.0% |
| post.go   | NewPostService         |     100.0% |
|           | Create                 |     100.0% |
| 合計      | -                      |      80.0% |

計測コマンド：

```bash
cd backend
go test ./internal/services -run '^(TestPostService|TestFollowService)' -coverprofile=/tmp/service-cover.out
go tool cover -func=/tmp/service-cover.out | grep -E 'post.go|follow.go'
```

## internal/repositories

Repositoryでは`sqlmock`を使用し、実際のPostgreSQLへ接続せずにSQLとDB結果のマッピングを検証する。

`post.go`では、次の観点をテストしている。

- `INSERT`へ投稿者IDと投稿内容を渡す
- DBから返された投稿データをモデルへマッピングする
- DBエラーをアプリケーションエラーへ分類する

`follow.go`では、次の観点をテストしている。

- フォロー登録SQLと引数を検証する
- フォロー解除SQLと引数を検証する
- フォロー中ユーザーIDを取得するSQLと結果マッピングを検証する
- DBエラーをアプリケーションエラーへ分類する

`timeline.go`では、次の観点をテストしている。

- `for_you`のタイムライン取得SQLと結果マッピングを検証する
- `following`のタイムライン取得SQLと引数を検証する
- Query・Scan・RowsのDBエラーをアプリケーションエラーへ分類する

`post_error.go`では、次の観点をテストしている。

- PostgreSQLのエラーコードをアプリケーションエラーへ分類する

### カバレッジ結果

| ファイル      | 関数                  | カバレッジ |
| ------------- | --------------------- | ---------: |
| follow.go     | NewFollowRepository   |     100.0% |
|               | Follow                |     100.0% |
|               | Unfollow              |     100.0% |
|               | ListFolloweeIDs       |     100.0% |
|               | execute               |     100.0% |
| post.go       | NewPostRepository     |     100.0% |
|               | Create                |     100.0% |
| post_error.go | classifyPostgresError |     100.0% |
| timeline.go   | NewTimelineRepository |     100.0% |
|               | List                  |     100.0% |
| 合計          | -                     |     100.0% |

計測コマンド：

```bash
cd backend
go test ./internal/repositories -coverprofile=/tmp/repository-cover.out
go tool cover -func=/tmp/repository-cover.out
```

## internal/apperrors

アプリケーションエラーでは、共通エラー型・HTTPレスポンス変換・ステータスコード分類を検証する。

`error.go`では、次の観点をテストしている。

- エラーメッセージを返す
- 元のエラーを保持する

`errorHandler.go`では、次の観点をテストしている。

- 想定外のエラーを共通エラーへ変換する
- エラー原因の詳細をレスポンスへ含めない
- エラー種別に応じたHTTPステータスコードを返す

`errorcode.go`では、次の観点をテストしている。

- エラーコード・メッセージ・元のエラーを共通エラーへまとめる

### カバレッジ結果

| ファイル        | 関数          | カバレッジ |
| --------------- | ------------- | ---------: |
| error.go        | Error         |     100.0% |
|                 | Unwrap        |     100.0% |
| errorHandler.go | ErrorHandler  |     100.0% |
|                 | statusCodeFor |     100.0% |
| errorcode.go    | Wrap          |     100.0% |
| 合計            | -             |     100.0% |

計測コマンド：

```bash
cd backend
go test ./internal/apperrors -coverprofile=/tmp/apperrors-cover.out
go tool cover -func=/tmp/apperrors-cover.out
```

## cmd/db-user

DBユーザー管理コマンドでは、Secrets Managerの認証情報、DBロール作成、app_user・migration_userの権限付与、既存テーブルの所有権移譲を検証する。

`main`関数では、次の観点をテストしている。

- CLI処理の終了コードをプロセス終了へ渡す

`runCLI`関数では、次の観点をテストしている。

- CLI引数を読み取り、正常終了・異常終了の終了コードと出力を返す
- `--ca-file`が未指定の場合はエラーにする

`run`関数では、次の観点をテストしている。

- app_user用Secretとmigration_user用Secretが同じ場合は拒否する
- 管理者接続失敗時は処理を止める
- トランザクション開始失敗時は処理を止める
- セットアップロック取得失敗時は処理を止める
- app_userとmigration_userを順にセットアップする
- commit失敗時は再実行を促すエラーにする

`setupUser`関数では、次の観点をテストしている。

- Secretの有無に応じてDBユーザーをセットアップする
- 既存DBユーザーの認証情報が不正な場合は拒否する
- 既存DBユーザーの認証情報で接続確認する
- Secret取得・保存に失敗した場合はエラーにする

`databaseUserPassword`関数では、次の観点をテストしている。

- 既存SecretからDBユーザーのパスワードを読み取る
- Secret内のユーザー名が期待するDBロールと異なる場合は拒否する
- Secret内のパスワードが空の場合は拒否する
- 既存Secretがない場合は新しいパスワードを生成する
- パスワード生成に失敗した場合はエラーにする

`createRole`関数では、次の観点をテストしている。

- 対応していないDBロール名を拒否する
- DBロール存在確認の失敗をエラーにする
- Secretがある既存DBロールは維持する
- Secretがない既存DBロールは拒否する

`configureAppRole`関数では、次の観点をテストしている。

- app_userを作成し、DB接続・schema利用・既存アプリテーブルへの読み書き権限を付与する
- app_userが存在するがSecretがない場合は拒否する
- マイグレーション前でテーブルが存在しない場合はテーブル権限付与をスキップする
- DBエラーの詳細をそのまま外へ出さない
- DB接続・schema利用・テーブル権限付与の各失敗を区別する

`configureMigrationRole`関数では、次の観点をテストしている。

- migration_userを作成し、DB接続・schema利用・schema作成権限を付与する
- migration_userが既に存在する場合は既存ロールを維持する
- migration_userが存在するがSecretがない場合は拒否する
- migration_user用Secretで既存パスワードを再利用する
- DBエラーの詳細をそのまま外へ出さない
- DB接続・schema利用・管理者membership・拡張準備の各失敗を区別する

`transferMigrationTables`関数では、次の観点をテストしている。

- アプリテーブルとschema_migrationsの所有者をmigration_userへ変更する
- 存在しないテーブルは所有者変更をスキップする
- DBエラーの詳細をそのまま外へ出さない

### カバレッジ結果

| ファイル | 関数                    | カバレッジ |
| -------- | ----------------------- | ---------: |
| main.go  | main                    |     100.0% |
|          | runCLI                  |      95.0% |
|          | run                     |     100.0% |
|          | setupUser               |      96.9% |
|          | databaseUserPassword    |     100.0% |
|          | createRole              |     100.0% |
|          | configureAppRole        |     100.0% |
|          | configureMigrationRole  |     100.0% |
|          | transferMigrationTables |     100.0% |
| 合計     | -                       |      98.4% |

計測コマンド：

```bash
cd backend
go test ./cmd/db-user -coverprofile=/tmp/db-user-cover.out
go tool cover -func=/tmp/db-user-cover.out
```

## internal/dbadmin

DB管理処理では、AWS SDK・Secrets Manager・RDS接続設定・SSMポートフォワード時の接続先差し替えを検証する。

`admin.go`では、次の観点をテストしている。

- PostgreSQL接続URLを組み立てる
- 明示した管理者Secretを使ってRDS接続情報を組み立てる
- 管理者Secretが未指定の場合はエラーにする
- RDS CAファイルが読めない場合はエラーにする
- RDS接続情報や管理者Secretの取得に失敗した場合はエラーにする
- 管理者Secretのユーザー名が`dbadmin`でない場合は拒否する
- 管理者Secretのパスワードが空の場合は拒否する
- 不正なローカル転送先を指定した場合はDB接続初期化エラーにする

`aws.go`では、次の観点をテストしている。

- AWS SDK設定の読み込み失敗を共通エラーにする
- RDSインスタンスから接続先ホスト名とポートを取得する
- RDSインスタンスのendpointが取得できない場合は共通エラーにする
- Secrets Managerに`AWSCURRENT`のバージョンが存在するか判定する
- Secretの文字列を取得する
- Secretの文字列を保存する
- AWS SDKのエラー詳細をそのまま外へ出さない

`tunnel.go`では、次の観点をテストしている。

- 実DBへ接続せずにDB接続ハンドルを作成する
- 不正なDB接続URLを拒否する
- SSMポートフォワード時もTLS検証先としてRDSホスト名を維持する
- 接続先の差し替えはloopback IPと有効なportだけ許可する
- portが無い・数値でないローカル転送先を拒否する
- ローカル転送先が未指定なら通常のRDS接続設定を維持する

### カバレッジ結果

| ファイル  | 関数                       | カバレッジ |
| --------- | -------------------------- | ---------: |
| admin.go  | ConnectionURL              |     100.0% |
|           | OpenAdministrator          |      83.3% |
| aws.go    | awsConfig                  |     100.0% |
|           | GetRDSEndpoint             |      92.3% |
|           | CurrentSecretVersionExists |      90.9% |
|           | GetSecretString            |      85.7% |
|           | PutSecretString            |      85.7% |
| tunnel.go | OpenDatabase               |     100.0% |
|           | configureLocalForward      |      93.3% |
| 合計      | -                          |      90.1% |

計測コマンド：

```bash
cd backend
go test ./internal/dbadmin -coverprofile=/tmp/dbadmin-cover.out
go tool cover -func=/tmp/dbadmin-cover.out
```

## テストの実行

バックエンド全体のテストは、`backend`ディレクトリで次を実行する。

```bash
go test ./...
```

Controller以外の層についても、依存先を差し替えて各層の責務を個別に検証する。
