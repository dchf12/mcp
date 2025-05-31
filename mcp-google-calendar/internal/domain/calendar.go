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
	if e.Start.DateTime != "" && e.End.DateTime != "" {
		startTime, err := time.Parse(time.RFC3339, e.Start.DateTime)
		if err != nil {
			return fmt.Errorf("invalid start datetime format: %w", err)
		}
		endTime, err := time.Parse(time.RFC3339, e.End.DateTime)
		if err != nil {
			return fmt.Errorf("invalid end datetime format: %w", err)
		}
		if endTime.Before(startTime) {
			return fmt.Errorf("event end time cannot be before start time")
		}
	}
	return nil
}
