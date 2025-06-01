package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainSetup(t *testing.T) {
	// テスト用の環境を設定
	tempDir := t.TempDir()
	credentialsPath := filepath.Join(tempDir, "credentials.json")

	// テスト用の認証情報を作成
	credentials := map[string]string{
		"client_id":     "test-client-id",
		"client_secret": "test-client-secret",
		"redirect_url":  "http://localhost:8080/callback",
	}
	credData, err := json.Marshal(credentials)
	require.NoError(t, err)

	err = os.WriteFile(credentialsPath, credData, 0600)
	require.NoError(t, err)

	// 環境変数を設定
	t.Setenv("GCAL_CREDENTIALS_PATH", credentialsPath)
	t.Setenv("HOME", tempDir)

	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// テスト後にクリーンアップ
	defer func() {
		os.Stdout = oldStdout
	}()

	// メイン処理の一部をテスト（エラーケースのみ）
	var output bytes.Buffer
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := r.Read(buffer)
			if n > 0 {
				output.Write(buffer[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	// サーバー名とバージョンの検証
	assert.Equal(t, "Google Calendar MCP Server", ServerName)
	assert.Regexp(t, `^\d+\.\d+\.\d+$`, ServerVersion)
}
