package types

// BotName is the localized display name returned by getMyName.
type BotName struct {
	Name string `json:"name"`
}

// BotDescription is the localized long description returned by getMyDescription.
type BotDescription struct {
	Description string `json:"description"`
}

// BotShortDescription is the localized profile description returned by
// getMyShortDescription.
type BotShortDescription struct {
	ShortDescription string `json:"short_description"`
}

// UserProfilePhotos contains the requested page of a user's profile photos.
type UserProfilePhotos struct {
	TotalCount int           `json:"total_count"`
	Photos     [][]PhotoSize `json:"photos"`
}

// UserProfileAudios contains the requested page of a user's profile audios.
type UserProfileAudios struct {
	TotalCount int     `json:"total_count"`
	Audios     []Audio `json:"audios"`
}

// ChatAdministratorRights describes the privileges suggested when a bot is
// added as a chat administrator.
type ChatAdministratorRights struct {
	IsAnonymous             bool `json:"is_anonymous"`
	CanManageChat           bool `json:"can_manage_chat"`
	CanDeleteMessages       bool `json:"can_delete_messages"`
	CanManageVideoChats     bool `json:"can_manage_video_chats"`
	CanRestrictMembers      bool `json:"can_restrict_members"`
	CanPromoteMembers       bool `json:"can_promote_members"`
	CanChangeInfo           bool `json:"can_change_info"`
	CanInviteUsers          bool `json:"can_invite_users"`
	CanPostStories          bool `json:"can_post_stories"`
	CanEditStories          bool `json:"can_edit_stories"`
	CanDeleteStories        bool `json:"can_delete_stories"`
	CanPostMessages         bool `json:"can_post_messages,omitempty"`
	CanEditMessages         bool `json:"can_edit_messages,omitempty"`
	CanPinMessages          bool `json:"can_pin_messages,omitempty"`
	CanManageTopics         bool `json:"can_manage_topics,omitempty"`
	CanManageDirectMessages bool `json:"can_manage_direct_messages,omitempty"`
	CanManageTags           bool `json:"can_manage_tags,omitempty"`
}

const (
	MenuButtonTypeCommands = "commands"
	MenuButtonTypeWebApp   = "web_app"
	MenuButtonTypeDefault  = "default"
)

// MenuButton is the compact discriminated union returned by getChatMenuButton.
// Type is one of MenuButtonTypeCommands, MenuButtonTypeWebApp, or
// MenuButtonTypeDefault.
type MenuButton struct {
	Type   string      `json:"type"`
	Text   string      `json:"text,omitempty"`
	WebApp *WebAppInfo `json:"web_app,omitempty"`
}

func CommandsMenuButton() MenuButton { return MenuButton{Type: MenuButtonTypeCommands} }

func WebAppMenuButton(text, url string) MenuButton {
	return MenuButton{Type: MenuButtonTypeWebApp, Text: text, WebApp: &WebAppInfo{URL: url}}
}

func DefaultMenuButton() MenuButton { return MenuButton{Type: MenuButtonTypeDefault} }

// BusinessBotRights represents the permissions granted through a business
// connection.
type BusinessBotRights struct {
	CanReply                   bool `json:"can_reply,omitempty"`
	CanReadMessages            bool `json:"can_read_messages,omitempty"`
	CanDeleteSentMessages      bool `json:"can_delete_sent_messages,omitempty"`
	CanDeleteAllMessages       bool `json:"can_delete_all_messages,omitempty"`
	CanEditName                bool `json:"can_edit_name,omitempty"`
	CanEditBio                 bool `json:"can_edit_bio,omitempty"`
	CanEditProfilePhoto        bool `json:"can_edit_profile_photo,omitempty"`
	CanEditUsername            bool `json:"can_edit_username,omitempty"`
	CanChangeGiftSettings      bool `json:"can_change_gift_settings,omitempty"`
	CanViewGiftsAndStars       bool `json:"can_view_gifts_and_stars,omitempty"`
	CanConvertGiftsToStars     bool `json:"can_convert_gifts_to_stars,omitempty"`
	CanTransferAndUpgradeGifts bool `json:"can_transfer_and_upgrade_gifts,omitempty"`
	CanTransferStars           bool `json:"can_transfer_stars,omitempty"`
	CanManageStories           bool `json:"can_manage_stories,omitempty"`
}

// ChatBoostSource is the compact union of premium, gift-code, and giveaway
// boost sources. Fields not used by Source are omitted.
type ChatBoostSource struct {
	Source            string `json:"source"`
	User              *User  `json:"user,omitempty"`
	GiveawayMessageID int    `json:"giveaway_message_id,omitempty"`
	PrizeStarCount    int    `json:"prize_star_count,omitempty"`
	IsUnclaimed       bool   `json:"is_unclaimed,omitempty"`
}

type ChatBoost struct {
	BoostID        string          `json:"boost_id"`
	AddDate        int64           `json:"add_date"`
	ExpirationDate int64           `json:"expiration_date"`
	Source         ChatBoostSource `json:"source"`
}

type UserChatBoosts struct {
	Boosts []ChatBoost `json:"boosts"`
}

type ChatBoostUpdated struct {
	Chat  Chat      `json:"chat"`
	Boost ChatBoost `json:"boost"`
}

type ChatBoostRemoved struct {
	Chat       Chat            `json:"chat"`
	BoostID    string          `json:"boost_id"`
	RemoveDate int64           `json:"remove_date"`
	Source     ChatBoostSource `json:"source"`
}
