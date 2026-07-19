package api

import (
	"context"
	"fmt"
	"strings"
)

const (
	PollRegular = "regular"
	PollQuiz    = "quiz"

	DiceEmoji       = "🎲"
	DartsEmoji      = "🎯"
	BasketballEmoji = "🏀"
	FootballEmoji   = "⚽"
	BowlingEmoji    = "🎳"
	SlotsEmoji      = "🎰"
)

// Bool returns a pointer for optional Bot API booleans whose false value is
// meaningful, such as sendPoll.is_anonymous.
func Bool(value bool) *bool { return &value }

type InputPollOption struct {
	Text          string               `json:"text"`
	TextParseMode string               `json:"text_parse_mode,omitempty"`
	TextEntities  []MessageEntity      `json:"text_entities,omitempty"`
	Media         InputPollOptionMedia `json:"media,omitempty"`
}

type SendPollParams struct {
	BusinessConnectionID   string            `json:"business_connection_id,omitempty"`
	ChatID                 any               `json:"chat_id"`
	MessageThreadID        int               `json:"message_thread_id,omitempty"`
	Question               string            `json:"question"`
	QuestionParseMode      string            `json:"question_parse_mode,omitempty"`
	QuestionEntities       []MessageEntity   `json:"question_entities,omitempty"`
	Options                []InputPollOption `json:"options"`
	IsAnonymous            *bool             `json:"is_anonymous,omitempty"`
	Type                   string            `json:"type,omitempty"`
	AllowsMultipleAnswers  bool              `json:"allows_multiple_answers,omitempty"`
	AllowsRevoting         *bool             `json:"allows_revoting,omitempty"`
	ShuffleOptions         bool              `json:"shuffle_options,omitempty"`
	AllowAddingOptions     bool              `json:"allow_adding_options,omitempty"`
	HideResultsUntilCloses bool              `json:"hide_results_until_closes,omitempty"`
	MembersOnly            bool              `json:"members_only,omitempty"`
	CountryCodes           []string          `json:"country_codes,omitempty"`
	CorrectOptionIDs       []int             `json:"correct_option_ids,omitempty"`
	Explanation            string            `json:"explanation,omitempty"`
	ExplanationParseMode   string            `json:"explanation_parse_mode,omitempty"`
	ExplanationEntities    []MessageEntity   `json:"explanation_entities,omitempty"`
	ExplanationMedia       InputPollMedia    `json:"explanation_media,omitempty"`
	OpenPeriod             int               `json:"open_period,omitempty"`
	CloseDate              int64             `json:"close_date,omitempty"`
	IsClosed               bool              `json:"is_closed,omitempty"`
	Description            string            `json:"description,omitempty"`
	DescriptionParseMode   string            `json:"description_parse_mode,omitempty"`
	DescriptionEntities    []MessageEntity   `json:"description_entities,omitempty"`
	Media                  InputPollMedia    `json:"media,omitempty"`
	DisableNotification    bool              `json:"disable_notification,omitempty"`
	ProtectContent         bool              `json:"protect_content,omitempty"`
	AllowPaidBroadcast     bool              `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID        string            `json:"message_effect_id,omitempty"`
	ReplyParameters        *ReplyParameters  `json:"reply_parameters,omitempty"`
	ReplyMarkup            ReplyMarkup       `json:"reply_markup,omitempty"`
}

func validateSendPoll(params SendPollParams) error {
	if err := validateChatID(params.ChatID, "sendPoll"); err != nil {
		return err
	}
	question := strings.TrimSpace(params.Question)
	if question == "" || len([]rune(question)) > 300 {
		return fmt.Errorf("hermes: sendPoll question must contain 1-300 characters")
	}
	if params.QuestionParseMode != "" && len(params.QuestionEntities) != 0 {
		return fmt.Errorf("hermes: sendPoll question parse mode and entities are mutually exclusive")
	}
	if len(params.Options) < 1 || len(params.Options) > 12 {
		return fmt.Errorf("hermes: sendPoll requires 1-12 options")
	}
	for index, option := range params.Options {
		length := len([]rune(strings.TrimSpace(option.Text)))
		if length < 1 || length > 100 {
			return fmt.Errorf("hermes: sendPoll option %d must contain 1-100 characters", index)
		}
		if option.TextParseMode != "" && len(option.TextEntities) != 0 {
			return fmt.Errorf("hermes: sendPoll option %d parse mode and entities are mutually exclusive", index)
		}
		if option.Media != nil && !option.Media.validPollMedia() {
			return fmt.Errorf("hermes: sendPoll option %d has invalid media", index)
		}
	}
	if params.OpenPeriod != 0 && params.CloseDate != 0 {
		return fmt.Errorf("hermes: sendPoll open_period and close_date are mutually exclusive")
	}
	if params.OpenPeriod != 0 && (params.OpenPeriod < 5 || params.OpenPeriod > 2628000) {
		return fmt.Errorf("hermes: sendPoll open_period must be 5-2628000 seconds")
	}
	if params.Type != "" && params.Type != PollRegular && params.Type != PollQuiz {
		return fmt.Errorf("hermes: sendPoll type must be regular or quiz")
	}
	if params.Type == PollQuiz {
		if len(params.CorrectOptionIDs) == 0 {
			return fmt.Errorf("hermes: quiz polls require correct_option_ids")
		}
		if params.AllowsMultipleAnswers {
			return fmt.Errorf("hermes: quiz polls can't allow multiple answers")
		}
	}
	previous := -1
	for _, optionID := range params.CorrectOptionIDs {
		if optionID < 0 || optionID >= len(params.Options) || optionID <= previous {
			return fmt.Errorf("hermes: correct_option_ids must be unique, increasing option indexes")
		}
		previous = optionID
	}
	if params.AllowAddingOptions {
		if params.Type == PollQuiz || params.IsAnonymous == nil || *params.IsAnonymous {
			return fmt.Errorf("hermes: adding poll options requires a non-anonymous regular poll")
		}
	}
	if len(params.CountryCodes) > 12 {
		return fmt.Errorf("hermes: sendPoll accepts at most 12 country codes")
	}
	for _, code := range params.CountryCodes {
		if len(code) != 2 {
			return fmt.Errorf("hermes: invalid poll country code %q", code)
		}
	}
	if len([]rune(params.Explanation)) > 200 || strings.Count(params.Explanation, "\n") > 2 {
		return fmt.Errorf("hermes: poll explanation exceeds Telegram limits")
	}
	if params.ExplanationParseMode != "" && len(params.ExplanationEntities) != 0 {
		return fmt.Errorf("hermes: poll explanation parse mode and entities are mutually exclusive")
	}
	if len([]rune(params.Description)) > 1024 {
		return fmt.Errorf("hermes: poll description must not exceed 1024 characters")
	}
	if params.DescriptionParseMode != "" && len(params.DescriptionEntities) != 0 {
		return fmt.Errorf("hermes: poll description parse mode and entities are mutually exclusive")
	}
	if params.ExplanationMedia != nil && !params.ExplanationMedia.validPollMedia() {
		return fmt.Errorf("hermes: invalid poll explanation media")
	}
	if params.Media != nil && !params.Media.validPollMedia() {
		return fmt.Errorf("hermes: invalid poll description media")
	}
	return nil
}

func (b *Client) SendPoll(ctx context.Context, params SendPollParams) (*Message, error) {
	if err := validateSendPoll(params); err != nil {
		return nil, err
	}
	return callMessage(ctx, b, "sendPoll", params)
}

// SendPollUpload sends a poll with one or more media files streamed from the
// supplied readers. Media values refer to uploads with Attachment(upload.Field).
func (b *Client) SendPollUpload(ctx context.Context, params SendPollParams, uploads ...Upload) (*Message, error) {
	if err := validateSendPoll(params); err != nil {
		return nil, err
	}
	if len(uploads) == 0 {
		return b.SendPoll(ctx, params)
	}
	fields, err := newFormFields(params.ChatID)
	if err != nil {
		return nil, err
	}
	fields.String("business_connection_id", params.BusinessConnectionID)
	fields.Int("message_thread_id", params.MessageThreadID)
	fields.String("question", params.Question)
	fields.String("question_parse_mode", params.QuestionParseMode)
	if len(params.QuestionEntities) != 0 {
		if err = fields.JSON("question_entities", params.QuestionEntities); err != nil {
			return nil, err
		}
	}
	if err = fields.JSON("options", params.Options); err != nil {
		return nil, err
	}
	fields.BoolPointer("is_anonymous", params.IsAnonymous)
	fields.String("type", params.Type)
	fields.Bool("allows_multiple_answers", params.AllowsMultipleAnswers)
	fields.BoolPointer("allows_revoting", params.AllowsRevoting)
	fields.Bool("shuffle_options", params.ShuffleOptions)
	fields.Bool("allow_adding_options", params.AllowAddingOptions)
	fields.Bool("hide_results_until_closes", params.HideResultsUntilCloses)
	fields.Bool("members_only", params.MembersOnly)
	if len(params.CountryCodes) != 0 {
		if err = fields.JSON("country_codes", params.CountryCodes); err != nil {
			return nil, err
		}
	}
	if len(params.CorrectOptionIDs) != 0 {
		if err = fields.JSON("correct_option_ids", params.CorrectOptionIDs); err != nil {
			return nil, err
		}
	}
	fields.String("explanation", params.Explanation)
	fields.String("explanation_parse_mode", params.ExplanationParseMode)
	if len(params.ExplanationEntities) != 0 {
		if err = fields.JSON("explanation_entities", params.ExplanationEntities); err != nil {
			return nil, err
		}
	}
	if params.ExplanationMedia != nil {
		if err = fields.JSON("explanation_media", params.ExplanationMedia); err != nil {
			return nil, err
		}
	}
	fields.Int("open_period", params.OpenPeriod)
	fields.Int64("close_date", params.CloseDate)
	fields.Bool("is_closed", params.IsClosed)
	fields.String("description", params.Description)
	fields.String("description_parse_mode", params.DescriptionParseMode)
	if len(params.DescriptionEntities) != 0 {
		if err = fields.JSON("description_entities", params.DescriptionEntities); err != nil {
			return nil, err
		}
	}
	if params.Media != nil {
		if err = fields.JSON("media", params.Media); err != nil {
			return nil, err
		}
	}
	fields.Bool("disable_notification", params.DisableNotification)
	fields.Bool("protect_content", params.ProtectContent)
	fields.Bool("allow_paid_broadcast", params.AllowPaidBroadcast)
	fields.String("message_effect_id", params.MessageEffectID)
	if params.ReplyParameters != nil {
		if err = fields.JSON("reply_parameters", params.ReplyParameters); err != nil {
			return nil, err
		}
	}
	if params.ReplyMarkup != nil {
		if err = fields.JSON("reply_markup", params.ReplyMarkup); err != nil {
			return nil, err
		}
	}
	if err = validateAttachmentUploads(params, uploads, "sendPoll"); err != nil {
		return nil, err
	}
	var message Message
	if err = b.CallMultipart(ctx, "sendPoll", fields, uploads, &message); err != nil {
		return nil, err
	}
	return &message, nil
}

type StopPollParams struct {
	BusinessConnectionID string                `json:"business_connection_id,omitempty"`
	ChatID               any                   `json:"chat_id"`
	MessageID            int                   `json:"message_id"`
	ReplyMarkup          *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (b *Client) StopPoll(ctx context.Context, params StopPollParams) (*Poll, error) {
	if err := validateChatID(params.ChatID, "stopPoll"); err != nil {
		return nil, err
	}
	if params.MessageID == 0 {
		return nil, fmt.Errorf("hermes: stopPoll message_id is required")
	}
	var poll Poll
	if err := b.Call(ctx, "stopPoll", params, &poll); err != nil {
		return nil, err
	}
	return &poll, nil
}

type SendDiceParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	Emoji                   string                   `json:"emoji,omitempty"`
	DisableNotification     bool                     `json:"disable_notification,omitempty"`
	ProtectContent          bool                     `json:"protect_content,omitempty"`
	AllowPaidBroadcast      bool                     `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID         string                   `json:"message_effect_id,omitempty"`
	SuggestedPostParameters *SuggestedPostParameters `json:"suggested_post_parameters,omitempty"`
	ReplyParameters         *ReplyParameters         `json:"reply_parameters,omitempty"`
	ReplyMarkup             ReplyMarkup              `json:"reply_markup,omitempty"`
}

func (b *Client) SendDice(ctx context.Context, params SendDiceParams) (*Message, error) {
	if err := validateChatID(params.ChatID, "sendDice"); err != nil {
		return nil, err
	}
	if params.Emoji != "" && !validDiceEmoji(params.Emoji) {
		return nil, fmt.Errorf("hermes: unsupported sendDice emoji %q", params.Emoji)
	}
	return callMessage(ctx, b, "sendDice", params)
}

func validDiceEmoji(emoji string) bool {
	switch emoji {
	case DiceEmoji, DartsEmoji, BasketballEmoji, FootballEmoji, BowlingEmoji, SlotsEmoji:
		return true
	default:
		return false
	}
}
