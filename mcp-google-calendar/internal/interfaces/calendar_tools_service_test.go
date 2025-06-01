package interfaces

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/dch/mcp-google-calendar/pkg/config"
)

func TestRegisterCalendarTools(t *testing.T) {
	tests := []struct {
		name        string
		server      *server.MCPServer
		conf        *config.Config
		token       *oauth2.Token
		expectError bool
		errorMsg    string
	}{
		{
			name:   "正常系: 全ての設定が有効",
			server: server.NewMCPServer("test-server", "1.0.0"),
			conf: &config.Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				RedirectURL:  "http://localhost:8080/callback",
			},
			token: &oauth2.Token{
				AccessToken:  "test-access-token",
				TokenType:    "Bearer",
				RefreshToken: "test-refresh-token",
				Expiry:       time.Now().Add(time.Hour),
			},
			expectError: false,
		},
		{
			name:        "異常系: サーバーがnil",
			server:      nil,
			conf:        &config.Config{},
			token:       &oauth2.Token{},
			expectError: true,
			errorMsg:    "サーバーインスタンスが指定されていません",
		},
		{
			name:        "異常系: 設定がnil",
			server:      server.NewMCPServer("test-server", "1.0.0"),
			conf:        nil,
			token:       &oauth2.Token{},
			expectError: true,
			errorMsg:    "OAuth設定が指定されていません",
		},
		{
			name:        "異常系: トークンがnil",
			server:      server.NewMCPServer("test-server", "1.0.0"),
			conf:        &config.Config{},
			token:       nil,
			expectError: true,
			errorMsg:    "認証トークンが指定されていません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト環境の準備
			tempDir := t.TempDir()
			t.Setenv("HOME", tempDir)
			configDir := filepath.Join(tempDir, ".config", "gcal_mcp")
			if err := os.MkdirAll(configDir, 0700); err != nil {
				t.Fatal(err)
			}

			// ツールの登録
			err := RegisterCalendarTools(tt.server, tt.conf, tt.token)

			// 結果の検証
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, tt.server)
		})
	}
}
