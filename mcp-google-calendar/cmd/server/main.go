package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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
)

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})))
}

func main() {
	// コンテキストの設定
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// シグナルハンドリングの設定
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		slog.Info("シャットダウンシグナルを受信しました", "signal", sig)

		// グレースフルシャットダウンのためのコンテキスト
		ctx, cancelShutdown := context.WithTimeout(context.Background(), ShutdownTimeout)
		defer cancelShutdown()

		// ここでクリーンアップ処理を実行
		slog.Info("クリーンアップ処理を開始します...")

		// キャンセル実行
		cancel()

		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				slog.Warn("強制シャットダウンを実行します")
			}
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

	// メトリクスエンドポイントの追加
	http.Handle("/metrics", promhttp.Handler())

	// サーバーの起動
	slog.Info("サーバーを起動します", "name", ServerName, "version", ServerVersion)
	if err := server.ServeStdio(s); err != nil {
		slog.Error("サーバーエラー", "error", err)
		os.Exit(1)
	}
}
