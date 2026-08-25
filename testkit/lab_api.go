package testkit

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sidwiskers/hermes"
)

// LabAPI is Lab's deterministic Bot API emulator. Common messaging methods
// are modeled automatically; RespondNext is the escape hatch for any method
// whose exact result matters to a test.
type LabAPI struct {
	mu       sync.Mutex
	recorder *Recorder
	lab      *Lab
	bot      hermes.User
	chats    map[int64]hermes.Chat
	users    map[int64]hermes.User
	messages map[int64][]hermes.Message
	planned  map[string][]roundTripResult
}

func newLabAPI(recorder *Recorder) *LabAPI {
	return &LabAPI{
		recorder: recorder,
		bot: hermes.User{
			ID:        9_000_000_001,
			IsBot:     true,
			FirstName: "Hermes Lab",
			Username:  "hermes_lab_bot",
		},
		chats:    make(map[int64]hermes.Chat),
		users:    make(map[int64]hermes.User),
		messages: make(map[int64][]hermes.Message),
		planned:  make(map[string][]roundTripResult),
	}
}

// Bot returns the virtual bot user returned by getMe and attached to generated
// outgoing messages.
func (a *LabAPI) Bot() hermes.User {
	if a == nil {
		return hermes.User{}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.bot
}

// RespondNext queues a successful result for the next matching method call.
// It overrides automatic emulation for that call only.
func (a *LabAPI) RespondNext(method string, result any) {
	a.enqueue(method, roundTripResult{response: Response{StatusCode: http.StatusOK, Result: result}})
}

// FailNext queues a Telegram API error for the next matching method call.
func (a *LabAPI) FailNext(method string, code int, description string, parameters *hermes.ResponseParameters) {
	if code <= 0 {
		code = http.StatusBadRequest
	}
	a.enqueue(method, roundTripResult{response: Response{
		StatusCode:  code,
		ErrorCode:   code,
		Description: description,
		Parameters:  parameters,
	}})
}

// RateLimitNext queues a Telegram 429 response. Fractions of a second are
// rounded up because Telegram reports retry_after in whole seconds.
func (a *LabAPI) RateLimitNext(method string, retryAfter time.Duration) {
	seconds := int64(retryAfter / time.Second)
	if retryAfter > 0 && retryAfter%time.Second != 0 {
		seconds++
	}
	if seconds < 1 {
		seconds = 1
	}
	maxInt := int64(^uint(0) >> 1)
	if seconds > maxInt {
		seconds = maxInt
	}
	a.FailNext(method, http.StatusTooManyRequests, "Too Many Requests", &hermes.ResponseParameters{RetryAfter: int(seconds)})
}

// TransportErrorNext makes the next matching method fail before an HTTP
// response is available.
func (a *LabAPI) TransportErrorNext(method string, err error) {
	if err == nil {
		err = errors.New("Hermes Lab transport failure")
	}
	a.enqueue(method, roundTripResult{err: err})
}

// Messages returns a snapshot of virtual bot messages retained in a chat.
func (a *LabAPI) Messages(chatID int64) []hermes.Message {
	if a == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	result := make([]hermes.Message, len(a.messages[chatID]))
	copy(result, a.messages[chatID])
	return result
}

// LastMessage returns the most recent virtual bot message in a chat.
func (a *LabAPI) LastMessage(chatID int64) (hermes.Message, bool) {
	if a == nil {
		return hermes.Message{}, false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	messages := a.messages[chatID]
	if len(messages) == 0 {
		return hermes.Message{}, false
	}
	return messages[len(messages)-1], true
}

func (a *LabAPI) enqueue(method string, result roundTripResult) {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.planned[method] = append(a.planned[method], result)
}

func (a *LabAPI) register(user hermes.User, chat hermes.Chat) {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.users[user.ID] = user
	a.chats[chat.ID] = chat
}

func (a *LabAPI) reset() {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.messages = make(map[int64][]hermes.Message)
	a.planned = make(map[string][]roundTripResult)
}

func (a *LabAPI) respond(request Request) (Response, error) {
	if result, ok := a.takePlanned(request.Method); ok {
		return result.response, result.err
	}

	switch request.Method {
	case "getMe":
		return successful(a.Bot()), nil
	case "sendMessage", "sendPhoto", "sendAnimation", "sendAudio", "sendDocument",
		"sendSticker", "sendVideo", "sendVideoNote", "sendVoice", "sendContact",
		"sendLocation", "sendVenue", "sendPoll", "sendDice", "sendGame",
		"sendInvoice", "sendPaidMedia", "sendChecklist", "sendLivePhoto",
		"sendRichMessage", "forwardMessage":
		return successful(a.emulateSend(request)), nil
	case "sendMediaGroup":
		count := mediaCount(request)
		messages := make([]hermes.Message, count)
		for index := range messages {
			messages[index] = a.emulateSend(request)
		}
		return successful(messages), nil
	case "copyMessage":
		message := a.emulateSend(request)
		return successful(map[string]int{"message_id": message.MessageID}), nil
	case "editMessageText", "editMessageCaption", "editMessageReplyMarkup", "editMessageMedia",
		"editMessageLiveLocation", "stopMessageLiveLocation", "editMessageChecklist":
		return a.emulateEdit(request), nil
	case "deleteMessage":
		return a.emulateDelete(request), nil
	case "deleteMessages":
		return a.emulateDeleteMany(request), nil
	default:
		return successful(true), nil
	}
}

func (a *LabAPI) takePlanned(method string) (roundTripResult, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	responses := a.planned[method]
	if len(responses) == 0 {
		return roundTripResult{}, false
	}
	response := responses[0]
	if len(responses) == 1 {
		delete(a.planned, method)
	} else {
		a.planned[method] = responses[1:]
	}
	return response, true
}

func (a *LabAPI) emulateSend(request Request) hermes.Message {
	a.mu.Lock()
	defer a.mu.Unlock()

	chatID, _ := requestInt64(request, "chat_id")
	messageID := 1
	if a.lab != nil {
		messageID = a.lab.takeOutgoingMessageID(chatID)
	}
	chat := a.chats[chatID]
	if chat.ID == 0 {
		chat = hermes.Chat{ID: chatID, Type: "private"}
	}
	bot := a.bot
	message := hermes.Message{
		MessageID: messageID,
		From:      &bot,
		Date:      labEpoch + int64(messageID),
		Chat:      chat,
		Text:      requestString(request, "text"),
		Caption:   requestString(request, "caption"),
	}
	if receiverID, ok := requestEphemeralReceiver(request); ok {
		receiver := a.users[receiverID]
		if receiver.ID == 0 {
			receiver = hermes.User{ID: receiverID, FirstName: "User " + strconv.FormatInt(receiverID, 10)}
		}
		message.ReceiverUser = &receiver
		message.EphemeralMessageID = messageID
	}
	if markup, ok := requestMarkup(request); ok {
		message.ReplyMarkup = markup
	}
	populateMessageMedia(&message, request)
	a.messages[chatID] = append(a.messages[chatID], message)
	return message
}

func requestEphemeralReceiver(request Request) (int64, bool) {
	value, ok := requestField(request, "ephemeral_message_parameters")
	if !ok {
		return 0, false
	}
	if encoded, isString := value.(string); isString {
		var decoded map[string]any
		if json.Unmarshal([]byte(encoded), &decoded) != nil {
			return 0, false
		}
		value = decoded
	}
	parameters, ok := value.(map[string]any)
	if !ok {
		return 0, false
	}
	receiver, ok := parameters["receiver_user_id"]
	if !ok {
		return 0, false
	}
	switch receiver := receiver.(type) {
	case float64:
		return int64(receiver), true
	case json.Number:
		parsed, err := receiver.Int64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func (a *LabAPI) emulateEdit(request Request) Response {
	if requestString(request, "inline_message_id") != "" {
		return successful(true)
	}
	chatID, chatOK := requestInt64(request, "chat_id")
	messageID, messageOK := requestInt(request, "message_id")
	if !chatOK || !messageOK {
		return failed(http.StatusBadRequest, "message identifier is required")
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	messages := a.messages[chatID]
	for index := range messages {
		if messages[index].MessageID != messageID {
			continue
		}
		if value, ok := requestField(request, "text"); ok {
			messages[index].Text = stringValue(value)
		}
		if value, ok := requestField(request, "caption"); ok {
			messages[index].Caption = stringValue(value)
		}
		if _, ok := requestField(request, "reply_markup"); ok {
			messages[index].ReplyMarkup, _ = requestMarkup(request)
		}
		a.messages[chatID] = messages
		return successful(messages[index])
	}
	return failed(http.StatusBadRequest, "message to edit not found")
}

func (a *LabAPI) emulateDelete(request Request) Response {
	chatID, chatOK := requestInt64(request, "chat_id")
	messageID, messageOK := requestInt(request, "message_id")
	if !chatOK || !messageOK {
		return failed(http.StatusBadRequest, "message identifier is required")
	}
	if !a.deleteMessage(chatID, messageID) {
		return failed(http.StatusBadRequest, "message to delete not found")
	}
	return successful(true)
}

func (a *LabAPI) emulateDeleteMany(request Request) Response {
	chatID, ok := requestInt64(request, "chat_id")
	if !ok {
		return failed(http.StatusBadRequest, "chat identifier is required")
	}
	for _, messageID := range requestInts(request, "message_ids") {
		a.deleteMessage(chatID, messageID)
	}
	return successful(true)
}

func (a *LabAPI) deleteMessage(chatID int64, messageID int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	messages := a.messages[chatID]
	for index := range messages {
		if messages[index].MessageID == messageID {
			a.messages[chatID] = append(messages[:index], messages[index+1:]...)
			return true
		}
	}
	return false
}

func successful(result any) Response {
	return Response{StatusCode: http.StatusOK, Result: result}
}

func failed(code int, description string) Response {
	return Response{StatusCode: code, ErrorCode: code, Description: description}
}

func requestField(request Request, name string) (any, bool) {
	if request.JSON != nil {
		value, ok := request.JSON[name]
		if ok {
			return value, true
		}
	}
	value, ok := request.Form[name]
	return value, ok
}

func requestString(request Request, name string) string {
	value, _ := requestField(request, name)
	return stringValue(value)
}

func stringValue(value any) string {
	switch value := value.(type) {
	case string:
		return value
	case json.Number:
		return value.String()
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	default:
		return ""
	}
}

func requestInt64(request Request, name string) (int64, bool) {
	value, ok := requestField(request, name)
	if !ok {
		return 0, false
	}
	switch value := value.(type) {
	case float64:
		return int64(value), true
	case int64:
		return value, true
	case int:
		return int64(value), true
	case json.Number:
		parsed, err := value.Int64()
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseInt(value, 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func requestInt(request Request, name string) (int, bool) {
	value, ok := requestInt64(request, name)
	return int(value), ok
}

func requestInts(request Request, name string) []int {
	value, ok := requestField(request, name)
	if !ok {
		return nil
	}
	if encoded, ok := value.(string); ok {
		var result []int
		_ = json.Unmarshal([]byte(encoded), &result)
		return result
	}
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	result := make([]int, 0, len(items))
	for _, item := range items {
		if number, ok := item.(float64); ok {
			result = append(result, int(number))
		}
	}
	return result
}

func requestMarkup(request Request) (*hermes.InlineKeyboardMarkup, bool) {
	value, ok := requestField(request, "reply_markup")
	if !ok || value == nil {
		return nil, ok
	}
	var data []byte
	switch value := value.(type) {
	case string:
		data = []byte(value)
	default:
		data, _ = json.Marshal(value)
	}
	var markup hermes.InlineKeyboardMarkup
	if err := json.Unmarshal(data, &markup); err != nil {
		return nil, false
	}
	return &markup, true
}

func mediaCount(request Request) int {
	value, ok := requestField(request, "media")
	if !ok {
		return 1
	}
	if items, ok := value.([]any); ok && len(items) != 0 {
		return len(items)
	}
	if encoded, ok := value.(string); ok {
		var items []any
		if json.Unmarshal([]byte(encoded), &items) == nil && len(items) != 0 {
			return len(items)
		}
	}
	return 1
}

func populateMessageMedia(message *hermes.Message, request Request) {
	if message == nil {
		return
	}
	fileID := requestString(request, strings.ToLower(strings.TrimPrefix(request.Method, "send")))
	if fileID == "" {
		fileID = "lab-file"
	}
	switch request.Method {
	case "sendPhoto":
		message.Photo = []hermes.PhotoSize{{FileID: fileID, FileUniqueID: "lab-photo", Width: 1, Height: 1}}
	case "sendAnimation":
		message.Animation = &hermes.Animation{FileID: fileID, FileUniqueID: "lab-animation", Width: 1, Height: 1}
	case "sendAudio":
		message.Audio = &hermes.Audio{FileID: fileID, FileUniqueID: "lab-audio"}
	case "sendDocument":
		message.Document = &hermes.Document{FileID: fileID, FileUniqueID: "lab-document"}
	case "sendSticker":
		message.Sticker = &hermes.Sticker{FileID: fileID, FileUniqueID: "lab-sticker", Type: "regular", Width: 1, Height: 1}
	case "sendVideo":
		message.Video = &hermes.Video{FileID: fileID, FileUniqueID: "lab-video", Width: 1, Height: 1}
	case "sendVideoNote":
		message.VideoNote = &hermes.VideoNote{FileID: fileID, FileUniqueID: "lab-video-note", Length: 1}
	case "sendVoice":
		message.Voice = &hermes.Voice{FileID: fileID, FileUniqueID: "lab-voice"}
	case "sendContact":
		message.Contact = &hermes.Contact{PhoneNumber: requestString(request, "phone_number"), FirstName: requestString(request, "first_name")}
	case "sendLocation":
		message.Location = &hermes.Location{Latitude: requestFloat(request, "latitude"), Longitude: requestFloat(request, "longitude")}
	}
}

func requestFloat(request Request, name string) float64 {
	value, _ := requestField(request, name)
	switch value := value.(type) {
	case float64:
		return value
	case string:
		parsed, _ := strconv.ParseFloat(value, 64)
		return parsed
	default:
		return 0
	}
}
