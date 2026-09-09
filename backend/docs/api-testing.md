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

### カバレッジ結果

constructorはテスト対象に含めない。constructorは依存関係を構造体へ設定するだけで、独立した振る舞いを持たないためである。

| 対象           | 関数     | カバレッジ |
| -------------- | -------- | ---------: |
| PostController | `Create` |     100.0% |
| FollowController | `Follow` |     100.0% |
| FollowController | `Unfollow` |     100.0% |

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

フォローServiceでは、次の観点をテストしている。

- フォロー・フォロー解除をrepositoryへ委譲する
- 自分自身のフォロー・フォロー解除を拒否する
- バリデーションエラー時にrepositoryを呼び出さない

### カバレッジ結果

| 対象        | 関数     | カバレッジ |
| ----------- | -------- | ---------: |
| PostService | `Create` |     100.0% |
| FollowService | `Follow` |     100.0% |
| FollowService | `Unfollow` |     100.0% |

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
- DBエラーをアプリケーションエラーへ分類する

### カバレッジ結果

constructorはテスト対象に含めない。

| 対象                            | 関数                    | カバレッジ |
| ------------------------------- | ----------------------- | ---------: |
| PostRepository                  | `Create`                |     100.0% |
| PostgreSQL error classification | `classifyPostgresError` |     100.0% |
| FollowRepository                | `Follow`                |     100.0% |
| FollowRepository                | `Unfollow`              |     100.0% |

計測コマンド：

```bash
cd backend
go test ./internal/repositories -run '^(TestPostRepository|TestFollowRepository|TestClassifyPostgresError)' -coverprofile=/tmp/repository-cover.out
go tool cover -func=/tmp/repository-cover.out | grep -E 'post|follow'
```

## テストの実行

バックエンド全体のテストは、`backend`ディレクトリで次を実行する。

```bash
go test ./...
```

Controller以外の層についても、依存先を差し替えて各層の責務を個別に検証する。
