Package config はGoogle Calendar APIのOAuth2.0認証を管理するためのパッケージです。

主な機能：

1. 認証情報の管理
  - credentials.jsonファイルからのOAuth設定の読み込み
  - 環境変数GCAL_CREDENTIALS_PATHによる設定ファイルの指定

2. OAuth認証フロー
  - デスクトップアプリケーション向けのOAuth2.0認証フローの実装
  - ローカルサーバーを使用したコールバック処理
  - ブラウザ自動起動によるユーザーフレンドリーな認証

3. トークン管理
  - 安全なトークンの保存（~/.config/gcal_mcp/token.json）
  - トークンの自動リフレッシュ
  - 保存済みトークンの再利用

使用例：
```go
	// 認証情報の読み込み
	credPath, err := config.GetCredentialsPath()
	if err != nil {
		log.Fatal(err)
	}
	conf, err := config.LoadCredentials(credPath)
	if err != nil {
		log.Fatal(err)
	}

	// OAuth認証フローの実行
	ctx := context.Background()
	token, err := config.AuthFlow(ctx, conf)
	if err != nil {
		log.Fatal(err)
	}

	// トークンを使用してクライアントを作成
	oauthConfig := conf.NewOAuthConfig()
	client := oauthConfig.Client(ctx, token)
```
パッケージの設定：

1. 環境変数:
  - GCAL_CREDENTIALS_PATH: credentials.jsonファイルへのパス

2. ファイル:
  - credentials.json: OAuth2.0クライアント認証情報
  - ~/.config/gcal_mcp/token.json: 保存されたアクセストークン

セキュリティ：

- トークンファイルはパーミッション600で保存
- ローカルサーバーは一時的に起動し、認証完了後に自動終了
- トークンの自動リフレッシュにより継続的なアクセスを保証
