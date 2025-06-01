// Package config は Google Calendar API の認証設定を管理します
package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/dch/mcp-google-calendar/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	// EnvCredentialsPath は認証情報ファイルへのパスを指定する環境変数名です
	EnvCredentialsPath = "GCAL_CREDENTIALS_PATH"
)

// Google Calendar APIのスコープ定義
var (
	// Scopes はGoogle Calendar APIの読み取りと書き込みに必要なスコープを定義します
	Scopes = []string{
		"https://www.googleapis.com/auth/calendar.readonly",
		"https://www.googleapis.com/auth/calendar.events",
	}
)

// Config はOAuth2の設定情報を保持します
// Google Cloud Consoleから取得したクライアント認証情報をJSONファイルから読み込むために使用します
type Config struct {
	ClientID     string `json:"client_id"`     // OAuth2.0クライアントID
	ClientSecret string `json:"client_secret"` // OAuth2.0クライアントシークレット
	RedirectURL  string `json:"redirect_url"`  // OAuth2.0コールバックURL

	// テスト用のカスタムエンドポイント（オプション）
	AuthURL  string // カスタム認証エンドポイント
	TokenURL string // カスタムトークンエンドポイント
}

// LoadCredentials は指定されたファイルパスからOAuthクレデンシャルを読み込みます
//
// filePath には credentials.json ファイルへの絶対パスを指定します
// このファイルは Google Cloud Console からダウンロードできます
//
// 戻り値:
//   - *Config: 読み込まれた認証設定
//   - error: 読み込み時のエラー（ファイルが存在しない、JSONが不正、必須フィールドが欠落など）
func LoadCredentials(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.NewConfigError("credentials", "failed to read file", err)
	}

	var conf Config
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, errors.NewConfigError("credentials", "invalid JSON format", err)
	}

	if err := conf.validate(); err != nil {
		return nil, err
	}

	return &conf, nil
}

// GetCredentialsPath は環境変数から認証情報ファイルのパスを取得します
//
// 環境変数 GCAL_CREDENTIALS_PATH で指定されたパスを返します
// ファイルの存在確認も行います
//
// 戻り値:
//   - string: 認証情報ファイルへの絶対パス
//   - error: 環境変数未設定、ファイルが存在しない、またはアクセス権限がない場合のエラー
func GetCredentialsPath() (string, error) {
	path := os.Getenv(EnvCredentialsPath)
	if path == "" {
		return "", errors.NewConfigError(EnvCredentialsPath, "environment variable not set", nil)
	}

	// 絶対パスに変換
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", errors.NewConfigError("credentials", "failed to resolve absolute path", err)
	}

	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			return "", errors.NewConfigError("credentials", "file does not exist", err)
		}
		return "", errors.NewConfigError("credentials", "failed to access file", err)
	}

	return absPath, nil
}

// validate は設定値のバリデーションを行います
//
// 必須フィールド（ClientID、ClientSecret、RedirectURL）が設定されているか確認します
func (c *Config) validate() error {
	if c.ClientID == "" {
		return errors.NewValidationError("client_id", "required field is empty", nil)
	}
	if c.ClientSecret == "" {
		return errors.NewValidationError("client_secret", "required field is empty", nil)
	}
	if c.RedirectURL == "" {
		return errors.NewValidationError("redirect_url", "required field is empty", nil)
	}
	return nil
}

// NewOAuthConfig はOAuth2設定を作成します
//
// Google Calendar APIのエンドポイントとスコープを含むoauth2.Configを生成します
// AuthURLとTokenURLが指定されている場合は、カスタムエンドポイントを使用します
//
// 戻り値:
//   - *oauth2.Config: OAuth2クライアントの設定
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
//
// credentials.jsonの内容から直接oauth2.Configを生成します
//
// 引数:
//   - jsonData: credentials.jsonファイルの内容
//
// 戻り値:
//   - *oauth2.Config: OAuth2クライアントの設定
//   - error: JSONのパースエラーまたはバリデーションエラー
func ConfigFromJSON(jsonData []byte) (*oauth2.Config, error) {
	var conf Config
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return nil, errors.NewConfigError("credentials", "invalid JSON format", err)
	}

	if err := conf.validate(); err != nil {
		return nil, err
	}

	return conf.NewOAuthConfig(), nil
}
