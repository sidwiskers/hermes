package api

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

func (client *Client) GetAvailableGifts(ctx context.Context) (Gifts, error) {
	return Call[Gifts](ctx, client, "getAvailableGifts", nil)
}

type SendGiftParams struct {
	UserID        int64           `json:"user_id,omitempty"`
	ChatID        any             `json:"chat_id,omitempty"`
	GiftID        string          `json:"gift_id"`
	PayForUpgrade bool            `json:"pay_for_upgrade,omitempty"`
	Text          string          `json:"text,omitempty"`
	TextParseMode string          `json:"text_parse_mode,omitempty"`
	TextEntities  []MessageEntity `json:"text_entities,omitempty"`
}

func (client *Client) SendGift(ctx context.Context, params SendGiftParams) error {
	hasUser := params.UserID != 0
	hasChat := params.ChatID != nil
	if hasUser == hasChat {
		return fmt.Errorf("hermes: sendGift requires exactly one of user_id or chat_id")
	}
	if hasChat {
		if err := validateChatID(params.ChatID, "sendGift"); err != nil {
			return err
		}
	}
	if strings.TrimSpace(params.GiftID) == "" {
		return fmt.Errorf("hermes: sendGift gift_id is required")
	}
	if utf8.RuneCountInString(params.Text) > 128 {
		return fmt.Errorf("hermes: sendGift text must not exceed 128 characters")
	}
	return client.callTrue(ctx, "sendGift", params)
}

type GiftPremiumSubscriptionParams struct {
	UserID        int64           `json:"user_id"`
	MonthCount    int             `json:"month_count"`
	StarCount     int             `json:"star_count"`
	Text          string          `json:"text,omitempty"`
	TextParseMode string          `json:"text_parse_mode,omitempty"`
	TextEntities  []MessageEntity `json:"text_entities,omitempty"`
}

func (client *Client) GiftPremiumSubscription(ctx context.Context, params GiftPremiumSubscriptionParams) error {
	if params.UserID == 0 {
		return fmt.Errorf("hermes: giftPremiumSubscription user_id is required")
	}
	expectedStars := 0
	switch params.MonthCount {
	case 3:
		expectedStars = 1000
	case 6:
		expectedStars = 1500
	case 12:
		expectedStars = 2500
	default:
		return fmt.Errorf("hermes: giftPremiumSubscription month_count must be 3, 6, or 12")
	}
	if params.StarCount != expectedStars {
		return fmt.Errorf("hermes: giftPremiumSubscription star_count must be %d for %d months", expectedStars, params.MonthCount)
	}
	if utf8.RuneCountInString(params.Text) > 128 {
		return fmt.Errorf("hermes: giftPremiumSubscription text must not exceed 128 characters")
	}
	return client.callTrue(ctx, "giftPremiumSubscription", params)
}

type OwnedGiftsFilter struct {
	ExcludeUnsaved              bool   `json:"exclude_unsaved,omitempty"`
	ExcludeSaved                bool   `json:"exclude_saved,omitempty"`
	ExcludeUnlimited            bool   `json:"exclude_unlimited,omitempty"`
	ExcludeLimitedUpgradable    bool   `json:"exclude_limited_upgradable,omitempty"`
	ExcludeLimitedNonUpgradable bool   `json:"exclude_limited_non_upgradable,omitempty"`
	ExcludeUnique               bool   `json:"exclude_unique,omitempty"`
	ExcludeFromBlockchain       bool   `json:"exclude_from_blockchain,omitempty"`
	SortByPrice                 bool   `json:"sort_by_price,omitempty"`
	Offset                      string `json:"offset,omitempty"`
	Limit                       int    `json:"limit,omitempty"`
}

func validateOwnedGiftsLimit(limit int, method string) error {
	if limit < 0 || limit > 100 {
		return fmt.Errorf("hermes: %s limit must be between 1 and 100 when set", method)
	}
	return nil
}

type GetBusinessAccountGiftsParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
	OwnedGiftsFilter
}

func (client *Client) GetBusinessAccountGifts(ctx context.Context, params GetBusinessAccountGiftsParams) (OwnedGifts, error) {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "getBusinessAccountGifts"); err != nil {
		return OwnedGifts{}, err
	}
	if err := validateOwnedGiftsLimit(params.Limit, "getBusinessAccountGifts"); err != nil {
		return OwnedGifts{}, err
	}
	return Call[OwnedGifts](ctx, client, "getBusinessAccountGifts", params)
}

type GetUserGiftsParams struct {
	UserID int64 `json:"user_id"`
	OwnedGiftsFilter
}

func (client *Client) GetUserGifts(ctx context.Context, params GetUserGiftsParams) (OwnedGifts, error) {
	if params.UserID == 0 {
		return OwnedGifts{}, fmt.Errorf("hermes: getUserGifts user_id is required")
	}
	if err := validateOwnedGiftsLimit(params.Limit, "getUserGifts"); err != nil {
		return OwnedGifts{}, err
	}
	return Call[OwnedGifts](ctx, client, "getUserGifts", params)
}

type GetChatGiftsParams struct {
	ChatID any `json:"chat_id"`
	OwnedGiftsFilter
}

func (client *Client) GetChatGifts(ctx context.Context, params GetChatGiftsParams) (OwnedGifts, error) {
	if err := validateChatID(params.ChatID, "getChatGifts"); err != nil {
		return OwnedGifts{}, err
	}
	if err := validateOwnedGiftsLimit(params.Limit, "getChatGifts"); err != nil {
		return OwnedGifts{}, err
	}
	return Call[OwnedGifts](ctx, client, "getChatGifts", params)
}

type OwnedGiftParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
	OwnedGiftID          string `json:"owned_gift_id"`
}

func validateOwnedGift(params OwnedGiftParams, method string) error {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, method); err != nil {
		return err
	}
	if strings.TrimSpace(params.OwnedGiftID) == "" {
		return fmt.Errorf("hermes: %s owned_gift_id is required", method)
	}
	return nil
}

func (client *Client) ConvertGiftToStars(ctx context.Context, params OwnedGiftParams) error {
	if err := validateOwnedGift(params, "convertGiftToStars"); err != nil {
		return err
	}
	return client.callTrue(ctx, "convertGiftToStars", params)
}

type UpgradeGiftParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
	OwnedGiftID          string `json:"owned_gift_id"`
	KeepOriginalDetails  bool   `json:"keep_original_details,omitempty"`
	StarCount            int    `json:"star_count,omitempty"`
}

func (client *Client) UpgradeGift(ctx context.Context, params UpgradeGiftParams) error {
	if err := validateOwnedGift(OwnedGiftParams{BusinessConnectionID: params.BusinessConnectionID, OwnedGiftID: params.OwnedGiftID}, "upgradeGift"); err != nil {
		return err
	}
	if params.StarCount < 0 {
		return fmt.Errorf("hermes: upgradeGift star_count must not be negative")
	}
	return client.callTrue(ctx, "upgradeGift", params)
}

type TransferGiftParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
	OwnedGiftID          string `json:"owned_gift_id"`
	NewOwnerChatID       int64  `json:"new_owner_chat_id"`
	StarCount            int    `json:"star_count,omitempty"`
}

func (client *Client) TransferGift(ctx context.Context, params TransferGiftParams) error {
	if err := validateOwnedGift(OwnedGiftParams{BusinessConnectionID: params.BusinessConnectionID, OwnedGiftID: params.OwnedGiftID}, "transferGift"); err != nil {
		return err
	}
	if params.NewOwnerChatID == 0 || params.StarCount < 0 {
		return fmt.Errorf("hermes: transferGift requires new_owner_chat_id and a non-negative star_count")
	}
	return client.callTrue(ctx, "transferGift", params)
}
