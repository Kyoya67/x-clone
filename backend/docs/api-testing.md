# APIテスト

## Controllerテスト

Controllerの依存先はFakeへ差し替え、HTTPリクエストに対するレスポンスを検証する。

投稿Controllerでは、次の観点をテストしている。

- 正常なリクエストで`201 Created`と投稿データを返す
- Serviceが返した内部エラーを`500 Internal Server Error`へ変換する
- DBなどの内部エラーの詳細をレスポンスへ含めない

フォローControllerでは、次の観点をテストしている。

- フォロー・フォロー解除で`204 No Content`を返す
- URLの`userId`をServiceへ渡す
- UUID形式でない`userId`を拒否する
- Serviceの入力エラーを`400 Bad Request`へ変換する
- Serviceの内部エラーを`500 Internal Server Error`へ変換し、詳細をレスポンスへ含めない
- ログインユーザーがフォローしているユーザーIDを`200 OK`で返す

### カバレッジ結果

| ファイル    | 関数                    | カバレッジ |
| ----------- | ----------------------- | ---------: |
| `follow.go` | `NewFollowController`   |     100.0% |
|             | `Follow`                |     100.0% |
|             | `Unfollow`              |     100.0% |
|             | `ListFollowing`         |      66.7% |
|             | `handleFollowAction`    |     100.0% |
|             | `followeeIDFromRequest` |     100.0% |
| `post.go`   | `NewPostController`     |     100.0% |
|             | `Create`                |     100.0% |

計測コマンド：

```bash
cd backend
go test ./internal/controllers -run '^(TestPostController|TestFollowController)' -coverprofile=/tmp/controller-cover.out
go tool cover -func=/tmp/controller-cover.out | grep -E 'post.go|follow.go'
```

## Serviceテスト

Serviceの依存先はFakeへ差し替え、業務ロジックとrepository呼び出しを検証する。

投稿Serviceでは、次の観点をテストしている。

- 投稿内容の前後の空白を除去してrepositoryへ渡す
- 1文字未満の投稿を拒否する
- バリデーションエラー時にrepositoryを呼び出さない
- フォロー中ユーザーIDの取得をrepositoryへ委譲する

フォローServiceでは、次の観点をテストしている。

- フォロー・フォロー解除をrepositoryへ委譲する
- 自分自身のフォロー・フォロー解除を拒否する
- バリデーションエラー時にrepositoryを呼び出さない

### カバレッジ結果

| ファイル    | 関数                     | カバレッジ |
| ----------- | ------------------------ | ---------: |
| `follow.go` | `NewFollowService`       |     100.0% |
|             | `Follow`                 |     100.0% |
|             | `Unfollow`               |     100.0% |
|             | `ListFolloweeIDs`        |     100.0% |
|             | `validateFollowRelation` |     100.0% |
| `post.go`   | `NewPostService`         |     100.0% |
|             | `Create`                 |     100.0% |

計測コマンド：

```bash
cd backend
go test ./internal/services -run '^(TestPostService|TestFollowService)' -coverprofile=/tmp/service-cover.out
go tool cover -func=/tmp/service-cover.out | grep -E 'post.go|follow.go'
```

## Repositoryテスト

Repositoryでは`sqlmock`を使用し、実際のPostgreSQLへ接続せずにSQLとDB結果のマッピングを検証する。

投稿Repositoryでは、次の観点をテストしている。

- `INSERT`へ投稿者IDと投稿内容を渡す
- DBから返された投稿データをモデルへマッピングする
- DBエラーをアプリケーションエラーへ分類する

フォローRepositoryでは、次の観点をテストしている。

- フォロー登録SQLと引数を検証する
- フォロー解除SQLと引数を検証する
- フォロー中ユーザーIDを取得するSQLと結果マッピングを検証する
- DBエラーをアプリケーションエラーへ分類する

### カバレッジ結果

| ファイル        | 関数                    | カバレッジ |
| --------------- | ----------------------- | ---------: |
| `follow.go`     | `NewFollowRepository`   |     100.0% |
|                 | `Follow`                |     100.0% |
|                 | `Unfollow`              |     100.0% |
|                 | `ListFolloweeIDs`       |      76.9% |
|                 | `execute`               |     100.0% |
| `post.go`       | `NewPostRepository`     |     100.0% |
|                 | `Create`                |     100.0% |
| `post_error.go` | `classifyPostgresError` |     100.0% |

計測コマンド：

```bash
cd backend
go test ./internal/repositories -run '^(TestPostRepository|TestFollowRepository|TestClassifyPostgresError)' -coverprofile=/tmp/repository-cover.out
go tool cover -func=/tmp/repository-cover.out | grep -E 'post|follow'
```

## DB管理処理テスト

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

| ファイル    | 関数                         | カバレッジ |
| ----------- | ---------------------------- | ---------: |
| `admin.go`  | `ConnectionURL`              |     100.0% |
|             | `OpenAdministrator`          |      83.3% |
| `aws.go`    | `awsConfig`                  |     100.0% |
|             | `GetRDSEndpoint`             |      92.3% |
|             | `CurrentSecretVersionExists` |      90.9% |
|             | `GetSecretString`            |      85.7% |
|             | `PutSecretString`            |      85.7% |
| `tunnel.go` | `OpenDatabase`               |     100.0% |
|             | `configureLocalForward`      |      93.3% |
| 合計        | -                            |      90.1% |

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
