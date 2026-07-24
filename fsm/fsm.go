package fsm

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/sidwiskers/hermes/framework"
	"github.com/sidwiskers/hermes/session"
)

var (
	// ErrSessionRequired reports a Machine without a session manager.
	ErrSessionRequired = errors.New("hermes/fsm: session manager is required")
	// ErrEventRequired reports a transition with an empty event.
	ErrEventRequired = errors.New("hermes/fsm: event is required")
	// ErrNoTransition is wrapped by TransitionError when no guard accepts an
	// event in the current state.
	ErrNoTransition = errors.New("hermes/fsm: no transition")
)

// Snapshot is the complete persistent value for one conversation.
type Snapshot[S comparable, D any] struct {
	State S
	Data  D
}

// Guard decides whether a rule applies. Guards should not mutate Snapshot;
// put mutations in the rule Action so they remain transactional.
type Guard[S comparable, D any] func(*framework.Context, Snapshot[S, D]) bool

// Action runs before a state change. Returning an error cancels the transition.
type Action[S comparable, D any] func(*framework.Context, *D) error

// Rule describes one ordered transition.
type Rule[S comparable, D any] struct {
	From   S
	Event  string
	To     S
	Guard  Guard[S, D]
	Action Action[S, D]
}

// TransitionError records the rejected state and event.
type TransitionError[S comparable] struct {
	State S
	Event string
}

// Error implements error.
func (e *TransitionError[S]) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("hermes/fsm: no transition from %v for event %q", e.State, e.Event)
}

// Unwrap lets errors.Is identify ErrNoTransition.
func (e *TransitionError[S]) Unwrap() error { return ErrNoTransition }

type transitionKey[S comparable] struct {
	state S
	event string
}

type transitionTable[S comparable, D any] struct {
	rules map[transitionKey[S]][]Rule[S, D]
	any   map[string][]Rule[S, D]
}

// Machine is a typed, concurrency-safe transition table backed by sessions.
// Registration and dispatch may run concurrently; rules with the same state
// and event are evaluated in registration order.
type Machine[S comparable, D any] struct {
	sessions *session.Manager[Snapshot[S, D]]
	initial  S

	mu      sync.Mutex
	table   atomic.Pointer[transitionTable[S, D]]
	startup *transitionTable[S, D]
	started atomic.Bool
}

// New creates a machine whose missing sessions begin at initial.
func New[S comparable, D any](sessions *session.Manager[Snapshot[S, D]], initial S) *Machine[S, D] {
	machine := &Machine[S, D]{
		sessions: sessions,
		initial:  initial,
	}
	machine.startup = new(transitionTable[S, D])
	return machine
}

// Middleware installs the machine's session manager.
func (m *Machine[S, D]) Middleware() framework.Middleware {
	if m == nil || m.sessions == nil {
		return func(framework.Handler) framework.Handler {
			return func(*framework.Context) error { return ErrSessionRequired }
		}
	}
	return m.sessions.Middleware()
}

// Add appends a state-specific rule.
func (m *Machine[S, D]) Add(rule Rule[S, D]) error {
	if m == nil || m.sessions == nil {
		return ErrSessionRequired
	}
	if rule.Event == "" {
		return ErrEventRequired
	}
	m.mu.Lock()
	key := transitionKey[S]{state: rule.From, event: rule.Event}
	if !m.started.Load() {
		if m.startup.rules == nil {
			m.startup.rules = make(map[transitionKey[S]][]Rule[S, D])
		}
		m.startup.rules[key] = append(m.startup.rules[key], rule)
		m.mu.Unlock()
		return nil
	}
	next := cloneTransitionTable(m.table.Load())
	if next.rules == nil {
		next.rules = make(map[transitionKey[S]][]Rule[S, D])
	}
	next.rules[key] = append(append([]Rule[S, D](nil), next.rules[key]...), rule)
	m.table.Store(next)
	m.mu.Unlock()
	return nil
}

// AddAny appends a fallback rule that is considered from every state after
// state-specific rules for the event.
func (m *Machine[S, D]) AddAny(event string, to S, guard Guard[S, D], action Action[S, D]) error {
	if m == nil || m.sessions == nil {
		return ErrSessionRequired
	}
	if event == "" {
		return ErrEventRequired
	}
	rule := Rule[S, D]{Event: event, To: to, Guard: guard, Action: action}
	m.mu.Lock()
	if !m.started.Load() {
		if m.startup.any == nil {
			m.startup.any = make(map[string][]Rule[S, D])
		}
		m.startup.any[event] = append(m.startup.any[event], rule)
		m.mu.Unlock()
		return nil
	}
	next := cloneTransitionTable(m.table.Load())
	if next.any == nil {
		next.any = make(map[string][]Rule[S, D])
	}
	next.any[event] = append(append([]Rule[S, D](nil), next.any[event]...), rule)
	m.table.Store(next)
	m.mu.Unlock()
	return nil
}

// Snapshot returns the current conversation value. Missing sessions use the
// initial state and D's zero value and report exists=false.
func (m *Machine[S, D]) Snapshot(c *framework.Context) (Snapshot[S, D], bool, error) {
	if m == nil || m.sessions == nil {
		return Snapshot[S, D]{}, false, ErrSessionRequired
	}
	value, exists, err := m.sessions.Get(c)
	if err != nil {
		return Snapshot[S, D]{}, false, err
	}
	if !exists {
		value.State = m.initial
	}
	return value, exists, nil
}

// State returns the current state, using the initial state for a missing
// session.
func (m *Machine[S, D]) State(c *framework.Context) (S, error) {
	snapshot, _, err := m.Snapshot(c)
	return snapshot.State, err
}

// Set replaces the complete conversation snapshot.
func (m *Machine[S, D]) Set(c *framework.Context, value Snapshot[S, D]) error {
	if m == nil || m.sessions == nil {
		return ErrSessionRequired
	}
	return m.sessions.Set(c, value)
}

// SetState changes only the state and preserves conversation data.
func (m *Machine[S, D]) SetState(c *framework.Context, state S) error {
	value, _, err := m.Snapshot(c)
	if err != nil {
		return err
	}
	value.State = state
	return m.sessions.Set(c, value)
}

// Reset deletes the persisted conversation. Its next update starts at the
// initial state with zero data.
func (m *Machine[S, D]) Reset(c *framework.Context) error {
	if m == nil || m.sessions == nil {
		return ErrSessionRequired
	}
	return m.sessions.Delete(c)
}

// Trigger selects and executes the first accepted rule for event. An action
// error leaves the state unchanged.
func (m *Machine[S, D]) Trigger(c *framework.Context, event string) error {
	if event == "" {
		return ErrEventRequired
	}
	current, _, err := m.Snapshot(c)
	if err != nil {
		return err
	}
	exact, fallback := m.rulesFor(current.State, event)
	if matched, err := m.tryRules(c, current, exact); matched {
		return err
	}
	if matched, err := m.tryRules(c, current, fallback); matched {
		return err
	}
	return &TransitionError[S]{State: current.State, Event: event}
}

func (m *Machine[S, D]) tryRules(
	c *framework.Context,
	current Snapshot[S, D],
	rules []Rule[S, D],
) (bool, error) {
	for _, rule := range rules {
		if rule.Guard != nil && !rule.Guard(c, current) {
			continue
		}
		next := current
		if rule.Action != nil {
			if err := rule.Action(c, &next.Data); err != nil {
				return true, err
			}
		}
		next.State = rule.To
		return true, m.sessions.Set(c, next)
	}
	return false, nil
}

// Handle returns a handler that triggers event.
func (m *Machine[S, D]) Handle(event string) framework.Handler {
	return func(c *framework.Context) error { return m.Trigger(c, event) }
}

// Then returns a handler that triggers event and invokes next only after a
// successful transition. A later error participates in the session manager's
// commit policy.
func (m *Machine[S, D]) Then(event string, next framework.Handler) framework.Handler {
	return func(c *framework.Context) error {
		if err := m.Trigger(c, event); err != nil {
			return err
		}
		if next == nil {
			return nil
		}
		return next(c)
	}
}

// In returns a filter matching any supplied current state. It returns false
// when used outside the machine middleware.
func (m *Machine[S, D]) In(states ...S) framework.Filter {
	set := make(map[S]struct{}, len(states))
	for _, state := range states {
		set[state] = struct{}{}
	}
	return func(c *framework.Context) bool {
		state, err := m.State(c)
		if err != nil {
			return false
		}
		_, ok := set[state]
		return ok
	}
}

func (m *Machine[S, D]) rulesFor(state S, event string) (exact, fallback []Rule[S, D]) {
	if m == nil {
		return nil, nil
	}
	m.ensureStarted()
	table := m.table.Load()
	if table == nil {
		return nil, nil
	}
	return table.rules[transitionKey[S]{state: state, event: event}], table.any[event]
}

func (m *Machine[S, D]) ensureStarted() {
	if m == nil || m.started.Load() {
		return
	}
	m.mu.Lock()
	if !m.started.Load() {
		table := m.startup
		if table == nil {
			table = new(transitionTable[S, D])
		}
		m.table.Store(table)
		m.startup = nil
		m.started.Store(true)
	}
	m.mu.Unlock()
}

func cloneTransitionTable[S comparable, D any](source *transitionTable[S, D]) *transitionTable[S, D] {
	target := new(transitionTable[S, D])
	if source == nil {
		return target
	}
	if len(source.rules) != 0 {
		target.rules = make(map[transitionKey[S]][]Rule[S, D], len(source.rules))
		for key, rules := range source.rules {
			target.rules[key] = rules
		}
	}
	if len(source.any) != 0 {
		target.any = make(map[string][]Rule[S, D], len(source.any))
		for event, rules := range source.any {
			target.any[event] = rules
		}
	}
	return target
}
