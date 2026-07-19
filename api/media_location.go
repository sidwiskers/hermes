package api

import (
	"context"
	"fmt"
)

type SendContactParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	PhoneNumber             string                   `json:"phone_number"`
	FirstName               string                   `json:"first_name"`
	LastName                string                   `json:"last_name,omitempty"`
	VCard                   string                   `json:"vcard,omitempty"`
	DisableNotification     bool                     `json:"disable_notification,omitempty"`
	ProtectContent          bool                     `json:"protect_content,omitempty"`
	AllowPaidBroadcast      bool                     `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID         string                   `json:"message_effect_id,omitempty"`
	SuggestedPostParameters *SuggestedPostParameters `json:"suggested_post_parameters,omitempty"`
	ReplyParameters         *ReplyParameters         `json:"reply_parameters,omitempty"`
	ReplyMarkup             ReplyMarkup              `json:"reply_markup,omitempty"`
	ReceiverUserID          int64                    `json:"receiver_user_id,omitempty"`
	CallbackQueryID         string                   `json:"callback_query_id,omitempty"`
}

func (b *Client) SendContact(ctx context.Context, p SendContactParams) (*Message, error) {
	if err := validateChatID(p.ChatID, "sendContact"); err != nil {
		return nil, err
	}
	if p.PhoneNumber == "" || p.FirstName == "" {
		return nil, fmt.Errorf("hermes: sendContact phone_number and first_name are required")
	}
	return callMessage(ctx, b, "sendContact", p)
}

type SendLocationParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	Latitude                float64                  `json:"latitude"`
	Longitude               float64                  `json:"longitude"`
	HorizontalAccuracy      float64                  `json:"horizontal_accuracy,omitempty"`
	LivePeriod              int                      `json:"live_period,omitempty"`
	Heading                 int                      `json:"heading,omitempty"`
	ProximityAlertRadius    int                      `json:"proximity_alert_radius,omitempty"`
	DisableNotification     bool                     `json:"disable_notification,omitempty"`
	ProtectContent          bool                     `json:"protect_content,omitempty"`
	AllowPaidBroadcast      bool                     `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID         string                   `json:"message_effect_id,omitempty"`
	SuggestedPostParameters *SuggestedPostParameters `json:"suggested_post_parameters,omitempty"`
	ReplyParameters         *ReplyParameters         `json:"reply_parameters,omitempty"`
	ReplyMarkup             ReplyMarkup              `json:"reply_markup,omitempty"`
	ReceiverUserID          int64                    `json:"receiver_user_id,omitempty"`
	CallbackQueryID         string                   `json:"callback_query_id,omitempty"`
}

func (b *Client) SendLocation(ctx context.Context, p SendLocationParams) (*Message, error) {
	if err := validateChatID(p.ChatID, "sendLocation"); err != nil {
		return nil, err
	}
	return callMessage(ctx, b, "sendLocation", p)
}

type SendVenueParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	Latitude                float64                  `json:"latitude"`
	Longitude               float64                  `json:"longitude"`
	Title                   string                   `json:"title"`
	Address                 string                   `json:"address"`
	FoursquareID            string                   `json:"foursquare_id,omitempty"`
	FoursquareType          string                   `json:"foursquare_type,omitempty"`
	GooglePlaceID           string                   `json:"google_place_id,omitempty"`
	GooglePlaceType         string                   `json:"google_place_type,omitempty"`
	DisableNotification     bool                     `json:"disable_notification,omitempty"`
	ProtectContent          bool                     `json:"protect_content,omitempty"`
	AllowPaidBroadcast      bool                     `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID         string                   `json:"message_effect_id,omitempty"`
	SuggestedPostParameters *SuggestedPostParameters `json:"suggested_post_parameters,omitempty"`
	ReplyParameters         *ReplyParameters         `json:"reply_parameters,omitempty"`
	ReplyMarkup             ReplyMarkup              `json:"reply_markup,omitempty"`
	ReceiverUserID          int64                    `json:"receiver_user_id,omitempty"`
	CallbackQueryID         string                   `json:"callback_query_id,omitempty"`
}

func (b *Client) SendVenue(ctx context.Context, p SendVenueParams) (*Message, error) {
	if err := validateChatID(p.ChatID, "sendVenue"); err != nil {
		return nil, err
	}
	if p.Title == "" || p.Address == "" {
		return nil, fmt.Errorf("hermes: sendVenue title and address are required")
	}
	return callMessage(ctx, b, "sendVenue", p)
}
