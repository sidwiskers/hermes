package api

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Bot API 10.2 method manifest, published 2026-07-14 at
// https://core.telegram.org/bots/api.
var botAPI10_2Methods = strings.Fields(`
getUpdates
setWebhook
deleteWebhook
getWebhookInfo
getMe
logOut
close
sendMessage
forwardMessage
forwardMessages
copyMessage
copyMessages
sendPhoto
sendLivePhoto
sendAudio
sendDocument
sendVideo
sendAnimation
sendVoice
sendVideoNote
sendPaidMedia
sendMediaGroup
sendLocation
sendVenue
sendContact
sendPoll
sendChecklist
sendDice
sendMessageDraft
sendChatAction
setMessageReaction
getUserProfilePhotos
getUserProfileAudios
setUserEmojiStatus
getFile
banChatMember
unbanChatMember
restrictChatMember
promoteChatMember
setChatAdministratorCustomTitle
setChatMemberTag
banChatSenderChat
unbanChatSenderChat
setChatPermissions
exportChatInviteLink
createChatInviteLink
editChatInviteLink
createChatSubscriptionInviteLink
editChatSubscriptionInviteLink
revokeChatInviteLink
approveChatJoinRequest
declineChatJoinRequest
answerChatJoinRequestQuery
sendChatJoinRequestWebApp
setChatPhoto
deleteChatPhoto
setChatTitle
setChatDescription
pinChatMessage
unpinChatMessage
unpinAllChatMessages
leaveChat
getChat
getChatAdministrators
getChatMemberCount
getChatMember
getUserPersonalChatMessages
setChatStickerSet
deleteChatStickerSet
getForumTopicIconStickers
createForumTopic
editForumTopic
closeForumTopic
reopenForumTopic
deleteForumTopic
unpinAllForumTopicMessages
editGeneralForumTopic
closeGeneralForumTopic
reopenGeneralForumTopic
hideGeneralForumTopic
unhideGeneralForumTopic
unpinAllGeneralForumTopicMessages
answerCallbackQuery
answerGuestQuery
getUserChatBoosts
getBusinessConnection
getManagedBotToken
replaceManagedBotToken
getManagedBotAccessSettings
setManagedBotAccessSettings
setMyCommands
deleteMyCommands
getMyCommands
setMyName
getMyName
setMyDescription
getMyDescription
setMyShortDescription
getMyShortDescription
setMyProfilePhoto
removeMyProfilePhoto
setChatMenuButton
getChatMenuButton
setMyDefaultAdministratorRights
getMyDefaultAdministratorRights
getAvailableGifts
sendGift
giftPremiumSubscription
verifyUser
verifyChat
removeUserVerification
removeChatVerification
readBusinessMessage
deleteBusinessMessages
setBusinessAccountName
setBusinessAccountUsername
setBusinessAccountBio
setBusinessAccountProfilePhoto
removeBusinessAccountProfilePhoto
setBusinessAccountGiftSettings
getBusinessAccountStarBalance
transferBusinessAccountStars
getBusinessAccountGifts
getUserGifts
getChatGifts
convertGiftToStars
upgradeGift
transferGift
postStory
repostStory
editStory
deleteStory
answerWebAppQuery
savePreparedInlineMessage
savePreparedKeyboardButton
editMessageText
editMessageCaption
editMessageMedia
editMessageLiveLocation
stopMessageLiveLocation
editMessageChecklist
editMessageReplyMarkup
stopPoll
editEphemeralMessageText
editEphemeralMessageMedia
editEphemeralMessageCaption
editEphemeralMessageReplyMarkup
approveSuggestedPost
declineSuggestedPost
deleteMessage
deleteMessages
deleteEphemeralMessage
deleteMessageReaction
deleteAllMessageReactions
sendSticker
getStickerSet
getCustomEmojiStickers
uploadStickerFile
createNewStickerSet
addStickerToSet
setStickerPositionInSet
deleteStickerFromSet
replaceStickerInSet
setStickerEmojiList
setStickerKeywords
setStickerMaskPosition
setStickerSetTitle
setStickerSetThumbnail
setCustomEmojiStickerSetThumbnail
deleteStickerSet
sendRichMessage
sendRichMessageDraft
answerInlineQuery
sendInvoice
createInvoiceLink
answerShippingQuery
answerPreCheckoutQuery
getMyStarBalance
getStarTransactions
refundStarPayment
editUserStarSubscription
setPassportDataErrors
sendGame
setGameScore
getGameHighScores
`)

func TestBotAPI10_2MethodManifest(t *testing.T) {
	t.Parallel()

	if len(botAPI10_2Methods) != 185 {
		t.Fatalf("manifest contains %d methods, want 185", len(botAPI10_2Methods))
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	productionStrings := make(map[string]struct{}, len(botAPI10_2Methods))
	files := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		parsed, parseErr := parser.ParseFile(files, name, nil, 0)
		if parseErr != nil {
			t.Fatalf("parse %s: %v", name, parseErr)
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			value, unquoteErr := strconv.Unquote(literal.Value)
			if unquoteErr == nil {
				productionStrings[value] = struct{}{}
			}
			return true
		})
	}
	for _, method := range botAPI10_2Methods {
		if _, exists := productionStrings[method]; !exists {
			t.Errorf("Bot API 10.2 method %s has no production entry point", method)
		}
	}
}
