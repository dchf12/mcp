package config

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/dch/mcp-google-calendar/internal/infrastructure/repository"
	"github.com/dch/mcp-google-calendar/pkg/errors"
	"golang.org/x/oauth2"
)

const (
	// LocalServerPort はOAuthコールバック用のローカルサーバーのデフォルトポート番号です
	// このポートでローカルサーバーが起動し、OAuth2.0認証後のコールバックを受け付けます
	LocalServerPort = 8080
)

// AuthFlow はOAuth認証フローを実行し、トークンを取得・保存します
//
// 以下の順序で認証を実行します：
// 1. 既存の有効なトークンがあれば、それを返します
// 2. ブラウザを開いてユーザーに認証を要求します
// 3. ローカルサーバーで認証コードを受け取ります
// 4. 認証コードをトークンと交換します
// 5. トークンを保存して返します
//
// 引数:
//   - ctx: コンテキスト（タイムアウトやキャンセルに使用）
//   - config: OAuth設定
//
// 戻り値:
//   - *oauth2.Token: 取得したアクセストークン
//   - error: 認証エラー、ネットワークエラー、またはトークン保存エラー
func AuthFlow(ctx context.Context, config *Config) (*oauth2.Token, error) {
	oauthConfig := config.NewOAuthConfig()

	// 既存のトークンを確認
	if token, err := LoadToken(); err == nil && token.Valid() {
		return token, nil
	}

	authURL := GetAuthURL(oauthConfig)

	if err := openBrowser(authURL); err != nil {
		// ブラウザを開けない場合はURLを表示するだけなので、エラーとしては扱わない
		slog.Info("ブラウザでURLを開けません。手動でアクセスしてください", "url", authURL)
	}

	// 認証コードを受け取るためのローカルサーバーを起動
	code, err := startLocalServer(ctx)
	if err != nil {
		return nil, errors.NewOAuthError("authorization", "failed to get authorization code", err)
	}

	// 認証コードをトークンと交換
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, errors.NewOAuthError("token_exchange", "failed to exchange code for token", err)
	}

	// トークンを保存
	if err := SaveToken(token); err != nil {
		return nil, errors.NewOAuthError("token_save", "failed to save token", err)
	}

	return token, nil
}

// GetAuthURL は認証用URLを生成します
//
// オフラインアクセスモードでURLを生成し、リフレッシュトークンの取得を可能にします
//
// 引数:
//   - config: OAuth設定
//
// 戻り値:
//   - string: 認証用URL（ユーザーがブラウザでアクセスするURL）
func GetAuthURL(config *oauth2.Config) string {
	return config.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

// SaveToken はOAuth2トークンをファイルに保存します
//
// ~/.config/gcal_mcp/token.json にトークンを保存します
// ファイルは600パーミッションで作成され、セキュアに保存されます
//
// 引数:
//   - token: 保存するOAuthトークン
//
// 戻り値:
//   - error: ファイル作成エラーまたは保存エラー
func SaveToken(token *oauth2.Token) error {
	repo, err := repository.DefaultTokenFileRepo()
	if err != nil {
		return errors.NewOAuthError("token_repository", "failed to create token repository", err)
	}
	return repo.Save(token)
}

// LoadToken は保存されたOAuth2トークンを読み込みます
//
// ~/.config/gcal_mcp/token.json からトークンを読み込みます
//
// 戻り値:
//   - *oauth2.Token: 読み込んだトークン
//   - error: ファイル読み込みエラーまたはJSONデコードエラー
func LoadToken() (*oauth2.Token, error) {
	repo, err := repository.DefaultTokenFileRepo()
	if err != nil {
		return nil, errors.NewOAuthError("token_repository", "failed to create token repository", err)
	}

	token, err := repo.Load()
	if err != nil {
		return nil, errors.NewOAuthError("token_load", "failed to load token", err)
	}

	return token, nil
}

// startLocalServer は認証コードを受け取るためのローカルサーバーを起動します
func startLocalServer(ctx context.Context) (string, error) {
	port := LocalServerPort

	codeChan := make(chan string)
	errChan := make(chan error)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("認証コードが見つかりません")
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("認証が完了しました。このウィンドウを閉じてください。"))
		codeChan <- code
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	// サーバーをゴルーチンで起動
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// コンテキストのキャンセルを監視
	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	select {
	case code := <-codeChan:
		server.Shutdown(context.Background())
		return code, nil
	case err := <-errChan:
		server.Shutdown(context.Background())
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// openBrowser はデフォルトブラウザでURLを開きます
func openBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("cmd", "/c", "start", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err
}
