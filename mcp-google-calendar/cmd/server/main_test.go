package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testServer struct {
	srv      *http.Server
	addr     string
	wg       sync.WaitGroup
	shutdown func(context.Context) error
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func setupTestServer(t *testing.T) *testServer {
	port, err := getFreePort()
	require.NoError(t, err)
	addr := fmt.Sprintf("localhost:%d", port)

	ts := &testServer{
		addr: addr,
	}

	// メトリクスサーバーの設定
	ts.srv = &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("test metrics"))
				return
			}
			if r.URL.Path == "/long-running" {
				time.Sleep(2 * time.Second)
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}),
	}

	// サーバーの起動
	ts.wg.Add(1)
	go func() {
		defer ts.wg.Done()
		if err := ts.srv.ListenAndServe(); err != http.ErrServerClosed {
			t.Logf("test server error: %v", err)
		}
	}()

	// サーバーが起動するまで待機
	ready := make(chan struct{})
	go func() {
		for {
			if _, err := http.Get(fmt.Sprintf("http://%s/metrics", addr)); err == nil {
				close(ready)
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case <-ready:
		// サーバーが起動した
	case <-time.After(5 * time.Second):
		t.Fatal("server failed to start within timeout")
	}

	// シャットダウン関数の設定
	ts.shutdown = func(ctx context.Context) error {
		return ts.srv.Shutdown(ctx)
	}

	return ts
}

func waitForServer(addr string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := http.Get(fmt.Sprintf("http://%s/metrics", addr)); err == nil {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func TestGracefulShutdown(t *testing.T) {
	t.Run("正常なシャットダウン", func(t *testing.T) {
		ts := setupTestServer(t)

		// メトリクスエンドポイントに疎通確認
		resp, err := http.Get(fmt.Sprintf("http://%s/metrics", ts.addr))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		// シャットダウン開始
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// 意図的な遅延を追加してシャットダウンを開始
		shutdownComplete := make(chan time.Duration, 1)
		go func() {
			start := time.Now()
			time.Sleep(100 * time.Millisecond)
			err := ts.shutdown(shutdownCtx)
			require.NoError(t, err)
			shutdownComplete <- time.Since(start)
		}()

		// シャットダウン完了を待機
		select {
		case duration := <-shutdownComplete:
			// サーバーが完全に停止したことを確認
			_, err = http.Get(fmt.Sprintf("http://%s/metrics", ts.addr))
			require.Error(t, err, "サーバーが停止していません")
			require.Greater(t, duration, 100*time.Millisecond, "シャットダウンが早すぎます")
			require.Less(t, duration, time.Second, "シャットダウンが遅すぎます")
		case <-time.After(2 * time.Second):
			t.Fatal("シャットダウンがタイムアウトしました")
		}
	})

	t.Run("タイムアウトによる強制シャットダウン", func(t *testing.T) {
		ts := setupTestServer(t)

		// 長時間のリクエストを送信
		go func() {
			_, _ = http.Get(fmt.Sprintf("http://%s/long-running", ts.addr))
		}()

		// リクエストが開始されるのを待機
		time.Sleep(100 * time.Millisecond)

		// シャットダウン開始（短いタイムアウトで）
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		shutdownStart := time.Now()
		err := ts.shutdown(shutdownCtx)
		shutdownDuration := time.Since(shutdownStart)

		require.Error(t, err, "タイムアウトエラーが発生するはずです")
		require.Equal(t, context.DeadlineExceeded, err, "タイムアウトエラーではありません")
		require.Greater(t, shutdownDuration, 400*time.Millisecond, "シャットダウンが早すぎます")
		require.Less(t, shutdownDuration, 600*time.Millisecond, "シャットダウンが遅すぎます")

		// サーバーが完全に停止したことを確認
		_, err = http.Get(fmt.Sprintf("http://%s/metrics", ts.addr))
		require.Error(t, err, "サーバーが停止していません")
	})
}
