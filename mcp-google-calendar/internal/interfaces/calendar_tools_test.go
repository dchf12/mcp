package interfaces

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dch/mcp-google-calendar/internal/domain"
	"github.com/dch/mcp-google-calendar/internal/usecase"
)

type mockCalendarRepository struct {
	calendars []domain.Calendar
	err       error
}

func (m *mockCalendarRepository) ListCalendars(ctx context.Context) ([]domain.Calendar, error) {
	return m.calendars, m.err
}

type mockEventRepository struct {
	event *domain.Event
	err   error
}

func (m *mockEventRepository) CreateEvent(ctx context.Context, calendarID string, event *domain.Event) (*domain.Event, error) {
	return m.event, m.err
}

func TestListCalendarTool_Execute(t *testing.T) {
	tests := []struct {
		name           string
		mockCalendars  []domain.Calendar
		mockError      error
		expectError    bool
		expectedLength int
	}{
		{
			name: "正常系: カレンダー一覧を取得",
			mockCalendars: []domain.Calendar{
				{
					ID:          "1",
					Title:       "メインカレンダー",
					Description: "説明",
					TimeZone:    "Asia/Tokyo",
				},
				{
					ID:          "2",
					Title:       "仕事用カレンダー",
					Description: "仕事の予定",
					TimeZone:    "Asia/Tokyo",
				},
			},
			mockError:      nil,
			expectError:    false,
			expectedLength: 2,
		},
		{
			name:           "異常系: UseCase実行エラー",
			mockCalendars:  nil,
			mockError:      assert.AnError,
			expectError:    true,
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			mockRepo := &mockCalendarRepository{
				calendars: tt.mockCalendars,
				err:       tt.mockError,
			}
			uc := usecase.NewGetCalendarsUseCase(mockRepo)
			tool := NewListCalendarTool(uc)

			// ツール実行
			result, err := tool.Execute(context.Background(), mcp.CallToolRequest{})

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.NotEmpty(t, result.Content)

			// メタデータの検証
			calendars, ok := result.Result.Meta["calendars"].([]domain.Calendar)
			require.True(t, ok)
			assert.Len(t, calendars, tt.expectedLength)
		})
	}
}

func TestCreateEventTool_Execute(t *testing.T) {
	now := time.Now()
	validEvent := &domain.Event{
		ID:          "1",
		Title:       "テスト会議",
		Description: "説明",
		Start: domain.DateTime{
			DateTime: now.Format(time.RFC3339),
			TimeZone: "Asia/Tokyo",
		},
		End: domain.DateTime{
			DateTime: now.Add(time.Hour).Format(time.RFC3339),
			TimeZone: "Asia/Tokyo",
		},
	}

	tests := []struct {
		name        string
		input       map[string]interface{}
		mockEvent   *domain.Event
		mockError   error
		expectError bool
	}{
		{
			name: "正常系: イベント作成成功",
			input: map[string]interface{}{
				"id":          "1",
				"title":       "テスト会議",
				"description": "説明",
				"start": map[string]interface{}{
					"dateTime": now.Format(time.RFC3339),
					"timeZone": "Asia/Tokyo",
				},
				"end": map[string]interface{}{
					"dateTime": now.Add(time.Hour).Format(time.RFC3339),
					"timeZone": "Asia/Tokyo",
				},
			},
			mockEvent:   validEvent,
			mockError:   nil,
			expectError: false,
		},
		{
			name: "異常系: UseCase実行エラー",
			input: map[string]interface{}{
				"id":          "1",
				"title":       "テスト会議",
				"description": "説明",
				"start": map[string]interface{}{
					"dateTime": now.Format(time.RFC3339),
					"timeZone": "Asia/Tokyo",
				},
				"end": map[string]interface{}{
					"dateTime": now.Add(time.Hour).Format(time.RFC3339),
					"timeZone": "Asia/Tokyo",
				},
			},
			mockEvent:   nil,
			mockError:   assert.AnError,
			expectError: true,
		},
		{
			name: "異常系: 不正な入力パラメータ",
			input: map[string]interface{}{
				"id":          "1",
				"description": "説明",
				"start": map[string]interface{}{
					"dateTime": now.Format(time.RFC3339),
					"timeZone": "Asia/Tokyo",
				},
				"end": map[string]interface{}{
					"dateTime": now.Add(time.Hour).Format(time.RFC3339),
					"timeZone": "Asia/Tokyo",
				},
			},
			mockEvent:   nil,
			mockError:   nil,
			expectError: true,
		},
		{
			name: "正常系: 説明なしでイベント作成成功",
			input: map[string]interface{}{
				"id":    "1",
				"title": "テスト会議",
				"start": map[string]interface{}{
					"dateTime": now.Format(time.RFC3339),
					"timeZone": "Asia/Tokyo",
				},
				"end": map[string]interface{}{
					"dateTime": now.Add(time.Hour).Format(time.RFC3339),
					"timeZone": "Asia/Tokyo",
				},
			},
			mockEvent:   validEvent,
			mockError:   nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			mockRepo := &mockEventRepository{
				event: tt.mockEvent,
				err:   tt.mockError,
			}
			uc := usecase.NewCreateEventUseCase(mockRepo)
			tool := NewCreateEventTool(uc)

			// ツール実行
			result, err := tool.Execute(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{Arguments: tt.input},
			})

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.NotEmpty(t, result.Content)

			// メタデータの検証
			event, ok := result.Result.Meta["event"].(*domain.Event)
			require.True(t, ok)
			assert.Equal(t, tt.mockEvent.ID, event.ID)
			assert.Equal(t, tt.mockEvent.Title, event.Title)
		})
	}
}

func TestCalendarTools_GetDefinition(t *testing.T) {
	t.Run("list_calendars", func(t *testing.T) {
		mockRepo := &mockCalendarRepository{}
		uc := usecase.NewGetCalendarsUseCase(mockRepo)
		tool := NewListCalendarTool(uc)
		def := tool.GetDefinition()

		assert.Equal(t, "list_calendars", def.Name)
		assert.NotEmpty(t, def.Description)
	})

	t.Run("create_event", func(t *testing.T) {
		mockRepo := &mockEventRepository{}
		uc := usecase.NewCreateEventUseCase(mockRepo)
		tool := NewCreateEventTool(uc)
		def := tool.GetDefinition()

		assert.Equal(t, "create_event", def.Name)
		assert.NotEmpty(t, def.Description)
	})
}
