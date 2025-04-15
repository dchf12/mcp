package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// MCPサーバーの初期化
	s := server.NewMCPServer(
		"Filesystem Server",
		"1.0.0",
		server.WithLogging(),
		server.WithRecovery(),
	)

	// ツールの登録
	s.AddTool(mcp.NewTool("list_directory",
		mcp.WithDescription("指定されたディレクトリの内容をリストアップします"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("リストアップするディレクトリのパス"),
		),
	), handleListDirectory)

	s.AddTool(mcp.NewTool("file_content",
		mcp.WithDescription("指定されたファイルの内容を取得します"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("内容を取得するファイルのパス"),
		),
	), handleFileContent)

	// サーバーの起動
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("サーバーエラー: %v\n", err)
	}
}

func handleListDirectory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok || path == "" {
		return nil, errors.New("有効なパスが指定されていません")
	}

	files, err := os.ReadDir(path)
	if err != nil {
		errMsg := fmt.Sprintf("ディレクトリ '%s' の読み取りに失敗しました: %v", path, err)
		fmt.Println(errMsg) // エラーメッセージをログに出力
		return mcp.NewToolResultText(errMsg), nil
	}

	var results []map[string]any
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		results = append(results, map[string]any{
			"name":        file.Name(),
			"isDir":       file.IsDir(),
			"size":        info.Size(),
			"modTime":     info.ModTime().Format(time.DateTime),
			"permissions": info.Mode().String(),
		})
	}
	return &mcp.CallToolResult{
		Result:  mcp.Result{Meta: map[string]any{"files": results}},
		Content: []mcp.Content{mcp.NewTextContent(strings.Join([]string{"ディレクトリの内容を取得しました", fmt.Sprintf("パス: %s", path)}, "\n"))},
		IsError: false,
	}, nil
}

func handleFileContent(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok || path == "" {
		return nil, errors.New("有効なパスが指定されていません")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ファイルを開けませんでした: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("ファイルの読み取りに失敗しました: %v", err)
	}

	return mcp.NewToolResultText(string(content)), nil
}

func fileTypeStr(isDir bool) string {
	if isDir {
		return "ディレクトリ"
	}
	return "ファイル"
}
