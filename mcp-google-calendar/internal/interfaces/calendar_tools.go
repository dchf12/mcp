package interfaces

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mitchellh/mapstructure"

	"github.com/dch/mcp-google-calendar/internal/domain"
	"github.com/dch/mcp-google-calendar/internal/usecase"
)

// ListCalendarTool は list_calendars ツールの入出力を定義します
type ListCalendarTool struct {
	getCalendarsUseCase *usecase.GetCalendarsUseCase
}

// NewListCalendarTool は ListCalendarTool の新しいインスタンスを作成します
func NewListCalendarTool(getCalendarsUseCase *usecase.GetCalendarsUseCase) *ListCalendarTool {
	return &ListCalendarTool{
		getCalendarsUseCase: getCalendarsUseCase,
	}
}

// GetDefinition は list_calendars ツールの定義を返します
func (t *ListCalendarTool) GetDefinition() mcp.Tool {
	return mcp.NewTool("list_calendars",
		mcp.WithDescription("利用可能なカレンダーの一覧を取得します"),
	)
}

// Execute は list_calendars ツールを実行します
func (t *ListCalendarTool) Execute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	calendars, err := t.getCalendarsUseCase.Execute(ctx)
	if err != nil {
		slog.Error("failed to get calendars", "error", err)
		return nil, fmt.Errorf("カレンダーの取得に失敗しました: %w", err)
	}

	// MCPクライアントが解析しやすいJSON形式でレスポンスを作成
	response := map[string]any{
		"status":    "success",
		"message":   "カレンダーの一覧を取得しました",
		"count":     len(calendars),
		"calendars": calendars,
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("レスポンスのJSON化に失敗しました: %w", err)
	}

	return &mcp.CallToolResult{
		Result:  mcp.Result{Meta: map[string]any{"calendars": calendars}},
		Content: []mcp.Content{mcp.NewTextContent(string(jsonBytes))},
		IsError: false,
	}, nil
}

// CreateEventTool は create_event ツールの入出力を定義します
type CreateEventTool struct {
	createEventUseCase *usecase.CreateEventUseCase
}

// CreateEventInput は create_event の入力を定義します
type CreateEventInput struct {
	CalendarID  string          `json:"calendar_id"`
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Start       domain.DateTime `json:"start"`
	End         domain.DateTime `json:"end"`
	Location    *string         `json:"location,omitempty"`
	Attendees   []string        `json:"attendees,omitempty"`
}

// NewCreateEventTool は CreateEventTool の新しいインスタンスを作成します
func NewCreateEventTool(createEventUseCase *usecase.CreateEventUseCase) *CreateEventTool {
	return &CreateEventTool{
		createEventUseCase: createEventUseCase,
	}
}

// GetDefinition は create_event ツールの定義を返します
func (t *CreateEventTool) GetDefinition() mcp.Tool {
	return mcp.NewTool("create_event",
		mcp.WithDescription("指定したカレンダーに新しいイベントを作成します"),
		mcp.WithString("calendar_id",
			mcp.Required(),
			mcp.Description("イベントを作成するカレンダーのID"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("イベントのタイトル"),
		),
		mcp.WithString("description",
			mcp.Description("イベントの説明"),
		),
		mcp.WithObject("start",
			mcp.Required(),
			mcp.Description("開始日時"),
		),
		mcp.WithObject("end",
			mcp.Required(),
			mcp.Description("終了日時"),
		),
		mcp.WithString("location",
			mcp.Description("イベントの場所"),
		),
		mcp.WithArray("attendees",
			mcp.Description("参加者のメールアドレス"),
		),
	)
}

// Execute は create_event ツールを実行します
func (t *CreateEventTool) Execute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input CreateEventInput
	if request.Params.Arguments == nil {
		return nil, fmt.Errorf("入力パラメータが指定されていません")
	}

	// デバッグのために引数をログ出力
	slog.Info("received create_event arguments", "arguments", request.Params.Arguments)

	// mapstructureの設定を調整してデコード
	config := &mapstructure.DecoderConfig{
		Result:           &input,
		WeaklyTypedInput: true,
		TagName:          "json",
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, fmt.Errorf("デコーダーの作成に失敗しました: %w", err)
	}

	if err := decoder.Decode(request.Params.Arguments); err != nil {
		return nil, fmt.Errorf("入力パラメータの解析に失敗しました: %w", err)
	}

	// デコード後の値をログ出力
	slog.Info("decoded input", "calendar_id", input.CalendarID, "title", input.Title)

	event := &domain.Event{
		Title:       input.Title,
		Description: input.Description,
		Start:       input.Start,
		End:         input.End,
		Location:    input.Location,
		Attendees:   input.Attendees,
	}

	createdEvent, err := t.createEventUseCase.Execute(ctx, input.CalendarID, event)
	if err != nil {
		slog.Error("failed to create event", "calendar_id", input.CalendarID, "title", input.Title, "error", err)

		// より詳細なエラーメッセージを作成
		errorMsg := fmt.Sprintf("イベントの作成に失敗しました (カレンダーID: %s, タイトル: %s): %v",
			input.CalendarID, input.Title, err)

		return nil, fmt.Errorf("イベントの作成に失敗しました: %s", errorMsg)
	}

	// MCPクライアントが解析しやすいJSON形式でレスポンスを作成
	response := map[string]any{
		"status":  "success",
		"message": "イベントを作成しました",
		"event":   createdEvent,
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("レスポンスのJSON化に失敗しました: %w", err)
	}

	return &mcp.CallToolResult{
		Result:  mcp.Result{Meta: map[string]any{"event": createdEvent}},
		Content: []mcp.Content{mcp.NewTextContent(string(jsonBytes))},
		IsError: false,
	}, nil
}
