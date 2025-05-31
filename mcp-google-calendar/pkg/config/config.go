// Package config は Google Calendar API の認証設定を管理します
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	// GCAL_CREDENTIALS_PATH は認証情報ファイルへのパスを指定する環境変数名です
	EnvCredentialsPath = "GCAL_CREDENTIALS_PATH"
)

// Google Calendar APIのスコープ定義
var (
	// Required scopes for reading and writing calendar events
	Scopes = []string{
		"https://www.googleapis.com/auth/calendar.readonly",
		"https://www.googleapis.com/auth/calendar.events",
	}
)

// Config は OAuth2 の設定情報を保持します
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`

	AuthURL  string
	TokenURL string
}

// LoadCredentials は指定されたファイルパスから OAuth クレデンシャルを読み込みます
func LoadCredentials(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	var conf Config
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file: %w", err)
	}

	if err := conf.validate(); err != nil {
		return nil, err
	}

	return &conf, nil
}

// GetCredentialsPath は環境変数から認証情報ファイルのパスを取得します
func GetCredentialsPath() (string, error) {
	path := os.Getenv(EnvCredentialsPath)
	if path == "" {
		return "", fmt.Errorf("GCAL_CREDENTIALS environment variable is not set")
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("credentials file does not exist at %s", path)
		}
		return "", fmt.Errorf("failed to access credentials file: %w", err)
	}

	return path, nil
}

// validate は設定値のバリデーションを行います
func (c *Config) validate() error {
	if c.ClientID == "" {
		return fmt.Errorf("client_id is required in the credentials file")
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("client_secret is required in the credentials file")
	}
	if c.RedirectURL == "" {
		return fmt.Errorf("redirect_url is required in the credentials file")
	}
	return nil
}

// NewOAuthConfig は OAuth2 設定を作成します
func (c *Config) NewOAuthConfig() *oauth2.Config {
	endpoint := google.Endpoint
	if c.AuthURL != "" || c.TokenURL != "" {
		// 片方だけ指定された場合はデフォルトを補完
		if c.AuthURL == "" {
			c.AuthURL = endpoint.AuthURL
		}
		if c.TokenURL == "" {
			c.TokenURL = endpoint.TokenURL
		}
		endpoint = oauth2.Endpoint{AuthURL: c.AuthURL, TokenURL: c.TokenURL}
	}
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectURL,
		Scopes:       Scopes,
		Endpoint:     endpoint,
	}
}

// ConfigFromJSON はJSONデータからOAuth設定を作成します
func ConfigFromJSON(jsonData []byte) (*oauth2.Config, error) {
	var conf Config
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse OAuth configuration: %w", err)
	}

	if err := conf.validate(); err != nil {
		return nil, err
	}

	return conf.NewOAuthConfig(), nil
}
