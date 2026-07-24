package types

type BotAccessSettings struct {
	IsAccessRestricted bool   `json:"is_access_restricted"`
	AddedUsers         []User `json:"added_users,omitempty"`
}

type ManagedBotCreated struct {
	Bot User `json:"bot"`
}

type ManagedBotUpdated struct {
	User User `json:"user"`
	Bot  User `json:"bot"`
}

type AcceptedGiftTypes struct {
	UnlimitedGifts      bool `json:"unlimited_gifts"`
	LimitedGifts        bool `json:"limited_gifts"`
	UniqueGifts         bool `json:"unique_gifts"`
	PremiumSubscription bool `json:"premium_subscription"`
	GiftsFromChannels   bool `json:"gifts_from_channels"`
}
