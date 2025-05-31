package domain

import (
	"fmt"
	"time"
)

// Calendar はカレンダーのドメインエンティティです
type Calendar struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	TimeZone    string `json:"timeZone"`
}

// Event はカレンダーイベントのドメインエンティティです
type Event struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Start       DateTime `json:"start"`
	End         DateTime `json:"end"`
	Location    *string  `json:"location,omitempty"`
	Attendees   []string `json:"attendees,omitempty"`
}

// DateTime は日時を表現する値オブジェクトです
type DateTime struct {
	DateTime string `json:"dateTime,omitempty"`
	Date     string `json:"date,omitempty"`
	TimeZone string `json:"timeZone,omitempty"`
}

// Validate はDateTimeオブジェクトのバリデーションを行います
func (dt *DateTime) Validate() error {
	if dt.DateTime != "" {
		_, err := time.Parse(time.RFC3339, dt.DateTime)
		if err != nil {
			return fmt.Errorf("invalid datetime format, must be RFC3339: %w", err)
		}
	}
	if dt.TimeZone != "" {
		loc, err := time.LoadLocation(dt.TimeZone)
		if err != nil {
			return fmt.Errorf("invalid timezone: %w", err)
		}
		if loc == nil {
			return fmt.Errorf("invalid timezone: location is nil")
		}
	}
	return nil
}

// Validate はCalendarエンティティのバリデーションを行います
func (c *Calendar) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("calendar ID is required")
	}
	if c.Title == "" {
		return fmt.Errorf("calendar title is required")
	}
	return nil
}

// Validate はEventエンティティのバリデーションを行います
func (e *Event) Validate() error {
	if e.Title == "" {
		return fmt.Errorf("event title is required")
	}
	if err := e.validateDateTime(); err != nil {
		return err
	}
	return nil
}

// validateDateTime はイベントの日時のバリデーションを行います
func (e *Event) validateDateTime() error {
	if e.Start.DateTime == "" && e.Start.Date == "" {
		return fmt.Errorf("event start date/time is required")
	}
	if e.End.DateTime == "" && e.End.Date == "" {
		return fmt.Errorf("event end date/time is required")
	}

	// 日時形式とタイムゾーンのバリデーション
	if err := e.Start.Validate(); err != nil {
		return fmt.Errorf("invalid start time: %w", err)
	}
	if err := e.End.Validate(); err != nil {
		return fmt.Errorf("invalid end time: %w", err)
	}

	// 開始時刻と終了時刻の前後関係のチェック
	if e.Start.DateTime != "" && e.End.DateTime != "" {
		startTime, _ := time.Parse(time.RFC3339, e.Start.DateTime) // エラーは既にValidate()でチェック済み
		endTime, _ := time.Parse(time.RFC3339, e.End.DateTime)     // エラーは既にValidate()でチェック済み
		if endTime.Before(startTime) {
			return fmt.Errorf("event end time cannot be before start time")
		}
	}
	return nil
}
