package interfaces

import (
	"context"

	"github.com/mark3labs/mcp-go/server"

	"github.com/dch/mcp-google-calendar/internal/infrastructure/gcal"
	"github.com/dch/mcp-google-calendar/internal/usecase"
	"github.com/dch/mcp-google-calendar/pkg/config"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

// RegisterCalendarTools はカレンダー関連のツールをサーバーに登録します
//
// conf: OAuth設定
// token: 認証トークン
// server: MCPサーバーインスタンス
func RegisterCalendarTools(s *server.MCPServer, conf *config.Config, token *oauth2.Token) error {
	// OAuth設定を作成
	oauthConfig := conf.NewOAuthConfig()

	// Google APIクライアントを作成
	ctx := context.Background()
	client := oauthConfig.Client(ctx, token)

	// Google Calendarアダプタを作成
	adapter, err := gcal.New(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
	}

	// ユースケースを初期化
	getCalendarsUC := usecase.NewGetCalendarsUseCase(adapter)
	createEventUC := usecase.NewCreateEventUseCase(adapter)

	// ツールを作成
	listTool := NewListCalendarTool(getCalendarsUC)
	createTool := NewCreateEventTool(createEventUC)

	// ツールを登録
	s.AddTool(listTool.GetDefinition(), listTool.Execute)
	s.AddTool(createTool.GetDefinition(), createTool.Execute)

	return nil
}
