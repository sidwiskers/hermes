package types

// ChatPermissions describes actions a non-administrator may perform in a chat.
type ChatPermissions struct {
	CanSendMessages       bool `json:"can_send_messages,omitempty"`
	CanSendAudios         bool `json:"can_send_audios,omitempty"`
	CanSendDocuments      bool `json:"can_send_documents,omitempty"`
	CanSendPhotos         bool `json:"can_send_photos,omitempty"`
	CanSendVideos         bool `json:"can_send_videos,omitempty"`
	CanSendVideoNotes     bool `json:"can_send_video_notes,omitempty"`
	CanSendVoiceNotes     bool `json:"can_send_voice_notes,omitempty"`
	CanSendPolls          bool `json:"can_send_polls,omitempty"`
	CanSendOtherMessages  bool `json:"can_send_other_messages,omitempty"`
	CanAddWebPagePreviews bool `json:"can_add_web_page_previews,omitempty"`
	CanChangeInfo         bool `json:"can_change_info,omitempty"`
	CanInviteUsers        bool `json:"can_invite_users,omitempty"`
	CanPinMessages        bool `json:"can_pin_messages,omitempty"`
	CanManageTopics       bool `json:"can_manage_topics,omitempty"`
	CanReactToMessages    bool `json:"can_react_to_messages,omitempty"`
	CanEditTag            bool `json:"can_edit_tag,omitempty"`
}

// ChatMember is a compact union of Telegram's chat-member variants. Fields
// that don't apply to the current Status are omitted by Telegram.
type ChatMember struct {
	Status string `json:"status"`
	User   User   `json:"user"`

	IsAnonymous bool   `json:"is_anonymous,omitempty"`
	CustomTitle string `json:"custom_title,omitempty"`
	Tag         string `json:"tag,omitempty"`
	CanBeEdited bool   `json:"can_be_edited,omitempty"`
	IsMember    bool   `json:"is_member,omitempty"`
	UntilDate   int64  `json:"until_date,omitempty"`

	CanManageChat           bool `json:"can_manage_chat,omitempty"`
	CanDeleteMessages       bool `json:"can_delete_messages,omitempty"`
	CanManageVideoChats     bool `json:"can_manage_video_chats,omitempty"`
	CanRestrictMembers      bool `json:"can_restrict_members,omitempty"`
	CanPromoteMembers       bool `json:"can_promote_members,omitempty"`
	CanChangeInfo           bool `json:"can_change_info,omitempty"`
	CanInviteUsers          bool `json:"can_invite_users,omitempty"`
	CanPostStories          bool `json:"can_post_stories,omitempty"`
	CanEditStories          bool `json:"can_edit_stories,omitempty"`
	CanDeleteStories        bool `json:"can_delete_stories,omitempty"`
	CanPostMessages         bool `json:"can_post_messages,omitempty"`
	CanEditMessages         bool `json:"can_edit_messages,omitempty"`
	CanPinMessages          bool `json:"can_pin_messages,omitempty"`
	CanManageTopics         bool `json:"can_manage_topics,omitempty"`
	CanManageDirectMessages bool `json:"can_manage_direct_messages,omitempty"`
	CanManageTags           bool `json:"can_manage_tags,omitempty"`

	CanSendMessages       bool `json:"can_send_messages,omitempty"`
	CanSendAudios         bool `json:"can_send_audios,omitempty"`
	CanSendDocuments      bool `json:"can_send_documents,omitempty"`
	CanSendPhotos         bool `json:"can_send_photos,omitempty"`
	CanSendVideos         bool `json:"can_send_videos,omitempty"`
	CanSendVideoNotes     bool `json:"can_send_video_notes,omitempty"`
	CanSendVoiceNotes     bool `json:"can_send_voice_notes,omitempty"`
	CanSendPolls          bool `json:"can_send_polls,omitempty"`
	CanSendOtherMessages  bool `json:"can_send_other_messages,omitempty"`
	CanAddWebPagePreviews bool `json:"can_add_web_page_previews,omitempty"`
	CanReactToMessages    bool `json:"can_react_to_messages,omitempty"`
	CanEditTag            bool `json:"can_edit_tag,omitempty"`
}

func (m ChatMember) IsAdministrator() bool {
	return m.Status == "creator" || m.Status == "administrator"
}

func (m ChatMember) IsRestricted() bool { return m.Status == "restricted" }
func (m ChatMember) IsBanned() bool     { return m.Status == "kicked" }

// ChatInviteLink describes an invite link created by an administrator.
type ChatInviteLink struct {
	InviteLink              string `json:"invite_link"`
	Creator                 User   `json:"creator"`
	CreatesJoinRequest      bool   `json:"creates_join_request"`
	IsPrimary               bool   `json:"is_primary"`
	IsRevoked               bool   `json:"is_revoked"`
	Name                    string `json:"name,omitempty"`
	ExpireDate              int64  `json:"expire_date,omitempty"`
	MemberLimit             int    `json:"member_limit,omitempty"`
	PendingJoinRequestCount int    `json:"pending_join_request_count,omitempty"`
	SubscriptionPeriod      int    `json:"subscription_period,omitempty"`
	SubscriptionPrice       int    `json:"subscription_price,omitempty"`
}
