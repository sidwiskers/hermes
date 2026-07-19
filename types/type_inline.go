package types

type InlineQueryResultsButton struct {
	Text           string      `json:"text"`
	WebApp         *WebAppInfo `json:"web_app,omitempty"`
	StartParameter string      `json:"start_parameter,omitempty"`
}

type SentWebAppMessage struct {
	InlineMessageID string `json:"inline_message_id,omitempty"`
}

type PreparedInlineMessage struct {
	ID             string `json:"id"`
	ExpirationDate int64  `json:"expiration_date"`
}

type PreparedKeyboardButton struct {
	ID string `json:"id"`
}

type SentGuestMessage struct {
	InlineMessageID string `json:"inline_message_id"`
}
