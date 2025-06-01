package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadCredentials(t *testing.T) {
	// テストデータの作成
	validConfig := Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
	}

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "正常系：有効なクレデンシャル",
			config:      &validConfig,
			expectError: false,
		},
		{
			name: "異常系：ClientID無し",
			config: &Config{
				ClientSecret: "test-client-secret",
				RedirectURL:  "http://localhost:8080/callback",
			},
			expectError: true,
		},
		{
			name: "異常系：ClientSecret無し",
			config: &Config{
				ClientID:    "test-client-id",
				RedirectURL: "http://localhost:8080/callback",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用の一時ファイルを作成
			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, "credentials.json")

			data, err := json.Marshal(tt.config)
			require.NoError(t, err)

			err = os.WriteFile(filePath, data, 0644)
			require.NoError(t, err)

			// テスト実行
			config, err := LoadCredentials(filePath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, config)
			assert.Equal(t, tt.config.ClientID, config.ClientID)
			assert.Equal(t, tt.config.ClientSecret, config.ClientSecret)
			assert.Equal(t, tt.config.RedirectURL, config.RedirectURL)
		})
	}
}

func TestGetCredentialsPath(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    bool
		createFile  bool
		expectError bool
	}{
		{
			name:        "正常系：環境変数とファイルが存在",
			setupEnv:    true,
			createFile:  true,
			expectError: false,
		},
		{
			name:        "異常系：環境変数無し",
			setupEnv:    false,
			createFile:  true,
			expectError: true,
		},
		{
			name:        "異常系：ファイル無し",
			setupEnv:    true,
			createFile:  false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数をクリア
			t.Setenv(EnvCredentialsPath, "")

			// 環境変数とファイルの設定
			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, "credentials.json")

			if tt.setupEnv {
				t.Setenv(EnvCredentialsPath, filePath)
			}

			if tt.createFile {
				err := os.WriteFile(filePath, []byte("{}"), 0644)
				require.NoError(t, err)
			}

			// テスト実行
			path, err := GetCredentialsPath()

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, path)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, filePath, path)
		})
	}
}

func TestConfigFromJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name: "正常系：有効な設定",
			input: `{
				"client_id": "test-client-id",
				"client_secret": "test-client-secret",
				"redirect_url": "http://localhost:8080/callback"
			}`,
			expectError: false,
		},
		{
			name: "異常系：無効なJSON",
			input: `{
				"client_id": "test-client-id",
				"client_secret": "test-client-secret",
				"redirect_url": "http://localhost:8080/callback"
			`,
			expectError: true,
		},
		{
			name: "異常系：必須フィールド無し",
			input: `{
				"redirect_url": "http://localhost:8080/callback"
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ConfigFromJSON([]byte(tt.input))

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, config)

			// OAuth2 設定の検証
			assert.Equal(t, Scopes, config.Scopes)
			assert.Equal(t, "https://accounts.google.com/o/oauth2/auth", config.Endpoint.AuthURL)
			assert.Equal(t, "https://oauth2.googleapis.com/token", config.Endpoint.TokenURL)
		})
	}
}
