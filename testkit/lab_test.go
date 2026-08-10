package testkit

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sidwiskers/hermes"
	"github.com/sidwiskers/hermes/fsm"
	"github.com/sidwiskers/hermes/session"
)

type profileState uint8

const (
	profileIdle profileState = iota
	profileWaiting
)

type profileData struct {
	Name string
}

func TestLabRunsStatefulConversation(t *testing.T) {
	lab := NewLab(t)
	store := session.NewMemory[fsm.Snapshot[profileState, profileData]](time.Hour)
	sessions := session.New(store, session.ByChatUser, session.WithNamespace("profile"))
	flow := fsm.New(sessions, profileIdle)
	if err := flow.Add(fsm.Rule[profileState, profileData]{From: profileIdle, Event: "begin", To: profileWaiting}); err != nil {
		t.Fatal(err)
	}

	lab.Bot.Use(flow.Middleware())
	lab.Bot.Command("profile", flow.Then("begin", func(c *hermes.Context) error {
		return c.Send("What should I call you?")
	}))
	lab.Bot.OnUpdate(func(c *hermes.Context) error {
		state, err := flow.State(c)
		if err != nil {
			return err
		}
		if state != profileWaiting {
			return nil
		}
		name := strings.TrimSpace(c.Text())
		if name == "" {
			return c.Send("Please send a name.")
		}
		if err := flow.Set(c, fsm.Snapshot[profileState, profileData]{State: profileIdle, Data: profileData{Name: name}}); err != nil {
			return err
		}
		return c.Send("Saved, " + name + ".")
	})

	alice := lab.PrivateUser(42, "alice")
	alice.Command("profile").Want(Sent("What should I call you?"))
	alice.Text("Ada").Want(Sent("Saved, Ada."))

	messages := alice.Messages()
	if len(messages) != 2 || messages[1].Text != "Saved, Ada." {
		t.Fatalf("messages = %#v", messages)
	}
	if messages[0].From == nil || !messages[0].From.IsBot || messages[0].Chat.ID != 42 {
		t.Fatalf("generated message = %#v", messages[0])
	}
}

func TestLabCallbackAcknowledgesAndEdits(t *testing.T) {
	lab := NewLab(t)
	keyboard := hermes.Keyboard(hermes.Row(hermes.Button("Save", "save")))
	lab.Bot.Command("settings", func(c *hermes.Context) error {
		return c.Send("Settings", hermes.WithKeyboard(keyboard))
	})
	lab.Bot.Callback("save", func(c *hermes.Context) error {
		if err := c.Acknowledge(); err != nil {
			return err
		}
		return c.Edit("Saved")
	})

	alice := lab.PrivateUser(7, "alice")
	alice.Command("settings").Want(Sent("Settings"))
	step := alice.Callback("save").Want(Acknowledged(), Edited("Saved"))
	if len(step.Requests()) != 2 {
		t.Fatalf("requests = %#v", step.Requests())
	}
	message, ok := lab.API.LastMessage(alice.Chat().ID)
	if !ok || message.Text != "Saved" || message.ReplyMarkup == nil {
		t.Fatalf("edited message = %#v, %v", message, ok)
	}
}

func TestLabDeletesVirtualMessage(t *testing.T) {
	lab := NewLab(t)
	lab.Bot.Command("panel", func(c *hermes.Context) error {
		return c.Send("Panel", hermes.WithKeyboard(hermes.Keyboard(hermes.Row(hermes.Button("Close", "close")))))
	})
	lab.Bot.Callback("close", func(c *hermes.Context) error { return c.Delete() })

	alice := lab.PrivateUser(9, "alice")
	alice.Command("panel").Want(Sent("Panel"))
	alice.Callback("close").Want(Deleted())
	if messages := alice.Messages(); len(messages) != 0 {
		t.Fatalf("messages after delete = %#v", messages)
	}
}

func TestLabFailureInjection(t *testing.T) {
	lab := NewLab(t)
	lab.Bot.Command("send", func(c *hermes.Context) error { return c.Send("hello") })
	alice := lab.PrivateUser(1, "alice")

	lab.API.RateLimitNext("sendMessage", 1500*time.Millisecond)
	step := alice.Command("send").WantAPIError(429)
	var apiError *hermes.APIError
	if !errors.As(step.Err(), &apiError) || apiError.RetryAfter() != 2 {
		t.Fatalf("error = %#v", step.Err())
	}

	transportFailure := errors.New("network unavailable")
	lab.API.TransportErrorNext("sendMessage", transportFailure)
	alice.Command("send").WantError(transportFailure)

	lab.API.RespondNext("sendMessage", hermes.Message{MessageID: 99, Chat: alice.Chat(), Text: "custom"})
	alice.Command("send").Want(Sent("hello"))
}

func TestLabCustomMessagesAndReplay(t *testing.T) {
	lab := NewLab(t)
	count := 0
	lab.Bot.On(hermes.PhotoMessage, func(c *hermes.Context) error {
		count++
		return c.Send("photo")
	})
	alice := lab.GroupUser(-1001, 55, "alice")
	step := alice.Message(hermes.Message{Photo: []hermes.PhotoSize{{FileID: "photo", FileUniqueID: "unique", Width: 1, Height: 1}}}).Want(Sent("photo"))
	replayed := step.Replay().Want(Sent("photo"))
	if count != 2 || step.Update().UpdateID != replayed.Update().UpdateID {
		t.Fatalf("count = %d, update = %#v", count, step.Update())
	}
	if step.Update().Message.Chat.Type != "supergroup" || step.Update().Message.From.ID != 55 {
		t.Fatalf("message = %#v", step.Update().Message)
	}
}

func TestLabAttributesConcurrentRequestsToTheirSteps(t *testing.T) {
	lab := NewLab(t)
	lab.Bot.Command("echo", func(c *hermes.Context) error { return c.Send(c.Args()) })
	alice := lab.PrivateUser(1, "alice")
	bob := lab.PrivateUser(2, "bob")

	var aliceStep, bobStep *Step
	var wait sync.WaitGroup
	wait.Add(2)
	go func() {
		defer wait.Done()
		aliceStep = alice.Command("echo", "alpha")
	}()
	go func() {
		defer wait.Done()
		bobStep = bob.Command("echo", "beta")
	}()
	wait.Wait()

	assertSingleTextRequest(t, aliceStep, "alpha")
	assertSingleTextRequest(t, bobStep, "beta")
}

func TestLabUsesChatLocalMessageIDs(t *testing.T) {
	lab := NewLab(t)
	lab.Bot.On(hermes.TextMessage, func(c *hermes.Context) error { return c.Send("received") })

	alice := lab.PrivateUser(10, "alice")
	first := alice.Text("one")
	second := alice.Text("two")
	messages := alice.Messages()
	if first.Update().Message.MessageID != 1 || messages[0].MessageID != 2 ||
		second.Update().Message.MessageID != 3 || messages[1].MessageID != 4 {
		t.Fatalf("private message IDs: first=%d messages=%v second=%d", first.Update().Message.MessageID, messageIDs(messages), second.Update().Message.MessageID)
	}

	bob := lab.PrivateUser(20, "bob")
	step := bob.Text("independent")
	if step.Update().Message.MessageID != 1 || bob.Messages()[0].MessageID != 2 {
		t.Fatalf("second chat did not start its own sequence")
	}
}

func TestLabAllocatesConcurrentCallbackIDsAtomically(t *testing.T) {
	lab := NewLab(t)
	lab.Bot.Command("panel", func(c *hermes.Context) error { return c.Send("Panel") })
	lab.Bot.CallbackPrefix("tap:", func(c *hermes.Context) error { return c.Acknowledge() })
	alice := lab.PrivateUser(30, "alice")
	alice.Command("panel").Want(Sent("Panel"))
	message, _ := lab.API.LastMessage(alice.Chat().ID)

	const callbacks = 8
	steps := make([]*Step, callbacks)
	var wait sync.WaitGroup
	wait.Add(callbacks)
	for index := range callbacks {
		go func() {
			defer wait.Done()
			steps[index] = alice.CallbackOn(message, "tap:"+string(rune('a'+index)))
		}()
	}
	wait.Wait()

	updateIDs := make(map[int64]bool, callbacks)
	callbackIDs := make(map[string]bool, callbacks)
	for _, step := range steps {
		step.Want(Acknowledged())
		updateIDs[step.Update().UpdateID] = true
		callbackIDs[step.Update().CallbackQuery.ID] = true
	}
	if len(updateIDs) != callbacks || len(callbackIDs) != callbacks {
		t.Fatalf("update IDs=%d callback IDs=%d", len(updateIDs), len(callbackIDs))
	}
}

func TestLabResetRetainsActorsAndRoutes(t *testing.T) {
	lab := NewLab(t)
	lab.Bot.Command("start", func(c *hermes.Context) error { return c.Send("ready") })
	alice := lab.PrivateUser(3, "alice")
	alice.Command("start").Want(Sent("ready"))
	lab.API.FailNext("unused", 400, "unused", nil)

	lab.Reset()
	if len(lab.Requests()) != 0 || len(alice.Messages()) != 0 {
		t.Fatalf("reset left requests or messages")
	}
	alice.Command("start").Want(Sent("ready"))
	message, ok := lab.API.LastMessage(3)
	if !ok || message.MessageID != 2 {
		t.Fatalf("message after reset = %#v, %v", message, ok)
	}
}

func TestLabLowLevelResponsesAndMessageErrors(t *testing.T) {
	lab := NewLab(t)
	alice := lab.PrivateUser(4, "alice")
	if bot := lab.API.Bot(); !bot.IsBot || bot.Username == "" {
		t.Fatalf("bot = %#v", bot)
	}

	lab.Bot.Command("count", func(c *hermes.Context) error {
		count, err := c.Bot.GetChatMemberCount(c.Context, c.Message.Chat.ID)
		if err != nil {
			return err
		}
		return c.Bot.SendMessageDraft(c.Context, hermes.SendMessageDraftParams{ChatID: c.Message.Chat.ID, DraftID: 1, Text: string(rune('0' + count))})
	})
	lab.API.RespondNext("getChatMemberCount", 3)
	alice.Command("count").Want(Called("getChatMemberCount"), Called("sendMessageDraft"))

	lab.Bot.Command("bad-edit", func(c *hermes.Context) error { return c.Edit("no") })
	alice.Command("bad-edit").WantAPIError(400)

	missing := lab.PrivateUser(99, "nobody").Callback("none")
	if !errors.Is(missing.Err(), ErrMessageRequired) {
		t.Fatalf("missing callback error = %v", missing.Err())
	}
}

func TestLabExpectationsAndMultipart(t *testing.T) {
	requests := []Request{
		{Method: "answerCallbackQuery", JSON: map[string]any{"text": "done"}},
		{Method: "sendPhoto", Files: map[string]File{"photo": {Name: "photo.jpg"}}},
	}
	for _, expectation := range []Expectation{
		Called("sendPhoto"),
		Answered("done"),
		Uploaded("sendPhoto", "photo", "photo.jpg"),
		Matching("sendPhoto", func(request Request) bool { return request.Files["photo"].Name == "photo.jpg" }),
	} {
		if err := expectation.match(requests); err != nil {
			t.Fatal(err)
		}
	}
	if err := NoCalls().match(nil); err != nil {
		t.Fatal(err)
	}
	for _, expectation := range []Expectation{
		Called("deleteMessage"), Sent("missing"), Edited("missing"), Deleted(), Uploaded("sendPhoto", "photo", "wrong"), Matching("sendPhoto", nil), NoCalls(),
	} {
		if err := expectation.match(requests); err == nil {
			t.Fatalf("expectation %T unexpectedly matched", expectation)
		}
	}
}

func TestRecorderResponderAndTransportError(t *testing.T) {
	client, recorder := NewClient()
	recorder.SetResponder(func(request Request) (Response, error) {
		if request.Method != "getMe" {
			t.Fatalf("method = %q", request.Method)
		}
		return successful(hermes.User{ID: 1, IsBot: true, FirstName: "Lab"}), nil
	})
	user, err := client.GetMe(context.Background())
	if err != nil || user.ID != 1 {
		t.Fatalf("user = %#v, err = %v", user, err)
	}

	recorder.Respond(hermes.User{ID: 2, IsBot: true, FirstName: "Queued"})
	user, err = client.GetMe(context.Background())
	if err != nil || user.ID != 2 {
		t.Fatalf("queued user = %#v, err = %v", user, err)
	}

	failure := errors.New("transport failed")
	recorder.EnqueueError(failure)
	_, err = client.GetMe(context.Background())
	if !errors.Is(err, failure) {
		t.Fatalf("error = %v", err)
	}
}

func TestRecorderStructShapesRemainCompatible(t *testing.T) {
	t.Parallel()
	_ = Request{"sendMessage", nil, nil, nil, nil, nil}
	_ = Response{http.StatusOK, true, 0, "", nil}
}

func TestLabMediaResponses(t *testing.T) {
	lab := NewLab(t)
	alice := lab.PrivateUser(5, "alice")
	lab.Bot.Command("photo", func(c *hermes.Context) error {
		message, err := c.Bot.SendPhoto(c.Context, hermes.SendPhotoParams{ChatID: c.Message.Chat.ID, Photo: "file-id"})
		if err != nil {
			return err
		}
		if len(message.Photo) == 0 || message.Photo[0].FileID != "file-id" {
			return errors.New("virtual photo missing")
		}
		return nil
	})
	alice.Command("photo").Want(Called("sendPhoto"))
}

func assertSingleTextRequest(t *testing.T, step *Step, text string) {
	t.Helper()
	if step == nil || step.Err() != nil {
		t.Fatalf("step = %#v", step)
	}
	requests := step.Requests()
	if len(requests) != 1 || requests[0].Method != "sendMessage" || requestString(requests[0], "text") != text {
		t.Fatalf("requests = %#v", requests)
	}
}

func messageIDs(messages []hermes.Message) []int {
	result := make([]int, len(messages))
	for index, message := range messages {
		result[index] = message.MessageID
	}
	return result
}
