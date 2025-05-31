package domain

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
