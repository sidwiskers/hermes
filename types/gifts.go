package types

import "encoding/json"

type Gifts struct {
	Gifts []Gift `json:"gifts"`
}

type UniqueGift struct {
	GiftID           string             `json:"gift_id"`
	BaseName         string             `json:"base_name"`
	Name             string             `json:"name"`
	Number           int                `json:"number"`
	Model            UniqueGiftModel    `json:"model"`
	Symbol           UniqueGiftSymbol   `json:"symbol"`
	Backdrop         UniqueGiftBackdrop `json:"backdrop"`
	IsPremium        bool               `json:"is_premium,omitempty"`
	IsBurned         bool               `json:"is_burned,omitempty"`
	IsFromBlockchain bool               `json:"is_from_blockchain,omitempty"`
	Colors           *UniqueGiftColors  `json:"colors,omitempty"`
	PublisherChat    *Chat              `json:"publisher_chat,omitempty"`
}

// OwnedGift is the compact union of regular and unique owned gifts.
type OwnedGift struct {
	Type                    string          `json:"type"`
	Gift                    *Gift           `json:"-"`
	UniqueGift              *UniqueGift     `json:"-"`
	OwnedGiftID             string          `json:"owned_gift_id,omitempty"`
	SenderUser              *User           `json:"sender_user,omitempty"`
	SendDate                int64           `json:"send_date"`
	Text                    string          `json:"text,omitempty"`
	Entities                []MessageEntity `json:"entities,omitempty"`
	IsPrivate               bool            `json:"is_private,omitempty"`
	IsSaved                 bool            `json:"is_saved,omitempty"`
	CanBeUpgraded           bool            `json:"can_be_upgraded,omitempty"`
	WasRefunded             bool            `json:"was_refunded,omitempty"`
	ConvertStarCount        int             `json:"convert_star_count,omitempty"`
	PrepaidUpgradeStarCount int             `json:"prepaid_upgrade_star_count,omitempty"`
	IsUpgradeSeparate       bool            `json:"is_upgrade_separate,omitempty"`
	UniqueGiftNumber        int             `json:"unique_gift_number,omitempty"`
	CanBeTransferred        bool            `json:"can_be_transferred,omitempty"`
	TransferStarCount       int             `json:"transfer_star_count,omitempty"`
	NextTransferDate        int64           `json:"next_transfer_date,omitempty"`
}

func (owned *OwnedGift) UnmarshalJSON(data []byte) error {
	type common OwnedGift
	var decoded struct {
		common
		Gift json.RawMessage `json:"gift"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*owned = OwnedGift(decoded.common)
	if len(decoded.Gift) == 0 || string(decoded.Gift) == "null" {
		return nil
	}
	switch owned.Type {
	case "regular":
		owned.Gift = new(Gift)
		return json.Unmarshal(decoded.Gift, owned.Gift)
	case "unique":
		owned.UniqueGift = new(UniqueGift)
		return json.Unmarshal(decoded.Gift, owned.UniqueGift)
	default:
		return nil
	}
}

type OwnedGifts struct {
	TotalCount int         `json:"total_count"`
	Gifts      []OwnedGift `json:"gifts"`
	NextOffset string      `json:"next_offset,omitempty"`
}
