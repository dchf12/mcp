package config

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGetAuthURL(t *testing.T) {

	config := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://example.com/auth",
			TokenURL: "https://example.com/token",
		},
	}

	authURL := GetAuthURL(config)

	assert.Contains(t, authURL, "https://example.com/auth")
	assert.Contains(t, authURL, "client_id=test-client-id")
	assert.Contains(t, authURL, "redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback")
	assert.Contains(t, authURL, "access_type=offline")
	assert.Contains(t, authURL, "state=state")

	expectedScope := url.QueryEscape(strings.Join(Scopes, " "))
	assert.Contains(t, authURL, "scope="+expectedScope)
}

func TestAuthFlow(t *testing.T) {
	// より長いタイムアウトを設定
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// モックのOAuthサーバー設定
	mockOAuth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			// トークンエンドポイントの処理
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"access_token": "test-access-token",
				"token_type": "Bearer",
				"refresh_token": "test-refresh-token",
				"expiry": "2025-05-31T12:00:00Z"
			}`))
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
	}))
	defer mockOAuth.Close()

	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  fmt.Sprintf("http://localhost:%d/callback", 8080),
		AuthURL:      mockOAuth.URL + "/auth",
		TokenURL:     mockOAuth.URL + "/token",
	}

	// AuthFlow を別ゴルーチンで実行
	tokenCh := make(chan *oauth2.Token, 1)
	errCh := make(chan error, 1)
	go func() {
		tok, err := AuthFlow(ctx, config)
		if err != nil {
			errCh <- err
			return
		}
		tokenCh <- tok
	}()

	// ブラウザリダイレクトを模倣して認証コードを送信
	callbackURL := fmt.Sprintf("http://localhost:%d/callback?code=test-code&state=state", 8080)
	_, _ = http.Get(callbackURL)

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case token := <-tokenCh:
		require.NotNil(t, token)
		assert.Equal(t, "test-access-token", token.AccessToken)
		assert.Equal(t, "test-refresh-token", token.RefreshToken)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for AuthFlow to finish")
	}
}

func TestSaveAndLoadToken(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	token := &oauth2.Token{
		AccessToken:  "test_access_token",
		TokenType:    "Bearer",
		RefreshToken: "test_refresh_token",
		Expiry:       time.Now().Add(time.Hour),
	}

	// トークンを保存
	err := SaveToken(token)
	require.NoError(t, err)

	// トークンを読み込み
	loaded, err := LoadToken()
	require.NoError(t, err)
	require.NotNil(t, loaded)

	assert.Equal(t, token.AccessToken, loaded.AccessToken)
	assert.Equal(t, token.RefreshToken, loaded.RefreshToken)
	assert.Equal(t, token.TokenType, loaded.TokenType)
	assert.True(t, token.Expiry.Equal(loaded.Expiry))
}

func TestAuthFlowWithExistingToken(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	// 有効なトークンを事前に保存
	validToken := &oauth2.Token{
		AccessToken:  "existing_access_token",
		TokenType:    "Bearer",
		RefreshToken: "existing_refresh_token",
		Expiry:       time.Now().Add(time.Hour),
	}
	err := SaveToken(validToken)
	require.NoError(t, err)

	ctx := context.Background()
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
	}

	// AuthFlowを実行（既存のトークンが返されるはず）
	token, err := AuthFlow(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, token)

	assert.Equal(t, validToken.AccessToken, token.AccessToken)
	assert.Equal(t, validToken.RefreshToken, token.RefreshToken)
}
