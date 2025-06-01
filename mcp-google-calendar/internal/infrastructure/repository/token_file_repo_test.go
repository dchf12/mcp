package repository

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestTokenFileRepo(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewTokenFileRepo(tempDir)
	require.NoError(t, err)

	t.Run("Save and Load token", func(t *testing.T) {
		token := &oauth2.Token{
			AccessToken:  "test_access_token",
			TokenType:    "Bearer",
			RefreshToken: "test_refresh_token",
			Expiry:       time.Now().Add(time.Hour),
		}

		err := repo.Save(token)
		if err != nil {
			t.Fatalf("failed to save token: %v", err)
		}

		loaded, err := repo.Load()
		if err != nil {
			t.Fatalf("failed to load token: %v", err)
		}

		if token.AccessToken != loaded.AccessToken {
			t.Errorf("got access token %q, want %q", loaded.AccessToken, token.AccessToken)
		}
		if token.RefreshToken != loaded.RefreshToken {
			t.Errorf("got refresh token %q, want %q", loaded.RefreshToken, token.RefreshToken)
		}
	})

	t.Run("File permissions", func(t *testing.T) {
		token := &oauth2.Token{
			AccessToken: "test_access_token",
		}

		err := repo.Save(token)
		if err != nil {
			t.Fatalf("failed to save token: %v", err)
		}

		info, err := os.Stat(filepath.Join(tempDir, "token.enc"))
		if err != nil {
			t.Fatalf("failed to stat token file: %v", err)
		}

		if info.Mode().Perm() != 0600 {
			t.Errorf("got file mode %v, want %v", info.Mode().Perm(), 0600)
		}
	})

	t.Run("Load non-existent file", func(t *testing.T) {
		repo, err := NewTokenFileRepo(t.TempDir())
		require.NoError(t, err)
		_, err = repo.Load()
		if err == nil {
			t.Error("expected error when loading non-existent file")
		}
	})

	t.Run("Load invalid JSON", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "token.enc"), []byte("invalid data"), 0600)
		if err != nil {
			t.Fatalf("failed to write invalid json: %v", err)
		}

		repo, err := NewTokenFileRepo(dir)
		require.NoError(t, err)
		_, err = repo.Load()
		if err == nil {
			t.Error("expected error when loading invalid JSON")
		}
	})
}
