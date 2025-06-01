package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"

	"github.com/dch/mcp-google-calendar/pkg/crypto"
)

// TokenFileRepo は OAuth2 トークンをファイルに保存・読み込みするリポジトリです
type TokenFileRepo struct {
	tokenPath  string
	encryption *crypto.TokenEncryption
}

// NewTokenFileRepo は TokenFileRepo の新しいインスタンスを作成します
func NewTokenFileRepo(configDir string) (*TokenFileRepo, error) {
	encryption, err := crypto.NewTokenEncryption()
	if err != nil {
		return nil, err
	}

	return &TokenFileRepo{
		tokenPath:  filepath.Join(configDir, "token.enc"),
		encryption: encryption,
	}, nil
}

// DefaultTokenFileRepo は ~/.config/gcal_mcp/token.enc をデフォルトのパスとして使用する
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

	return NewTokenFileRepo(configDir)
}

// Save は OAuth2 トークンを暗号化してファイルに保存します
func (r *TokenFileRepo) Save(token *oauth2.Token) error {
	if token == nil {
		return errors.New("token cannot be nil")
	}

	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	encryptedData, err := r.encryption.Encrypt(data)
	if err != nil {
		return err
	}

	return os.WriteFile(r.tokenPath, encryptedData, 0600)
}

// Load は保存された OAuth2 トークンを復号化してファイルから読み込みます
func (r *TokenFileRepo) Load() (*oauth2.Token, error) {
	encryptedData, err := os.ReadFile(r.tokenPath)
	if err != nil {
		return nil, err
	}

	data, err := r.encryption.Decrypt(encryptedData)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}
