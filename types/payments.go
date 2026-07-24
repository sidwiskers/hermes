package types

// LabeledPrice is one component of an invoice or shipping price.
type LabeledPrice struct {
	Label  string `json:"label"`
	Amount int    `json:"amount"`
}

type Invoice struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	StartParameter string `json:"start_parameter"`
	Currency       string `json:"currency"`
	TotalAmount    int    `json:"total_amount"`
}

type ShippingOption struct {
	ID     string         `json:"id"`
	Title  string         `json:"title"`
	Prices []LabeledPrice `json:"prices"`
}

type SuccessfulPayment struct {
	Currency                   string     `json:"currency"`
	TotalAmount                int        `json:"total_amount"`
	InvoicePayload             string     `json:"invoice_payload"`
	SubscriptionExpirationDate int64      `json:"subscription_expiration_date,omitempty"`
	IsRecurring                bool       `json:"is_recurring,omitempty"`
	IsFirstRecurring           bool       `json:"is_first_recurring,omitempty"`
	ShippingOptionID           string     `json:"shipping_option_id,omitempty"`
	OrderInfo                  *OrderInfo `json:"order_info,omitempty"`
	TelegramPaymentChargeID    string     `json:"telegram_payment_charge_id"`
	ProviderPaymentChargeID    string     `json:"provider_payment_charge_id"`
}

type RefundedPayment struct {
	Currency                string `json:"currency"`
	TotalAmount             int    `json:"total_amount"`
	InvoicePayload          string `json:"invoice_payload"`
	TelegramPaymentChargeID string `json:"telegram_payment_charge_id"`
	ProviderPaymentChargeID string `json:"provider_payment_charge_id,omitempty"`
}

// PaidMedia is a compact discriminated response union. Type determines which
// media field is populated.
type PaidMedia struct {
	Type      string      `json:"type"`
	LivePhoto *LivePhoto  `json:"live_photo,omitempty"`
	Photo     []PhotoSize `json:"photo,omitempty"`
	Width     int         `json:"width,omitempty"`
	Height    int         `json:"height,omitempty"`
	Duration  int         `json:"duration,omitempty"`
	Video     *Video      `json:"video,omitempty"`
}

type PaidMediaInfo struct {
	StarCount int         `json:"star_count"`
	PaidMedia []PaidMedia `json:"paid_media"`
}

type PaidMediaPurchased struct {
	From             User   `json:"from"`
	PaidMediaPayload string `json:"paid_media_payload"`
}

type StarAmount struct {
	Amount         int `json:"amount"`
	NanostarAmount int `json:"nanostar_amount,omitempty"`
}

type RevenueWithdrawalState struct {
	Type string `json:"type"`
	Date int64  `json:"date,omitempty"`
	URL  string `json:"url,omitempty"`
}

type GiftBackground struct {
	CenterColor int `json:"center_color"`
	EdgeColor   int `json:"edge_color"`
	TextColor   int `json:"text_color"`
}

type Gift struct {
	ID                     string          `json:"id"`
	Sticker                Sticker         `json:"sticker"`
	StarCount              int             `json:"star_count"`
	UpgradeStarCount       int             `json:"upgrade_star_count,omitempty"`
	IsPremium              bool            `json:"is_premium,omitempty"`
	HasColors              bool            `json:"has_colors,omitempty"`
	TotalCount             int             `json:"total_count,omitempty"`
	RemainingCount         int             `json:"remaining_count,omitempty"`
	PersonalTotalCount     int             `json:"personal_total_count,omitempty"`
	PersonalRemainingCount int             `json:"personal_remaining_count,omitempty"`
	Background             *GiftBackground `json:"background,omitempty"`
	UniqueGiftVariantCount int             `json:"unique_gift_variant_count,omitempty"`
	PublisherChat          *Chat           `json:"publisher_chat,omitempty"`
}

type AffiliateInfo struct {
	AffiliateUser      *User `json:"affiliate_user,omitempty"`
	AffiliateChat      *Chat `json:"affiliate_chat,omitempty"`
	CommissionPerMille int   `json:"commission_per_mille"`
	Amount             int   `json:"amount"`
	NanostarAmount     int   `json:"nanostar_amount,omitempty"`
}

// TransactionPartner is a compact discriminated response union covering all
// seven Bot API 10.2 transaction partner forms.
type TransactionPartner struct {
	Type                        string                  `json:"type"`
	TransactionType             string                  `json:"transaction_type,omitempty"`
	User                        *User                   `json:"user,omitempty"`
	Chat                        *Chat                   `json:"chat,omitempty"`
	Affiliate                   *AffiliateInfo          `json:"affiliate,omitempty"`
	InvoicePayload              string                  `json:"invoice_payload,omitempty"`
	SubscriptionPeriod          int                     `json:"subscription_period,omitempty"`
	PaidMedia                   []PaidMedia             `json:"paid_media,omitempty"`
	PaidMediaPayload            string                  `json:"paid_media_payload,omitempty"`
	Gift                        *Gift                   `json:"gift,omitempty"`
	PremiumSubscriptionDuration int                     `json:"premium_subscription_duration,omitempty"`
	SponsorUser                 *User                   `json:"sponsor_user,omitempty"`
	CommissionPerMille          int                     `json:"commission_per_mille,omitempty"`
	WithdrawalState             *RevenueWithdrawalState `json:"withdrawal_state,omitempty"`
	RequestCount                int                     `json:"request_count,omitempty"`
}

type StarTransaction struct {
	ID             string              `json:"id"`
	Amount         int                 `json:"amount"`
	NanostarAmount int                 `json:"nanostar_amount,omitempty"`
	Date           int64               `json:"date"`
	Source         *TransactionPartner `json:"source,omitempty"`
	Receiver       *TransactionPartner `json:"receiver,omitempty"`
}

type StarTransactions struct {
	Transactions []StarTransaction `json:"transactions"`
}
