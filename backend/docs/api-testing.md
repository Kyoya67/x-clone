# APIテスト

## Controllerテスト

Controllerの依存先はFakeへ差し替え、HTTPリクエストに対するレスポンスを検証する。

投稿Controllerでは、次の観点をテストしている。

- 正常なリクエストで`201 Created`と投稿データを返す
- Serviceが返した内部エラーを`500 Internal Server Error`へ変換する
- DBなどの内部エラーの詳細をレスポンスへ含めない

### カバレッジ結果

constructorはテスト対象に含めない。constructorは依存関係を構造体へ設定するだけで、独立した振る舞いを持たないためである。

| 対象 | 関数 | カバレッジ |
|---|---|---:|
| PostController | `Create` | 100.0% |

計測コマンド：

`````bash
cd backend
go test ./internal/controllers -run '^TestPostController' -coverprofile=/tmp/post-controller-cover.out
go tool cover -func=/tmp/post-controller-cover.out | grep 'post.go'
`````

## Serviceテスト

Serviceの依存先はFakeへ差し替え、業務ロジックとrepository呼び出しを検証する。

投稿Serviceでは、次の観点をテストしている。

- 投稿内容の前後の空白を除去してrepositoryへ渡す
- 1文字未満の投稿を拒否する
- バリデーションエラー時にrepositoryを呼び出さない

### カバレッジ結果

| 対象 | 関数 | カバレッジ |
|---|---|---:|
| PostService | `Create` | 100.0% |

計測コマンド：

`````bash
cd backend
go test ./internal/services -run '^TestPostService' -coverprofile=/tmp/post-service-cover.out
go tool cover -func=/tmp/post-service-cover.out | grep 'post.go'
`````

## Repositoryテスト

Repositoryでは`sqlmock`を使用し、実際のPostgreSQLへ接続せずにSQLとDB結果のマッピングを検証する。

投稿Repositoryでは、次の観点をテストしている。

- `INSERT`へ投稿者IDと投稿内容を渡す
- DBから返された投稿データをモデルへマッピングする
- DBエラーをアプリケーションエラーへ分類する

### カバレッジ結果

constructorはテスト対象に含めない。

| 対象 | 関数 | カバレッジ |
|---|---|---:|
| PostRepository | `Create` | 80.0% |
| PostgreSQL error classification | `classifyPostgresError` | 100.0% |

計測コマンド：

`````bash
cd backend
go test ./internal/repositories -run '^TestPostRepository|^TestClassifyPostgresError' -coverprofile=/tmp/post-repository-cover.out
go tool cover -func=/tmp/post-repository-cover.out | grep 'post'
`````

## テストの実行

バックエンド全体のテストは、`backend`ディレクトリで次を実行する。

`````bash
go test ./...
`````

Controller以外の層についても、依存先を差し替えて各層の責務を個別に検証する。
