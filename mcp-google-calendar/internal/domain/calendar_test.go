package domain

import (
	"testing"
)

func TestCalendar_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cal     Calendar
		wantErr bool
	}{
		{
			name: "valid calendar",
			cal: Calendar{
				ID:    "primary",
				Title: "メインカレンダー",
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			cal: Calendar{
				Title: "メインカレンダー",
			},
			wantErr: true,
		},
		{
			name: "missing title",
			cal: Calendar{
				ID: "primary",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cal.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Calendar.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   Event
		wantErr bool
	}{
		{
			name: "valid event with datetime",
			event: Event{
				Title: "会議",
				Start: DateTime{DateTime: "2025-06-01T10:00:00+09:00"},
				End:   DateTime{DateTime: "2025-06-01T11:00:00+09:00"},
			},
			wantErr: false,
		},
		{
			name: "valid event with date",
			event: Event{
				Title: "終日イベント",
				Start: DateTime{Date: "2025-06-01"},
				End:   DateTime{Date: "2025-06-01"},
			},
			wantErr: false,
		},
		{
			name: "missing title",
			event: Event{
				Start: DateTime{DateTime: "2025-06-01T10:00:00+09:00"},
				End:   DateTime{DateTime: "2025-06-01T11:00:00+09:00"},
			},
			wantErr: true,
		},
		{
			name: "missing start time",
			event: Event{
				Title: "会議",
				End:   DateTime{DateTime: "2025-06-01T11:00:00+09:00"},
			},
			wantErr: true,
		},
		{
			name: "missing end time",
			event: Event{
				Title: "会議",
				Start: DateTime{DateTime: "2025-06-01T10:00:00+09:00"},
			},
			wantErr: true,
		},
		{
			name: "end before start",
			event: Event{
				Title: "会議",
				Start: DateTime{DateTime: "2025-06-01T11:00:00+09:00"},
				End:   DateTime{DateTime: "2025-06-01T10:00:00+09:00"},
			},
			wantErr: true,
		},
		{
			name: "invalid datetime format",
			event: Event{
				Title: "会議",
				Start: DateTime{DateTime: "2025-06-01 10:00:00"},
				End:   DateTime{DateTime: "2025-06-01T11:00:00+09:00"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Event.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
