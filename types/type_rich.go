package types

import "encoding/json"

// RichText is rich inline content. Telegram accepts plain strings, slices of
// RichText, and the typed RichText* values declared below.
type RichText = any

// richTextEnvelope keeps discriminator injection in one place. It is used by
// the small MarshalJSON methods below so callers never set Telegram's type
// field manually.
func richTextEnvelope(kind string, value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || data[0] != '{' {
		return data, nil
	}
	prefix, err := json.Marshal(kind)
	if err != nil {
		return nil, err
	}
	result := make([]byte, 0, len(data)+len(prefix)+8)
	result = append(result, `{"type":`...)
	result = append(result, prefix...)
	if len(data) > 2 {
		result = append(result, ',')
	}
	result = append(result, data[1:]...)
	return result, nil
}

type RichTextBold struct {
	Text RichText `json:"text"`
}

func (value RichTextBold) MarshalJSON() ([]byte, error) {
	type plain RichTextBold
	return richTextEnvelope("bold", plain(value))
}

type RichTextItalic struct {
	Text RichText `json:"text"`
}

func (value RichTextItalic) MarshalJSON() ([]byte, error) {
	type plain RichTextItalic
	return richTextEnvelope("italic", plain(value))
}

type RichTextUnderline struct {
	Text RichText `json:"text"`
}

func (value RichTextUnderline) MarshalJSON() ([]byte, error) {
	type plain RichTextUnderline
	return richTextEnvelope("underline", plain(value))
}

type RichTextStrikethrough struct {
	Text RichText `json:"text"`
}

func (value RichTextStrikethrough) MarshalJSON() ([]byte, error) {
	type plain RichTextStrikethrough
	return richTextEnvelope("strikethrough", plain(value))
}

type RichTextSpoiler struct {
	Text RichText `json:"text"`
}

func (value RichTextSpoiler) MarshalJSON() ([]byte, error) {
	type plain RichTextSpoiler
	return richTextEnvelope("spoiler", plain(value))
}

type RichTextDateTime struct {
	Text           RichText `json:"text"`
	UnixTime       int64    `json:"unix_time"`
	DateTimeFormat string   `json:"date_time_format"`
}

func (value RichTextDateTime) MarshalJSON() ([]byte, error) {
	type plain RichTextDateTime
	return richTextEnvelope("date_time", plain(value))
}

type RichTextTextMention struct {
	Text RichText `json:"text"`
	User User     `json:"user"`
}

func (value RichTextTextMention) MarshalJSON() ([]byte, error) {
	type plain RichTextTextMention
	return richTextEnvelope("text_mention", plain(value))
}

type RichTextSubscript struct {
	Text RichText `json:"text"`
}

func (value RichTextSubscript) MarshalJSON() ([]byte, error) {
	type plain RichTextSubscript
	return richTextEnvelope("subscript", plain(value))
}

type RichTextSuperscript struct {
	Text RichText `json:"text"`
}

func (value RichTextSuperscript) MarshalJSON() ([]byte, error) {
	type plain RichTextSuperscript
	return richTextEnvelope("superscript", plain(value))
}

type RichTextMarked struct {
	Text RichText `json:"text"`
}

func (value RichTextMarked) MarshalJSON() ([]byte, error) {
	type plain RichTextMarked
	return richTextEnvelope("marked", plain(value))
}

type RichTextCode struct {
	Text RichText `json:"text"`
}

func (value RichTextCode) MarshalJSON() ([]byte, error) {
	type plain RichTextCode
	return richTextEnvelope("code", plain(value))
}

type RichTextCustomEmoji struct {
	CustomEmojiID   string `json:"custom_emoji_id"`
	AlternativeText string `json:"alternative_text"`
}

func (value RichTextCustomEmoji) MarshalJSON() ([]byte, error) {
	type plain RichTextCustomEmoji
	return richTextEnvelope("custom_emoji", plain(value))
}

type RichTextMathematicalExpression struct {
	Expression string `json:"expression"`
}

func (value RichTextMathematicalExpression) MarshalJSON() ([]byte, error) {
	type plain RichTextMathematicalExpression
	return richTextEnvelope("mathematical_expression", plain(value))
}

type RichTextURL struct {
	Text RichText `json:"text"`
	URL  string   `json:"url"`
}

func (value RichTextURL) MarshalJSON() ([]byte, error) {
	type plain RichTextURL
	return richTextEnvelope("url", plain(value))
}

type RichTextEmailAddress struct {
	Text         RichText `json:"text"`
	EmailAddress string   `json:"email_address"`
}

func (value RichTextEmailAddress) MarshalJSON() ([]byte, error) {
	type plain RichTextEmailAddress
	return richTextEnvelope("email_address", plain(value))
}

type RichTextPhoneNumber struct {
	Text        RichText `json:"text"`
	PhoneNumber string   `json:"phone_number"`
}

func (value RichTextPhoneNumber) MarshalJSON() ([]byte, error) {
	type plain RichTextPhoneNumber
	return richTextEnvelope("phone_number", plain(value))
}

type RichTextBankCardNumber struct {
	Text           RichText `json:"text"`
	BankCardNumber string   `json:"bank_card_number"`
}

func (value RichTextBankCardNumber) MarshalJSON() ([]byte, error) {
	type plain RichTextBankCardNumber
	return richTextEnvelope("bank_card_number", plain(value))
}

type RichTextMention struct {
	Text     RichText `json:"text"`
	Username string   `json:"username"`
}

func (value RichTextMention) MarshalJSON() ([]byte, error) {
	type plain RichTextMention
	return richTextEnvelope("mention", plain(value))
}

type RichTextHashtag struct {
	Text    RichText `json:"text"`
	Hashtag string   `json:"hashtag"`
}

func (value RichTextHashtag) MarshalJSON() ([]byte, error) {
	type plain RichTextHashtag
	return richTextEnvelope("hashtag", plain(value))
}

type RichTextCashtag struct {
	Text    RichText `json:"text"`
	Cashtag string   `json:"cashtag"`
}

func (value RichTextCashtag) MarshalJSON() ([]byte, error) {
	type plain RichTextCashtag
	return richTextEnvelope("cashtag", plain(value))
}

type RichTextBotCommand struct {
	Text       RichText `json:"text"`
	BotCommand string   `json:"bot_command"`
}

func (value RichTextBotCommand) MarshalJSON() ([]byte, error) {
	type plain RichTextBotCommand
	return richTextEnvelope("bot_command", plain(value))
}

type RichTextAnchor struct {
	Name string `json:"name"`
}

func (value RichTextAnchor) MarshalJSON() ([]byte, error) {
	type plain RichTextAnchor
	return richTextEnvelope("anchor", plain(value))
}

type RichTextAnchorLink struct {
	Text       RichText `json:"text"`
	AnchorName string   `json:"anchor_name"`
}

func (value RichTextAnchorLink) MarshalJSON() ([]byte, error) {
	type plain RichTextAnchorLink
	return richTextEnvelope("anchor_link", plain(value))
}

type RichTextReference struct {
	Text RichText `json:"text"`
	Name string   `json:"name"`
}

func (value RichTextReference) MarshalJSON() ([]byte, error) {
	type plain RichTextReference
	return richTextEnvelope("reference", plain(value))
}

type RichTextReferenceLink struct {
	Text          RichText `json:"text"`
	ReferenceName string   `json:"reference_name"`
}

func (value RichTextReferenceLink) MarshalJSON() ([]byte, error) {
	type plain RichTextReferenceLink
	return richTextEnvelope("reference_link", plain(value))
}

type RichBlockCaption struct {
	Text   RichText `json:"text"`
	Credit RichText `json:"credit,omitempty"`
}

type RichBlockTableCell struct {
	Text     RichText `json:"text,omitempty"`
	IsHeader bool     `json:"is_header,omitempty"`
	Colspan  int      `json:"colspan,omitempty"`
	Rowspan  int      `json:"rowspan,omitempty"`
	Align    string   `json:"align"`
	VAlign   string   `json:"valign"`
}

type RichBlockListItem struct {
	RichBlockListItemBotAPIFields
	Blocks      []RichBlock `json:"blocks"`
	HasCheckbox bool        `json:"has_checkbox,omitempty"`
	IsChecked   bool        `json:"is_checked,omitempty"`
	Value       int         `json:"value,omitempty"`
	Type        string      `json:"type,omitempty"`
}

// RichBlock is a compact typed representation of every received rich block.
// Type identifies which subset of fields is populated.
type RichBlock struct {
	Type       string                 `json:"type"`
	Text       RichText               `json:"text,omitempty"`
	Size       int                    `json:"size,omitempty"`
	Language   string                 `json:"language,omitempty"`
	Expression string                 `json:"expression,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Items      []RichBlockListItem    `json:"items,omitempty"`
	Blocks     []RichBlock            `json:"blocks,omitempty"`
	Credit     RichText               `json:"credit,omitempty"`
	Caption    *RichBlockCaption      `json:"caption,omitempty"`
	Cells      [][]RichBlockTableCell `json:"cells,omitempty"`
	IsBordered bool                   `json:"is_bordered,omitempty"`
	IsStriped  bool                   `json:"is_striped,omitempty"`
	Summary    RichText               `json:"summary,omitempty"`
	IsOpen     bool                   `json:"is_open,omitempty"`
	Location   *Location              `json:"location,omitempty"`
	Zoom       int                    `json:"zoom,omitempty"`
	Width      int                    `json:"width,omitempty"`
	Height     int                    `json:"height,omitempty"`
	Animation  *Animation             `json:"animation,omitempty"`
	Audio      *Audio                 `json:"audio,omitempty"`
	Photo      []PhotoSize            `json:"photo,omitempty"`
	Video      *Video                 `json:"video,omitempty"`
	VoiceNote  *Voice                 `json:"voice_note,omitempty"`
	HasSpoiler bool                   `json:"has_spoiler,omitempty"`
}

type RichMessage struct {
	Blocks []RichBlock `json:"blocks"`
	IsRTL  bool        `json:"is_rtl,omitempty"`
}
