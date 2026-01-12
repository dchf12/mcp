package interfaces

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

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

		return nil, fmt.Errorf(
			"イベントの作成に失敗しました (カレンダーID: %s, タイトル: %s): %w",
			input.CalendarID,
			input.Title,
			err,
		)
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

// ListEventsTool は list_events ツールの入出力を定義します
type ListEventsTool struct {
	getEventsUseCase *usecase.GetEventsUseCase
}

// ListEventsInput は list_events の入力を定義します
type ListEventsInput struct {
	CalendarID string `json:"calendar_id"`
	TimeMin    string `json:"time_min"`
	TimeMax    string `json:"time_max"`
	MaxResults *int   `json:"max_results,omitempty"`
}

// NewListEventsTool は ListEventsTool の新しいインスタンスを作成します
func NewListEventsTool(getEventsUseCase *usecase.GetEventsUseCase) *ListEventsTool {
	return &ListEventsTool{
		getEventsUseCase: getEventsUseCase,
	}
}

// GetDefinition は list_events ツールの定義を返します
func (t *ListEventsTool) GetDefinition() mcp.Tool {
	return mcp.NewTool("list_events",
		mcp.WithDescription("指定したカレンダーから予定一覧を取得します"),
		mcp.WithString("calendar_id",
			mcp.Required(),
			mcp.Description("予定を取得するカレンダーのID"),
		),
		mcp.WithString("time_min",
			mcp.Required(),
			mcp.Description("取得開始日時（RFC3339形式、例: 2025-01-01T00:00:00Z）"),
		),
		mcp.WithString("time_max",
			mcp.Required(),
			mcp.Description("取得終了日時（RFC3339形式、例: 2025-01-31T23:59:59Z）"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("最大取得件数（デフォルト: 250）"),
		),
	)
}

// Execute は list_events ツールを実行します
func (t *ListEventsTool) Execute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input ListEventsInput
	if request.Params.Arguments == nil {
		return nil, fmt.Errorf("入力パラメータが指定されていません")
	}

	slog.Info("received list_events arguments", "arguments", request.Params.Arguments)

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

	slog.Info("decoded input", "calendar_id", input.CalendarID, "time_min", input.TimeMin, "time_max", input.TimeMax)

	// RFC3339形式の文字列をtime.Timeに変換
	timeMin, err := time.Parse(time.RFC3339, input.TimeMin)
	if err != nil {
		return nil, fmt.Errorf("time_min の形式が不正です（RFC3339形式で指定してください）: %w", err)
	}

	timeMax, err := time.Parse(time.RFC3339, input.TimeMax)
	if err != nil {
		return nil, fmt.Errorf("time_max の形式が不正です（RFC3339形式で指定してください）: %w", err)
	}

	params := domain.GetEventsParams{
		CalendarID: input.CalendarID,
		TimeMin:    timeMin,
		TimeMax:    timeMax,
	}

	if input.MaxResults != nil {
		params.MaxResults = *input.MaxResults
	}

	events, err := t.getEventsUseCase.Execute(ctx, params)
	if err != nil {
		slog.Error("failed to get events", "calendar_id", input.CalendarID, "error", err)
		return nil, fmt.Errorf(
			"予定の取得に失敗しました (カレンダーID: %s): %w",
			input.CalendarID,
			err,
		)
	}

	response := map[string]any{
		"status":   "success",
		"message":  "予定一覧を取得しました",
		"count":    len(events),
		"events":   events,
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("レスポンスのJSON化に失敗しました: %w", err)
	}

	return &mcp.CallToolResult{
		Result:  mcp.Result{Meta: map[string]any{"events": events}},
		Content: []mcp.Content{mcp.NewTextContent(string(jsonBytes))},
		IsError: false,
	}, nil
}
