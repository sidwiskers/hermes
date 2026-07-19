package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"sync"
	"testing"
)

// TestTypedBooleanMethodContracts exercises the validation, serialization,
// method selection, envelope decoding, and boolean result handling of the
// broad typed surface. Focused tests cover response-rich and multipart methods.
func TestTypedBooleanMethodContracts(t *testing.T) {
	t.Parallel()

	var (
		mu         sync.Mutex
		lastMethod string
	)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		var payload any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil && err != io.EOF {
			t.Errorf("decode %s request: %v", request.URL.Path, err)
		}
		mu.Lock()
		lastMethod = path.Base(request.URL.Path)
		mu.Unlock()
		writer.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	ctx := context.Background()
	tests := []struct {
		method string
		call   func() error
	}{
		{"logOut", func() error { return client.LogOut(ctx) }},
		{"close", func() error { return client.Close(ctx) }},
		{"sendMessageDraft", func() error {
			return client.SendMessageDraft(ctx, SendMessageDraftParams{ChatID: 1, DraftID: 2, Text: "draft"})
		}},
		{"readBusinessMessage", func() error {
			return client.ReadBusinessMessage(ctx, ReadBusinessMessageParams{BusinessConnectionID: "business", ChatID: 1, MessageID: 2})
		}},
		{"deleteBusinessMessages", func() error {
			return client.DeleteBusinessMessages(ctx, DeleteBusinessMessagesParams{BusinessConnectionID: "business", MessageIDs: []int{1, 2}})
		}},
		{"setBusinessAccountName", func() error {
			return client.SetBusinessAccountName(ctx, SetBusinessAccountNameParams{BusinessConnectionID: "business", FirstName: "Hermes"})
		}},
		{"setBusinessAccountUsername", func() error {
			return client.SetBusinessAccountUsername(ctx, SetBusinessAccountUsernameParams{BusinessConnectionID: "business", Username: "hermes_bot"})
		}},
		{"setBusinessAccountBio", func() error {
			return client.SetBusinessAccountBio(ctx, SetBusinessAccountBioParams{BusinessConnectionID: "business", Bio: "Fast"})
		}},
		{"removeBusinessAccountProfilePhoto", func() error {
			return client.RemoveBusinessAccountProfilePhoto(ctx, RemoveBusinessAccountProfilePhotoParams{BusinessConnectionID: "business"})
		}},
		{"setBusinessAccountGiftSettings", func() error {
			return client.SetBusinessAccountGiftSettings(ctx, SetBusinessAccountGiftSettingsParams{BusinessConnectionID: "business", ShowGiftButton: true})
		}},
		{"transferBusinessAccountStars", func() error {
			return client.TransferBusinessAccountStars(ctx, TransferBusinessAccountStarsParams{BusinessConnectionID: "business", StarCount: 10})
		}},
		{"deleteChatPhoto", func() error { return client.DeleteChatPhoto(ctx, int64(1)) }},
		{"setChatTitle", func() error {
			return client.SetChatTitle(ctx, SetChatTitleParams{ChatID: 1, Title: "Hermes"})
		}},
		{"setChatDescription", func() error {
			return client.SetChatDescription(ctx, SetChatDescriptionParams{ChatID: 1, Description: "Description"})
		}},
		{"pinChatMessage", func() error {
			return client.PinChatMessage(ctx, PinChatMessageParams{ChatID: 1, MessageID: 2})
		}},
		{"unpinChatMessage", func() error {
			return client.UnpinChatMessage(ctx, UnpinChatMessageParams{ChatID: 1, MessageID: 2})
		}},
		{"unpinAllChatMessages", func() error { return client.UnpinAllChatMessages(ctx, int64(1)) }},
		{"setChatStickerSet", func() error {
			return client.SetChatStickerSet(ctx, SetChatStickerSetParams{ChatID: 1, StickerSetName: "set_by_bot"})
		}},
		{"deleteChatStickerSet", func() error { return client.DeleteChatStickerSet(ctx, int64(1)) }},
		{"setMyCommands", func() error {
			return client.SetMyCommands(ctx, SetMyCommandsParams{Commands: []BotCommand{{Command: "start", Description: "Start"}}, Scope: DefaultCommandScope()})
		}},
		{"deleteMyCommands", func() error {
			return client.DeleteMyCommands(ctx, DeleteMyCommandsParams{Scope: AllPrivateChatsCommandScope()})
		}},
		{"banChatMember", func() error {
			return client.BanChatMember(ctx, BanChatMemberParams{ChatID: 1, UserID: 2})
		}},
		{"unbanChatMember", func() error {
			return client.UnbanChatMember(ctx, UnbanChatMemberParams{ChatID: 1, UserID: 2})
		}},
		{"promoteChatMember", func() error {
			return client.PromoteChatMember(ctx, PromoteChatMemberParams{ChatID: 1, UserID: 2, CanManageChat: true})
		}},
		{"setChatAdministratorCustomTitle", func() error {
			return client.SetChatAdministratorCustomTitle(ctx, SetChatAdministratorCustomTitleParams{ChatID: 1, UserID: 2, CustomTitle: "Admin"})
		}},
		{"setChatMemberTag", func() error {
			return client.SetChatMemberTag(ctx, SetChatMemberTagParams{ChatID: 1, UserID: 2, Tag: "staff"})
		}},
		{"banChatSenderChat", func() error {
			return client.BanChatSenderChat(ctx, ChatSenderParams{ChatID: 1, SenderChatID: 2})
		}},
		{"unbanChatSenderChat", func() error {
			return client.UnbanChatSenderChat(ctx, ChatSenderParams{ChatID: 1, SenderChatID: 2})
		}},
		{"setChatPermissions", func() error {
			return client.SetChatPermissions(ctx, SetChatPermissionsParams{ChatID: 1, Permissions: ChatPermissions{CanSendMessages: true}})
		}},
		{"approveChatJoinRequest", func() error {
			return client.ApproveChatJoinRequest(ctx, ChatJoinRequestParams{ChatID: 1, UserID: 2})
		}},
		{"declineChatJoinRequest", func() error {
			return client.DeclineChatJoinRequest(ctx, ChatJoinRequestParams{ChatID: 1, UserID: 2})
		}},
		{"answerChatJoinRequestQuery", func() error {
			return client.AnswerChatJoinRequestQuery(ctx, AnswerChatJoinRequestQueryParams{ChatJoinRequestQueryID: "query", Result: JoinRequestApprove})
		}},
		{"sendChatJoinRequestWebApp", func() error {
			return client.SendChatJoinRequestWebApp(ctx, SendChatJoinRequestWebAppParams{ChatJoinRequestQueryID: "query", WebAppURL: "https://example.com"})
		}},
		{"editEphemeralMessageText", func() error {
			return client.EditEphemeralText(ctx, EditEphemeralMessageTextParams{ChatID: 1, ReceiverUserID: 2, EphemeralMessageID: 3, Text: "text"})
		}},
		{"editEphemeralMessageMedia", func() error {
			return client.EditEphemeralMedia(ctx, EditEphemeralMessageMediaParams{ChatID: 1, ReceiverUserID: 2, EphemeralMessageID: 3, Media: InputMedia{Type: "photo", Media: "file-id"}})
		}},
		{"editEphemeralMessageCaption", func() error {
			return client.EditEphemeralCaption(ctx, EditEphemeralMessageCaptionParams{ChatID: 1, ReceiverUserID: 2, EphemeralMessageID: 3, Caption: "caption"})
		}},
		{"editEphemeralMessageReplyMarkup", func() error {
			return client.EditEphemeralReplyMarkup(ctx, EditEphemeralMessageReplyMarkupParams{ChatID: 1, ReceiverUserID: 2, EphemeralMessageID: 3})
		}},
		{"deleteEphemeralMessage", func() error {
			return client.DeleteEphemeral(ctx, DeleteEphemeralMessageParams{ChatID: 1, ReceiverUserID: 2, EphemeralMessageID: 3})
		}},
		{"closeForumTopic", func() error {
			return client.CloseForumTopic(ctx, ForumTopicTargetParams{ChatID: 1, MessageThreadID: 2})
		}},
		{"reopenForumTopic", func() error {
			return client.ReopenForumTopic(ctx, ForumTopicTargetParams{ChatID: 1, MessageThreadID: 2})
		}},
		{"deleteForumTopic", func() error {
			return client.DeleteForumTopic(ctx, ForumTopicTargetParams{ChatID: 1, MessageThreadID: 2})
		}},
		{"unpinAllForumTopicMessages", func() error {
			return client.UnpinAllForumTopicMessages(ctx, ForumTopicTargetParams{ChatID: 1, MessageThreadID: 2})
		}},
		{"editGeneralForumTopic", func() error {
			return client.EditGeneralForumTopic(ctx, EditGeneralForumTopicParams{ChatID: 1, Name: "General"})
		}},
		{"closeGeneralForumTopic", func() error {
			return client.CloseGeneralForumTopic(ctx, GeneralForumTopicParams{ChatID: 1})
		}},
		{"reopenGeneralForumTopic", func() error {
			return client.ReopenGeneralForumTopic(ctx, GeneralForumTopicParams{ChatID: 1})
		}},
		{"hideGeneralForumTopic", func() error {
			return client.HideGeneralForumTopic(ctx, GeneralForumTopicParams{ChatID: 1})
		}},
		{"unhideGeneralForumTopic", func() error {
			return client.UnhideGeneralForumTopic(ctx, GeneralForumTopicParams{ChatID: 1})
		}},
		{"unpinAllGeneralForumTopicMessages", func() error {
			return client.UnpinAllGeneralForumTopicMessages(ctx, GeneralForumTopicParams{ChatID: 1})
		}},
		{"setUserEmojiStatus", func() error {
			return client.SetUserEmojiStatus(ctx, SetUserEmojiStatusParams{UserID: 2, EmojiStatusCustomEmojiID: "emoji"})
		}},
		{"setMyName", func() error { return client.SetMyName(ctx, SetMyNameParams{Name: "Hermes"}) }},
		{"setMyDescription", func() error {
			return client.SetMyDescription(ctx, SetMyDescriptionParams{Description: "Description"})
		}},
		{"setMyShortDescription", func() error {
			return client.SetMyShortDescription(ctx, SetMyShortDescriptionParams{ShortDescription: "Short"})
		}},
		{"setChatMenuButton", func() error {
			return client.SetChatMenuButton(ctx, SetChatMenuButtonParams{ChatID: 1, MenuButton: pointerMenuButton(CommandsMenuButton())})
		}},
		{"setMyDefaultAdministratorRights", func() error {
			return client.SetMyDefaultAdministratorRights(ctx, SetMyDefaultAdministratorRightsParams{ForChannels: true})
		}},
		{"verifyUser", func() error {
			return client.VerifyUser(ctx, VerifyUserParams{UserID: 2, CustomDescription: "Verified"})
		}},
		{"verifyChat", func() error {
			return client.VerifyChat(ctx, VerifyChatParams{ChatID: 1, CustomDescription: "Verified"})
		}},
		{"removeUserVerification", func() error {
			return client.RemoveUserVerification(ctx, RemoveUserVerificationParams{UserID: 2})
		}},
		{"removeChatVerification", func() error {
			return client.RemoveChatVerification(ctx, RemoveChatVerificationParams{ChatID: 1})
		}},
		{"deleteMessageReaction", func() error {
			return client.DeleteMessageReaction(ctx, DeleteMessageReactionParams{ChatID: 1, MessageID: 2, UserID: 3})
		}},
		{"deleteAllMessageReactions", func() error {
			return client.DeleteAllMessageReactions(ctx, DeleteAllMessageReactionsParams{ChatID: 1, UserID: 2})
		}},
		{"setManagedBotAccessSettings", func() error {
			return client.SetManagedBotAccessSettings(ctx, SetManagedBotAccessSettingsParams{UserID: 2, AddedUserIDs: []int64{3}})
		}},
		{"answerShippingQuery", func() error {
			return client.AnswerShippingQuery(ctx, AnswerShippingQueryParams{ShippingQueryID: "shipping", ErrorMessage: "Unavailable"})
		}},
		{"answerPreCheckoutQuery", func() error {
			return client.AnswerPreCheckoutQuery(ctx, AnswerPreCheckoutQueryParams{PreCheckoutQueryID: "checkout", OK: true})
		}},
		{"refundStarPayment", func() error {
			return client.RefundStarPayment(ctx, RefundStarPaymentParams{UserID: 2, TelegramPaymentChargeID: "charge"})
		}},
		{"editUserStarSubscription", func() error {
			return client.EditUserStarSubscription(ctx, EditUserStarSubscriptionParams{UserID: 2, TelegramPaymentChargeID: "charge", IsCanceled: true})
		}},
		{"convertGiftToStars", func() error {
			return client.ConvertGiftToStars(ctx, OwnedGiftParams{BusinessConnectionID: "business", OwnedGiftID: "gift"})
		}},
		{"upgradeGift", func() error {
			return client.UpgradeGift(ctx, UpgradeGiftParams{BusinessConnectionID: "business", OwnedGiftID: "gift", StarCount: 1})
		}},
		{"transferGift", func() error {
			return client.TransferGift(ctx, TransferGiftParams{BusinessConnectionID: "business", OwnedGiftID: "gift", NewOwnerChatID: 2})
		}},
		{"setPassportDataErrors", func() error {
			return client.SetPassportDataErrors(ctx, SetPassportDataErrorsParams{UserID: 2, Errors: []PassportElementError{PassportElementErrorDataField{
				Type: "personal_details", FieldName: "first_name", DataHash: "hash", Message: "Invalid",
			}}})
		}},
		{"approveSuggestedPost", func() error {
			return client.ApproveSuggestedPost(ctx, ApproveSuggestedPostParams{ChatID: 1, MessageID: 2})
		}},
		{"declineSuggestedPost", func() error {
			return client.DeclineSuggestedPost(ctx, DeclineSuggestedPostParams{ChatID: 1, MessageID: 2, Comment: "No"})
		}},
		{"deleteMessage", func() error {
			return client.DeleteMessage(ctx, DeleteMessageParams{ChatID: 1, MessageID: 2})
		}},
		{"deleteMessages", func() error {
			return client.DeleteMessages(ctx, DeleteMessagesParams{ChatID: 1, MessageIDs: []int{2, 3}})
		}},
		{"answerCallbackQuery", func() error {
			return client.AnswerCallback(ctx, AnswerCallbackQueryParams{CallbackQueryID: "callback"})
		}},
		{"sendChatAction", func() error {
			return client.SendChatAction(ctx, SendChatActionParams{ChatID: 1, Action: ActionTyping})
		}},
		{"removeMyProfilePhoto", func() error { return client.RemoveMyProfilePhoto(ctx) }},
		{"deleteStory", func() error {
			return client.DeleteStory(ctx, DeleteStoryParams{BusinessConnectionID: "business", StoryID: 2})
		}},
	}

	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			mu.Lock()
			lastMethod = ""
			mu.Unlock()
			if err := test.call(); err != nil {
				t.Fatal(err)
			}
			mu.Lock()
			got := lastMethod
			mu.Unlock()
			if got != test.method {
				t.Fatalf("method = %q, want %q", got, test.method)
			}
		})
	}
}

func pointerMenuButton(value MenuButton) *MenuButton { return &value }

func TestTypedBooleanMethodsRejectFalse(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(writer, `{"ok":true,"result":false}`)
	}))
	defer server.Close()
	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	if err := client.Close(context.Background()); err == nil {
		t.Fatal("expected false-result error")
	}
}
