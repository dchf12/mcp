package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dch/mcp-google-calendar/internal/interfaces"
	"github.com/dch/mcp-google-calendar/pkg/config"
	"github.com/mark3labs/mcp-go/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	ServerName      = "Google Calendar MCP Server"
	ServerVersion   = "0.1.0"
	ShutdownTimeout = 10 * time.Second
	MetricsAddr     = ":9090"
)

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})))
}

func main() {
	// コンテキストの設定
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// シグナルハンドリングの設定
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		slog.Info("シャットダウンシグナルを受信しました", "signal", sig)

		// グレースフルシャットダウンのためのコンテキスト
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), ShutdownTimeout)
		defer cancelShutdown()

		// コンテキストのキャンセルを実行
		cancel()

		// サブシステムの停止を待機
		doneChan := make(chan struct{})
		go func() {
			wg.Wait()
			close(doneChan)
		}()

		// シャットダウンタイムアウトの監視
		select {
		case <-shutdownCtx.Done():
			slog.Warn("シャットダウンタイムアウト - 強制終了します")
		case <-doneChan:
			slog.Info("全てのサブシステムが正常に停止しました")
		}
	}()

	// MCPサーバーの初期化
	s := server.NewMCPServer(
		ServerName,
		ServerVersion,
		server.WithLogging(),
		server.WithRecovery(),
	)

	// 認証情報の設定を読み込み
	credPath, err := config.GetCredentialsPath()
	if err != nil {
		slog.Error("認証情報の読み込みに失敗しました", "error", err)
		os.Exit(1)
	}

	conf, err := config.LoadCredentials(credPath)
	if err != nil {
		slog.Error("認証情報の解析に失敗しました", "error", err)
		os.Exit(1)
	}

	// OAuth認証フローを実行
	token, err := config.AuthFlow(ctx, conf)
	if err != nil {
		slog.Error("認証フローに失敗しました", "error", err)
		os.Exit(1)
	}

	// カレンダーツールの登録
	if err := interfaces.RegisterCalendarTools(s, conf, token); err != nil {
		slog.Error("カレンダーツールの登録に失敗しました", "error", err)
		os.Exit(1)
	}

	// メトリクスサーバーの起動
	metricsServer := &http.Server{
		Addr:    MetricsAddr,
		Handler: promhttp.Handler(),
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("メトリクスサーバーを起動します", "addr", MetricsAddr)
		if err := metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("メトリクスサーバーエラー", "error", err)
		}
	}()

	// メトリクスサーバーのグレースフルシャットダウン
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		slog.Info("メトリクスサーバーをシャットダウンします")
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("メトリクスサーバーのシャットダウンに失敗しました", "error", err)
		}
	}()

	// MCPサーバーの起動
	slog.Info("サーバーを起動します", "name", ServerName, "version", ServerVersion)
	if err := server.ServeStdio(s); err != nil {
		slog.Error("サーバーエラー", "error", err)
		os.Exit(1)
	}
}
