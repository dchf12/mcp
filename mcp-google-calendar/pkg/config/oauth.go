package config

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"

	"golang.org/x/oauth2"
)

const (
	// LocalServerPort はOAuthコールバック用のローカルサーバーのデフォルトポート番号です
	LocalServerPort = 8080
)

// AuthFlow はOAuth認証フローを実行します
func AuthFlow(ctx context.Context, config *Config) (*oauth2.Token, error) {
	oauthConfig := config.NewOAuthConfig()

	authURL := GetAuthURL(oauthConfig)

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("ブラウザで以下のURLを開いてください:\n%s\n", authURL)
	}

	// 認証コードを受け取るためのローカルサーバーを起動
	code, err := startLocalServer(ctx)
	if err != nil {
		return nil, fmt.Errorf("認証コードの取得に失敗しました: %w", err)
	}

	// 認証コードをトークンと交換
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("トークンの取得に失敗しました: %w", err)
	}

	return token, nil
}

// GetAuthURL は認証用URLを生成します
func GetAuthURL(config *oauth2.Config) string {
	return config.AuthCodeURL("state", oauth2.AccessTypeOffline)
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
