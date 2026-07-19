package api

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

type SendInvoiceParams struct {
	ChatID                    any                      `json:"chat_id"`
	MessageThreadID           int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID     int                      `json:"direct_messages_topic_id,omitempty"`
	Title                     string                   `json:"title"`
	Description               string                   `json:"description"`
	Payload                   string                   `json:"payload"`
	ProviderToken             string                   `json:"provider_token,omitempty"`
	Currency                  string                   `json:"currency"`
	Prices                    []LabeledPrice           `json:"prices"`
	MaxTipAmount              int                      `json:"max_tip_amount,omitempty"`
	SuggestedTipAmounts       []int                    `json:"suggested_tip_amounts,omitempty"`
	StartParameter            string                   `json:"start_parameter,omitempty"`
	ProviderData              string                   `json:"provider_data,omitempty"`
	PhotoURL                  string                   `json:"photo_url,omitempty"`
	PhotoSize                 int                      `json:"photo_size,omitempty"`
	PhotoWidth                int                      `json:"photo_width,omitempty"`
	PhotoHeight               int                      `json:"photo_height,omitempty"`
	NeedName                  bool                     `json:"need_name,omitempty"`
	NeedPhoneNumber           bool                     `json:"need_phone_number,omitempty"`
	NeedEmail                 bool                     `json:"need_email,omitempty"`
	NeedShippingAddress       bool                     `json:"need_shipping_address,omitempty"`
	SendPhoneNumberToProvider bool                     `json:"send_phone_number_to_provider,omitempty"`
	SendEmailToProvider       bool                     `json:"send_email_to_provider,omitempty"`
	IsFlexible                bool                     `json:"is_flexible,omitempty"`
	DisableNotification       bool                     `json:"disable_notification,omitempty"`
	ProtectContent            bool                     `json:"protect_content,omitempty"`
	AllowPaidBroadcast        bool                     `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID           string                   `json:"message_effect_id,omitempty"`
	SuggestedPostParameters   *SuggestedPostParameters `json:"suggested_post_parameters,omitempty"`
	ReplyParameters           *ReplyParameters         `json:"reply_parameters,omitempty"`
	ReplyMarkup               *InlineKeyboardMarkup    `json:"reply_markup,omitempty"`
}

func validateInvoice(
	title string,
	description string,
	payload string,
	currency string,
	prices []LabeledPrice,
	maxTip int,
	suggestedTips []int,
) error {
	titleLength := utf8.RuneCountInString(title)
	if titleLength < 1 || titleLength > 32 {
		return fmt.Errorf("hermes: invoice title must contain 1-32 characters")
	}
	descriptionLength := utf8.RuneCountInString(description)
	if descriptionLength < 1 || descriptionLength > 255 {
		return fmt.Errorf("hermes: invoice description must contain 1-255 characters")
	}
	if len(payload) < 1 || len(payload) > 128 {
		return fmt.Errorf("hermes: invoice payload must contain 1-128 bytes")
	}
	if len(currency) != 3 {
		return fmt.Errorf("hermes: invoice currency must be a three-letter code")
	}
	if len(prices) == 0 {
		return fmt.Errorf("hermes: invoice prices are required")
	}
	for index, price := range prices {
		if strings.TrimSpace(price.Label) == "" {
			return fmt.Errorf("hermes: invoice price %d has no label", index)
		}
	}
	if len(suggestedTips) > 4 {
		return fmt.Errorf("hermes: invoice accepts at most four suggested tips")
	}
	previous := 0
	for _, tip := range suggestedTips {
		if tip <= previous || tip > maxTip {
			return fmt.Errorf("hermes: invoice suggested tips must be positive, increasing, and at most max_tip_amount")
		}
		previous = tip
	}
	return nil
}

func (client *Client) SendInvoice(ctx context.Context, params SendInvoiceParams) (*Message, error) {
	if err := validateChatID(params.ChatID, "sendInvoice"); err != nil {
		return nil, err
	}
	if err := validateInvoice(params.Title, params.Description, params.Payload, params.Currency, params.Prices, params.MaxTipAmount, params.SuggestedTipAmounts); err != nil {
		return nil, err
	}
	return callMessage(ctx, client, "sendInvoice", params)
}

type CreateInvoiceLinkParams struct {
	BusinessConnectionID      string         `json:"business_connection_id,omitempty"`
	Title                     string         `json:"title"`
	Description               string         `json:"description"`
	Payload                   string         `json:"payload"`
	ProviderToken             string         `json:"provider_token,omitempty"`
	Currency                  string         `json:"currency"`
	Prices                    []LabeledPrice `json:"prices"`
	SubscriptionPeriod        int            `json:"subscription_period,omitempty"`
	MaxTipAmount              int            `json:"max_tip_amount,omitempty"`
	SuggestedTipAmounts       []int          `json:"suggested_tip_amounts,omitempty"`
	ProviderData              string         `json:"provider_data,omitempty"`
	PhotoURL                  string         `json:"photo_url,omitempty"`
	PhotoSize                 int            `json:"photo_size,omitempty"`
	PhotoWidth                int            `json:"photo_width,omitempty"`
	PhotoHeight               int            `json:"photo_height,omitempty"`
	NeedName                  bool           `json:"need_name,omitempty"`
	NeedPhoneNumber           bool           `json:"need_phone_number,omitempty"`
	NeedEmail                 bool           `json:"need_email,omitempty"`
	NeedShippingAddress       bool           `json:"need_shipping_address,omitempty"`
	SendPhoneNumberToProvider bool           `json:"send_phone_number_to_provider,omitempty"`
	SendEmailToProvider       bool           `json:"send_email_to_provider,omitempty"`
	IsFlexible                bool           `json:"is_flexible,omitempty"`
}

func (client *Client) CreateInvoiceLink(ctx context.Context, params CreateInvoiceLinkParams) (string, error) {
	if err := validateInvoice(params.Title, params.Description, params.Payload, params.Currency, params.Prices, params.MaxTipAmount, params.SuggestedTipAmounts); err != nil {
		return "", err
	}
	if params.SubscriptionPeriod != 0 && (params.Currency != "XTR" || params.SubscriptionPeriod != 2_592_000) {
		return "", fmt.Errorf("hermes: invoice subscriptions require XTR and a 2592000-second period")
	}
	var link string
	if err := client.Call(ctx, "createInvoiceLink", params, &link); err != nil {
		return "", err
	}
	return link, nil
}

type AnswerShippingQueryParams struct {
	ShippingQueryID string           `json:"shipping_query_id"`
	OK              bool             `json:"ok"`
	ShippingOptions []ShippingOption `json:"shipping_options,omitempty"`
	ErrorMessage    string           `json:"error_message,omitempty"`
}

func (client *Client) AnswerShippingQuery(ctx context.Context, params AnswerShippingQueryParams) error {
	if strings.TrimSpace(params.ShippingQueryID) == "" {
		return fmt.Errorf("hermes: answerShippingQuery shipping_query_id is required")
	}
	if params.OK && len(params.ShippingOptions) == 0 {
		return fmt.Errorf("hermes: answerShippingQuery requires shipping_options when ok is true")
	}
	if !params.OK && strings.TrimSpace(params.ErrorMessage) == "" {
		return fmt.Errorf("hermes: answerShippingQuery requires error_message when ok is false")
	}
	return client.callTrue(ctx, "answerShippingQuery", params)
}

type AnswerPreCheckoutQueryParams struct {
	PreCheckoutQueryID string `json:"pre_checkout_query_id"`
	OK                 bool   `json:"ok"`
	ErrorMessage       string `json:"error_message,omitempty"`
}

func (client *Client) AnswerPreCheckoutQuery(ctx context.Context, params AnswerPreCheckoutQueryParams) error {
	if strings.TrimSpace(params.PreCheckoutQueryID) == "" {
		return fmt.Errorf("hermes: answerPreCheckoutQuery pre_checkout_query_id is required")
	}
	if !params.OK && strings.TrimSpace(params.ErrorMessage) == "" {
		return fmt.Errorf("hermes: answerPreCheckoutQuery requires error_message when ok is false")
	}
	return client.callTrue(ctx, "answerPreCheckoutQuery", params)
}

func (client *Client) GetMyStarBalance(ctx context.Context) (StarAmount, error) {
	return Call[StarAmount](ctx, client, "getMyStarBalance", nil)
}

type GetStarTransactionsParams struct {
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit,omitempty"`
}

func (client *Client) GetStarTransactions(ctx context.Context, params GetStarTransactionsParams) (StarTransactions, error) {
	if params.Limit < 0 || params.Limit > 100 {
		return StarTransactions{}, fmt.Errorf("hermes: getStarTransactions limit must be 1-100 or zero")
	}
	return Call[StarTransactions](ctx, client, "getStarTransactions", params)
}

type RefundStarPaymentParams struct {
	UserID                  int64  `json:"user_id"`
	TelegramPaymentChargeID string `json:"telegram_payment_charge_id"`
}

func (client *Client) RefundStarPayment(ctx context.Context, params RefundStarPaymentParams) error {
	if params.UserID == 0 || strings.TrimSpace(params.TelegramPaymentChargeID) == "" {
		return fmt.Errorf("hermes: refundStarPayment requires user_id and telegram_payment_charge_id")
	}
	return client.callTrue(ctx, "refundStarPayment", params)
}

type EditUserStarSubscriptionParams struct {
	UserID                  int64  `json:"user_id"`
	TelegramPaymentChargeID string `json:"telegram_payment_charge_id"`
	IsCanceled              bool   `json:"is_canceled"`
}

func (client *Client) EditUserStarSubscription(ctx context.Context, params EditUserStarSubscriptionParams) error {
	if params.UserID == 0 || strings.TrimSpace(params.TelegramPaymentChargeID) == "" {
		return fmt.Errorf("hermes: editUserStarSubscription requires user_id and telegram_payment_charge_id")
	}
	return client.callTrue(ctx, "editUserStarSubscription", params)
}

type InputPaidMedia interface {
	inputPaidMedia()
	paidMediaSource() string
}

type InputPaidMediaLivePhoto struct {
	Media string `json:"media"`
	Photo string `json:"photo"`
}

func (InputPaidMediaLivePhoto) inputPaidMedia()               {}
func (value InputPaidMediaLivePhoto) paidMediaSource() string { return value.Media }
func (value InputPaidMediaLivePhoto) MarshalJSON() ([]byte, error) {
	type plain InputPaidMediaLivePhoto
	return marshalTaggedObject("live_photo", plain(value))
}

type InputPaidMediaPhoto struct {
	Media string `json:"media"`
}

func (InputPaidMediaPhoto) inputPaidMedia()               {}
func (value InputPaidMediaPhoto) paidMediaSource() string { return value.Media }
func (value InputPaidMediaPhoto) MarshalJSON() ([]byte, error) {
	type plain InputPaidMediaPhoto
	return marshalTaggedObject("photo", plain(value))
}

type InputPaidMediaVideo struct {
	Media             string `json:"media"`
	Thumbnail         string `json:"thumbnail,omitempty"`
	Cover             string `json:"cover,omitempty"`
	StartTimestamp    int    `json:"start_timestamp,omitempty"`
	Width             int    `json:"width,omitempty"`
	Height            int    `json:"height,omitempty"`
	Duration          int    `json:"duration,omitempty"`
	SupportsStreaming bool   `json:"supports_streaming,omitempty"`
}

func (InputPaidMediaVideo) inputPaidMedia()               {}
func (value InputPaidMediaVideo) paidMediaSource() string { return value.Media }
func (value InputPaidMediaVideo) MarshalJSON() ([]byte, error) {
	type plain InputPaidMediaVideo
	return marshalTaggedObject("video", plain(value))
}

type SendPaidMediaParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	StarCount               int                      `json:"star_count"`
	Media                   []InputPaidMedia         `json:"media"`
	Payload                 string                   `json:"payload,omitempty"`
	Caption                 string                   `json:"caption,omitempty"`
	ParseMode               string                   `json:"parse_mode,omitempty"`
	CaptionEntities         []MessageEntity          `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia   bool                     `json:"show_caption_above_media,omitempty"`
	DisableNotification     bool                     `json:"disable_notification,omitempty"`
	ProtectContent          bool                     `json:"protect_content,omitempty"`
	AllowPaidBroadcast      bool                     `json:"allow_paid_broadcast,omitempty"`
	SuggestedPostParameters *SuggestedPostParameters `json:"suggested_post_parameters,omitempty"`
	ReplyParameters         *ReplyParameters         `json:"reply_parameters,omitempty"`
	ReplyMarkup             ReplyMarkup              `json:"reply_markup,omitempty"`
}

func validateSendPaidMedia(params SendPaidMediaParams, uploads []Upload) error {
	if err := validateChatID(params.ChatID, "sendPaidMedia"); err != nil {
		return err
	}
	if params.StarCount < 1 || params.StarCount > 25_000 {
		return fmt.Errorf("hermes: sendPaidMedia star_count must be 1-25000")
	}
	if len(params.Media) < 1 || len(params.Media) > 10 {
		return fmt.Errorf("hermes: sendPaidMedia requires 1-10 media items")
	}
	for index, item := range params.Media {
		if item == nil || strings.TrimSpace(item.paidMediaSource()) == "" {
			return fmt.Errorf("hermes: sendPaidMedia item %d has no media", index)
		}
		if live, ok := item.(InputPaidMediaLivePhoto); ok && strings.TrimSpace(live.Photo) == "" {
			return fmt.Errorf("hermes: sendPaidMedia live photo %d has no static photo", index)
		}
	}
	if len(params.Payload) > 128 {
		return fmt.Errorf("hermes: sendPaidMedia payload must not exceed 128 bytes")
	}
	return validateAttachmentUploads(params.Media, uploads, "sendPaidMedia")
}

func (client *Client) SendPaidMedia(ctx context.Context, params SendPaidMediaParams) (*Message, error) {
	if err := validateSendPaidMedia(params, nil); err != nil {
		return nil, err
	}
	return callMessage(ctx, client, "sendPaidMedia", params)
}

func (client *Client) SendPaidMediaUpload(
	ctx context.Context,
	params SendPaidMediaParams,
	uploads ...Upload,
) (*Message, error) {
	if len(uploads) == 0 {
		return client.SendPaidMedia(ctx, params)
	}
	if err := validateSendPaidMedia(params, uploads); err != nil {
		return nil, err
	}
	fields, err := newFormFields(params.ChatID)
	if err != nil {
		return nil, err
	}
	fields.String("business_connection_id", params.BusinessConnectionID)
	fields.Int("message_thread_id", params.MessageThreadID)
	fields.Int("direct_messages_topic_id", params.DirectMessagesTopicID)
	fields.Int("star_count", params.StarCount)
	if err = fields.JSON("media", params.Media); err != nil {
		return nil, err
	}
	fields.String("payload", params.Payload)
	fields.String("caption", params.Caption)
	fields.String("parse_mode", params.ParseMode)
	if len(params.CaptionEntities) != 0 {
		if err = fields.JSON("caption_entities", params.CaptionEntities); err != nil {
			return nil, err
		}
	}
	fields.Bool("show_caption_above_media", params.ShowCaptionAboveMedia)
	fields.Bool("disable_notification", params.DisableNotification)
	fields.Bool("protect_content", params.ProtectContent)
	fields.Bool("allow_paid_broadcast", params.AllowPaidBroadcast)
	if params.SuggestedPostParameters != nil {
		if err = fields.JSON("suggested_post_parameters", params.SuggestedPostParameters); err != nil {
			return nil, err
		}
	}
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
	var message Message
	if err = client.CallMultipart(ctx, "sendPaidMedia", fields, uploads, &message); err != nil {
		return nil, err
	}
	return &message, nil
}
