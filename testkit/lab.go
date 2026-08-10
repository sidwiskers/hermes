package testkit

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/sidwiskers/hermes"
)

var (
	// ErrLabRequired reports an Actor or Step detached from its Lab.
	ErrLabRequired = errors.New("hermes/testkit: lab is required")
	// ErrMessageRequired reports a callback without an available message.
	ErrMessageRequired = errors.New("hermes/testkit: callback message is required")
)

// TB is the subset of testing.TB used by Lab expectations.
type TB interface {
	Helper()
	Fatalf(format string, args ...any)
}

// Lab is a deterministic in-process Telegram environment. It owns a Hermes
// bot, a virtual Bot API, and actors that can deliver updates without network
// access. Lab is intended for tests and is safe for concurrent use.
type Lab struct {
	Bot *hermes.Bot
	API *LabAPI

	mu           sync.Mutex
	tb           TB
	nextUpdate   int64
	nextMessages map[int64]int
	nextCallback int64
	nextStep     uint64
}

// NewLab creates a bot connected to a deterministic virtual Bot API. Options
// may configure ordinary Bot behavior; Lab's transport, endpoint, and username
// are always applied last so tests cannot accidentally use the network.
func NewLab(tb TB, options ...hermes.Option) *Lab {
	recorder := new(Recorder)
	virtualAPI := newLabAPI(recorder)
	httpClient := &http.Client{Transport: recorder}

	options = append(options,
		hermes.WithHTTPClient(httpClient),
		hermes.WithBaseURL("https://telegram.invalid"),
		hermes.WithBotUsername(virtualAPI.Bot().Username),
	)
	lab := &Lab{
		Bot:          hermes.New("TEST_TOKEN", options...),
		API:          virtualAPI,
		tb:           tb,
		nextMessages: make(map[int64]int),
	}
	virtualAPI.lab = lab
	recorder.SetResponder(virtualAPI.respond)
	return lab
}

// PrivateUser returns a virtual user in a private chat whose ID matches the
// user ID, as it does on Telegram.
func (l *Lab) PrivateUser(id int64, username string) *Actor {
	user := labUser(id, username)
	chat := hermes.Chat{
		ID:        id,
		Type:      "private",
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
	return l.actor(user, chat)
}

// GroupUser returns a virtual user speaking in a supergroup.
func (l *Lab) GroupUser(chatID, userID int64, username string) *Actor {
	user := labUser(userID, username)
	chat := hermes.Chat{ID: chatID, Type: "supergroup", Title: "Hermes Lab Group"}
	return l.actor(user, chat)
}

func (l *Lab) actor(user hermes.User, chat hermes.Chat) *Actor {
	actor := &Actor{lab: l, user: user, chat: chat}
	if l != nil && l.API != nil {
		l.API.register(user, chat)
	}
	return actor
}

// Handle delivers an arbitrary update synchronously and returns the resulting
// handler error and outbound requests as one Step.
func (l *Lab) Handle(update *hermes.Update) *Step {
	if l == nil || l.Bot == nil || l.API == nil || l.API.recorder == nil {
		return &Step{lab: l, update: update, err: ErrLabRequired}
	}
	stepID := l.takeStepID()
	ctx := context.WithValue(context.Background(), labStepContextKey{}, stepID)
	err := l.Bot.Handle(ctx, update)
	return &Step{
		lab:      l,
		update:   update,
		err:      err,
		requests: l.API.recorder.requestsForStep(stepID),
	}
}

// Requests returns every Bot API request made since the Lab was created or
// last reset.
func (l *Lab) Requests() []Request {
	if l == nil || l.API == nil || l.API.recorder == nil {
		return nil
	}
	return l.API.recorder.Requests()
}

// Reset clears updates, messages, queued failures, and recorded requests while
// retaining the bot's registered routes and virtual actors.
func (l *Lab) Reset() {
	if l == nil {
		return
	}
	l.mu.Lock()
	l.nextUpdate = 0
	l.nextMessages = make(map[int64]int)
	l.nextCallback = 0
	l.mu.Unlock()
	if l.API != nil {
		l.API.reset()
		if l.API.recorder != nil {
			l.API.recorder.Reset()
		}
	}
}

func (l *Lab) takeMessageIDs(chatID int64, suppliedMessageID int) (int64, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.nextUpdate++
	if suppliedMessageID > l.nextMessages[chatID] {
		l.nextMessages[chatID] = suppliedMessageID
	}
	if suppliedMessageID == 0 {
		l.nextMessages[chatID]++
		suppliedMessageID = l.nextMessages[chatID]
	}
	return l.nextUpdate, suppliedMessageID
}

func (l *Lab) takeOutgoingMessageID(chatID int64) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.nextMessages[chatID]++
	return l.nextMessages[chatID]
}

func (l *Lab) takeCallbackIDs() (int64, string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.nextUpdate++
	l.nextCallback++
	return l.nextUpdate, "lab-callback-" + strconv.FormatInt(l.nextCallback, 10)
}

func (l *Lab) takeStepID() uint64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.nextStep++
	return l.nextStep
}

func labUser(id int64, username string) hermes.User {
	username = strings.TrimPrefix(strings.TrimSpace(username), "@")
	firstName := username
	if firstName == "" {
		firstName = "User " + strconv.FormatInt(id, 10)
	}
	return hermes.User{ID: id, FirstName: firstName, Username: username}
}

// Actor is one virtual Telegram user in one chat.
type Actor struct {
	lab  *Lab
	user hermes.User
	chat hermes.Chat
}

// User returns the actor's Telegram user value.
func (a *Actor) User() hermes.User {
	if a == nil {
		return hermes.User{}
	}
	return a.user
}

// Chat returns the actor's Telegram chat value.
func (a *Actor) Chat() hermes.Chat {
	if a == nil {
		return hermes.Chat{}
	}
	return a.chat
}

// Command delivers a slash command. Optional arguments are joined with one
// space, matching the text a Telegram user would send.
func (a *Actor) Command(command string, arguments ...string) *Step {
	command = strings.TrimSpace(command)
	if command == "" {
		return a.Text("/")
	}
	if !strings.HasPrefix(command, "/") {
		command = "/" + command
	}
	if len(arguments) != 0 {
		command += " " + strings.Join(arguments, " ")
	}
	return a.Text(command)
}

// Text delivers an ordinary text message.
func (a *Actor) Text(text string) *Step {
	return a.Message(hermes.Message{Text: text})
}

// Message delivers any message shape. Missing actor-owned fields such as the
// sender, chat, IDs, and date are filled deterministically; explicitly supplied
// values are preserved.
func (a *Actor) Message(message hermes.Message) *Step {
	if a == nil || a.lab == nil {
		return &Step{err: ErrLabRequired}
	}
	if message.Chat.ID == 0 {
		message.Chat = a.chat
	}
	updateID, messageID := a.lab.takeMessageIDs(message.Chat.ID, message.MessageID)
	message.MessageID = messageID
	if message.Date == 0 {
		message.Date = labEpoch + updateID
	}
	if message.From == nil {
		user := a.user
		message.From = &user
	}
	if a.lab.API != nil {
		a.lab.API.register(*message.From, message.Chat)
	}
	return a.lab.Handle(&hermes.Update{UpdateID: updateID, Message: &message})
}

// Callback presses a callback button on the latest virtual bot message in the
// actor's chat.
func (a *Actor) Callback(data string) *Step {
	if a == nil || a.lab == nil || a.lab.API == nil {
		return &Step{err: ErrLabRequired}
	}
	message, ok := a.lab.API.LastMessage(a.chat.ID)
	if !ok {
		return &Step{lab: a.lab, err: ErrMessageRequired}
	}
	return a.CallbackOn(message, data)
}

// CallbackOn presses callback data on a specific accessible message.
func (a *Actor) CallbackOn(message hermes.Message, data string) *Step {
	if a == nil || a.lab == nil {
		return &Step{err: ErrLabRequired}
	}
	if message.Chat.ID == 0 {
		message.Chat = a.chat
	}
	if a.lab.API != nil {
		a.lab.API.register(a.user, message.Chat)
	}
	updateID, callbackID := a.lab.takeCallbackIDs()
	if message.Date == 0 {
		message.Date = labEpoch + updateID
	}
	update := &hermes.Update{
		UpdateID: updateID,
		CallbackQuery: &hermes.CallbackQuery{
			ID:           callbackID,
			From:         a.user,
			Message:      hermes.AccessibleMessage(&message),
			ChatInstance: "lab-chat-" + strconv.FormatInt(message.Chat.ID, 10),
			Data:         data,
		},
	}
	return a.lab.Handle(update)
}

// Messages returns the virtual bot messages currently retained for this chat.
func (a *Actor) Messages() []hermes.Message {
	if a == nil || a.lab == nil || a.lab.API == nil {
		return nil
	}
	return a.lab.API.Messages(a.chat.ID)
}

// Step is one delivered update, its handler result, and only the Bot API
// requests caused by that update.
type Step struct {
	lab      *Lab
	update   *hermes.Update
	err      error
	requests []Request
}

// Err returns the handler error produced by the step.
func (s *Step) Err() error {
	if s == nil {
		return ErrLabRequired
	}
	return s.err
}

// Update returns the exact update delivered by the step.
func (s *Step) Update() *hermes.Update {
	if s == nil {
		return nil
	}
	return s.update
}

// Requests returns a snapshot of the Bot API requests caused by the step.
func (s *Step) Requests() []Request {
	if s == nil {
		return nil
	}
	result := make([]Request, len(s.requests))
	copy(result, s.requests)
	return result
}

// Replay delivers the same update again, including its original update ID.
// This is useful for testing deduplication and webhook retry behavior.
func (s *Step) Replay() *Step {
	if s == nil || s.lab == nil {
		return &Step{err: ErrLabRequired}
	}
	return s.lab.Handle(s.update)
}

// Want requires a successful handler and every expectation to match. A
// failure is reported through the Lab's testing object.
func (s *Step) Want(expectations ...Expectation) *Step {
	if s == nil || s.lab == nil {
		panic(ErrLabRequired)
	}
	s.lab.helper()
	if s.err != nil {
		s.lab.fatalf("Hermes Lab: unexpected handler error: %v", s.err)
		return s
	}
	for _, expectation := range expectations {
		if expectation == nil {
			continue
		}
		if err := expectation.match(s.requests); err != nil {
			s.lab.fatalf("Hermes Lab: %v", err)
			return s
		}
	}
	return s
}

// WantError requires the handler error to match target through errors.Is.
func (s *Step) WantError(target error) *Step {
	if s == nil || s.lab == nil {
		panic(ErrLabRequired)
	}
	s.lab.helper()
	if target == nil || !errors.Is(s.err, target) {
		s.lab.fatalf("Hermes Lab: handler error %v does not match %v", s.err, target)
	}
	return s
}

// WantAPIError requires a Telegram API error with the supplied error code.
func (s *Step) WantAPIError(code int) *Step {
	if s == nil || s.lab == nil {
		panic(ErrLabRequired)
	}
	s.lab.helper()
	var apiError *hermes.APIError
	if !errors.As(s.err, &apiError) || apiError.Code != code {
		s.lab.fatalf("Hermes Lab: handler error %v is not Telegram API error %d", s.err, code)
	}
	return s
}

func (l *Lab) helper() {
	if l != nil && l.tb != nil {
		l.tb.Helper()
	}
}

func (l *Lab) fatalf(format string, arguments ...any) {
	if l == nil || l.tb == nil {
		panic(fmt.Sprintf(format, arguments...))
	}
	l.tb.Fatalf(format, arguments...)
}

type labStepContextKey struct{}

const labEpoch = int64(1_700_000_000)
