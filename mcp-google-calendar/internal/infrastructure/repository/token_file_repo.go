package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

// TokenFileRepo は OAuth2 トークンをファイルに保存・読み込みするリポジトリです
type TokenFileRepo struct {
	tokenPath string
}

// NewTokenFileRepo は TokenFileRepo の新しいインスタンスを作成します
func NewTokenFileRepo(configDir string) *TokenFileRepo {
	return &TokenFileRepo{
		tokenPath: filepath.Join(configDir, "token.json"),
	}
}

// DefaultTokenFileRepo は ~/.config/gcal_mcp/token.json をデフォルトのパスとして使用する
// TokenFileRepo のインスタンスを作成します
func DefaultTokenFileRepo() (*TokenFileRepo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".config", "gcal_mcp")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, err
	}

	return NewTokenFileRepo(configDir), nil
}

// Save は OAuth2 トークンをファイルに保存します
func (r *TokenFileRepo) Save(token *oauth2.Token) error {
	if token == nil {
		return errors.New("token cannot be nil")
	}

	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return os.WriteFile(r.tokenPath, data, 0600)
}

// Load は保存された OAuth2 トークンをファイルから読み込みます
func (r *TokenFileRepo) Load() (*oauth2.Token, error) {
	data, err := os.ReadFile(r.tokenPath)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}
