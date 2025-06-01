package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dch/mcp-google-calendar/internal/interfaces"
	"github.com/dch/mcp-google-calendar/pkg/config"
	"github.com/mark3labs/mcp-go/server"
)

const (
	ServerName    = "Google Calendar MCP Server"
	ServerVersion = "0.1.0"
)

func main() {
	// コンテキストの設定
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// シグナルハンドリングの設定
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("シャットダウンシグナルを受信しました")
		cancel()
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
		log.Fatalf("認証情報の読み込みに失敗しました: %v", err)
	}

	conf, err := config.LoadCredentials(credPath)
	if err != nil {
		log.Fatalf("認証情報の解析に失敗しました: %v", err)
	}

	// OAuth認証フローを実行
	token, err := config.AuthFlow(ctx, conf)
	if err != nil {
		log.Fatalf("認証フローに失敗しました: %v", err)
	}

	// カレンダーツールの登録
	if err := interfaces.RegisterCalendarTools(s, conf, token); err != nil {
		log.Fatalf("カレンダーツールの登録に失敗しました: %v", err)
	}

	// サーバーの起動
	fmt.Printf("%s (v%s) を起動します\n", ServerName, ServerVersion)
	if err := server.ServeStdio(s); err != nil {
		log.Printf("サーバーエラー: %v\n", err)
		os.Exit(1)
	}
}
