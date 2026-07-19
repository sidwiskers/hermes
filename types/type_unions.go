package types

import (
	"bytes"
	"encoding/json"
)

// BackgroundFill is the compact response form of Telegram's background-fill
// union. Type selects the fields populated by Telegram.
type BackgroundFill struct {
	Type          string `json:"type"`
	Color         int    `json:"color,omitempty"`
	TopColor      int    `json:"top_color,omitempty"`
	BottomColor   int    `json:"bottom_color,omitempty"`
	RotationAngle int    `json:"rotation_angle,omitempty"`
	Colors        []int  `json:"colors,omitempty"`
}

// BackgroundType is the compact response form of Telegram's background-type
// union. Type selects the fields populated by Telegram.
type BackgroundType struct {
	Type             string          `json:"type"`
	Fill             *BackgroundFill `json:"fill,omitempty"`
	DarkThemeDimming int             `json:"dark_theme_dimming,omitempty"`
	Document         *Document       `json:"document,omitempty"`
	IsBlurred        bool            `json:"is_blurred,omitempty"`
	IsMoving         bool            `json:"is_moving,omitempty"`
	Intensity        int             `json:"intensity,omitempty"`
	IsInverted       bool            `json:"is_inverted,omitempty"`
	ThemeName        string          `json:"theme_name,omitempty"`
}

// MessageOrigin is the compact response form of Telegram's forwarded-message
// origin union. Type selects the sender fields populated by Telegram.
type MessageOrigin struct {
	Type            string `json:"type"`
	Date            int64  `json:"date"`
	SenderUser      *User  `json:"sender_user,omitempty"`
	SenderUserName  string `json:"sender_user_name,omitempty"`
	SenderChat      *Chat  `json:"sender_chat,omitempty"`
	Chat            *Chat  `json:"chat,omitempty"`
	MessageID       int    `json:"message_id,omitempty"`
	AuthorSignature string `json:"author_signature,omitempty"`
}

// InputMessageContent is implemented by the five supported inline-message
// content objects. The marker prevents unrelated request values at compile
// time without adding reflection to encoding.
type InputMessageContent interface {
	inputMessageContent()
}

// InlineQueryResult is the sealed request union accepted by inline-mode
// methods. Concrete variants provide their discriminator and identifier;
// callers cannot accidentally pass an unrelated object.
type InlineQueryResult interface {
	inlineQueryResult()
	InlineQueryResultType() string
	InlineQueryResultID() string
}

// PassportElementError is the sealed request union accepted by
// setPassportDataErrors. Each concrete source exposes only its valid hash
// fields and injects the source discriminator during encoding.
type PassportElementError interface {
	passportElementError()
	PassportElementErrorSource() string
	PassportElementErrorType() string
	PassportElementErrorMessage() string
}

// PassportElementErrorRaw is the forward-compatibility form for a Passport
// error source introduced after the current Hermes release.
type PassportElementErrorRaw struct {
	Source      string   `json:"source"`
	Type        string   `json:"type"`
	FieldName   string   `json:"field_name,omitempty"`
	DataHash    string   `json:"data_hash,omitempty"`
	FileHash    string   `json:"file_hash,omitempty"`
	FileHashes  []string `json:"file_hashes,omitempty"`
	ElementHash string   `json:"element_hash,omitempty"`
	Message     string   `json:"message"`
}

func (PassportElementErrorRaw) passportElementError() {}

func (value PassportElementErrorRaw) PassportElementErrorSource() string { return value.Source }

func (value PassportElementErrorRaw) PassportElementErrorType() string { return value.Type }

func (value PassportElementErrorRaw) PassportElementErrorMessage() string { return value.Message }

// InlineQueryResultRaw is the compatibility and forward-compatibility form
// for an inline result variant not yet represented by a concrete type.
type InlineQueryResultRaw struct {
	Type string `json:"type"`
	ID   string `json:"id"`

	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	HideURL     bool   `json:"hide_url,omitempty"`

	PhotoURL    string `json:"photo_url,omitempty"`
	PhotoWidth  int    `json:"photo_width,omitempty"`
	PhotoHeight int    `json:"photo_height,omitempty"`

	GIFURL      string `json:"gif_url,omitempty"`
	GIFWidth    int    `json:"gif_width,omitempty"`
	GIFHeight   int    `json:"gif_height,omitempty"`
	GIFDuration int    `json:"gif_duration,omitempty"`

	MPEG4URL      string `json:"mpeg4_url,omitempty"`
	MPEG4Width    int    `json:"mpeg4_width,omitempty"`
	MPEG4Height   int    `json:"mpeg4_height,omitempty"`
	MPEG4Duration int    `json:"mpeg4_duration,omitempty"`

	VideoURL      string `json:"video_url,omitempty"`
	MIMEType      string `json:"mime_type,omitempty"`
	VideoWidth    int    `json:"video_width,omitempty"`
	VideoHeight   int    `json:"video_height,omitempty"`
	VideoDuration int    `json:"video_duration,omitempty"`

	AudioURL      string `json:"audio_url,omitempty"`
	AudioDuration int    `json:"audio_duration,omitempty"`
	Performer     string `json:"performer,omitempty"`

	VoiceURL      string `json:"voice_url,omitempty"`
	VoiceDuration int    `json:"voice_duration,omitempty"`
	DocumentURL   string `json:"document_url,omitempty"`
	StickerURL    string `json:"sticker_url,omitempty"`

	PhotoFileID    string `json:"photo_file_id,omitempty"`
	GIFFileID      string `json:"gif_file_id,omitempty"`
	MPEG4FileID    string `json:"mpeg4_file_id,omitempty"`
	VideoFileID    string `json:"video_file_id,omitempty"`
	AudioFileID    string `json:"audio_file_id,omitempty"`
	VoiceFileID    string `json:"voice_file_id,omitempty"`
	DocumentFileID string `json:"document_file_id,omitempty"`
	StickerFileID  string `json:"sticker_file_id,omitempty"`

	ThumbnailURL      string `json:"thumbnail_url,omitempty"`
	ThumbnailWidth    int    `json:"thumbnail_width,omitempty"`
	ThumbnailHeight   int    `json:"thumbnail_height,omitempty"`
	ThumbnailMIMEType string `json:"thumbnail_mime_type,omitempty"`

	Caption               string          `json:"caption,omitempty"`
	ParseMode             string          `json:"parse_mode,omitempty"`
	CaptionEntities       []MessageEntity `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool            `json:"show_caption_above_media,omitempty"`

	Latitude             float64 `json:"latitude,omitempty"`
	Longitude            float64 `json:"longitude,omitempty"`
	HorizontalAccuracy   float64 `json:"horizontal_accuracy,omitempty"`
	LivePeriod           int     `json:"live_period,omitempty"`
	Heading              int     `json:"heading,omitempty"`
	ProximityAlertRadius int     `json:"proximity_alert_radius,omitempty"`
	Address              string  `json:"address,omitempty"`
	FoursquareID         string  `json:"foursquare_id,omitempty"`
	FoursquareType       string  `json:"foursquare_type,omitempty"`
	GooglePlaceID        string  `json:"google_place_id,omitempty"`
	GooglePlaceType      string  `json:"google_place_type,omitempty"`

	PhoneNumber   string `json:"phone_number,omitempty"`
	FirstName     string `json:"first_name,omitempty"`
	LastName      string `json:"last_name,omitempty"`
	VCard         string `json:"vcard,omitempty"`
	GameShortName string `json:"game_short_name,omitempty"`

	InputMessageContent InputMessageContent   `json:"input_message_content,omitempty"`
	ReplyMarkup         *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (InlineQueryResultRaw) inlineQueryResult() {}

func (value InlineQueryResultRaw) InlineQueryResultType() string { return value.Type }

func (value InlineQueryResultRaw) InlineQueryResultID() string { return value.ID }

func marshalDiscriminated(kind string, value any) ([]byte, error) {
	return marshalTagged("type", kind, value)
}

func marshalTagged(field, kind string, value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	discriminator, err := json.Marshal(kind)
	if err != nil {
		return nil, err
	}
	name, err := json.Marshal(field)
	if err != nil {
		return nil, err
	}
	result := make([]byte, 0, len(data)+len(name)+len(discriminator)+4)
	result = append(result, '{')
	result = append(result, name...)
	result = append(result, ':')
	result = append(result, discriminator...)
	if len(data) > 2 {
		result = append(result, ',')
	}
	result = append(result, data[1:]...)
	return result, nil
}

// MaybeInaccessibleMessage preserves which member of Telegram's Message or
// InaccessibleMessage response union was decoded.
type MaybeInaccessibleMessage struct {
	Message             *Message
	InaccessibleMessage *InaccessibleMessage
	Raw                 json.RawMessage
}

func AccessibleMessage(message *Message) *MaybeInaccessibleMessage {
	return &MaybeInaccessibleMessage{Message: message}
}

func UnavailableMessage(message *InaccessibleMessage) *MaybeInaccessibleMessage {
	return &MaybeInaccessibleMessage{InaccessibleMessage: message}
}

func (value *MaybeInaccessibleMessage) Accessible() (*Message, bool) {
	if value == nil || value.Message == nil {
		return nil, false
	}
	return value.Message, true
}

func (value *MaybeInaccessibleMessage) Inaccessible() (*InaccessibleMessage, bool) {
	if value == nil || value.InaccessibleMessage == nil {
		return nil, false
	}
	return value.InaccessibleMessage, true
}

func (value *MaybeInaccessibleMessage) UnmarshalJSON(data []byte) error {
	var probe struct {
		Date int64 `json:"date"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	value.Message = nil
	value.InaccessibleMessage = nil
	value.Raw = append(value.Raw[:0], data...)
	if probe.Date == 0 {
		value.InaccessibleMessage = new(InaccessibleMessage)
		return json.Unmarshal(data, value.InaccessibleMessage)
	}
	value.Message = new(Message)
	return json.Unmarshal(data, value.Message)
}

func (value MaybeInaccessibleMessage) MarshalJSON() ([]byte, error) {
	switch {
	case value.Message != nil:
		return json.Marshal(value.Message)
	case value.InaccessibleMessage != nil:
		return json.Marshal(value.InaccessibleMessage)
	case len(bytes.TrimSpace(value.Raw)) != 0:
		return value.Raw, nil
	default:
		return []byte("null"), nil
	}
}
