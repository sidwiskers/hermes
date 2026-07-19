package fsm

import (
	"context"
	"errors"
	"testing"

	"github.com/sidwiskers/hermes/framework"
	"github.com/sidwiskers/hermes/session"
	"github.com/sidwiskers/hermes/types"
)

type testState string

const (
	stateStart testState = "start"
	stateName  testState = "name"
	stateDone  testState = "done"
)

type testData struct {
	Name string
}

func TestMachineTransitionsAndPersistsData(t *testing.T) {
	store := session.NewMemory[Snapshot[testState, testData]](0)
	sessions := session.New(store, session.ByUser)
	machine := New(sessions, stateStart)
	if err := machine.Add(Rule[testState, testData]{From: stateStart, Event: "begin", To: stateName}); err != nil {
		t.Fatal(err)
	}
	if err := machine.Add(Rule[testState, testData]{
		From: stateName, Event: "submit", To: stateDone,
		Guard: func(_ *framework.Context, current Snapshot[testState, testData]) bool {
			return current.Data.Name == ""
		},
		Action: func(_ *framework.Context, data *testData) error {
			data.Name = "Ada"
			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}

	ctx := testContext(42)
	handler := machine.Middleware()(func(c *framework.Context) error {
		if err := machine.Trigger(c, "begin"); err != nil {
			return err
		}
		if state, err := machine.State(c); err != nil || state != stateName {
			t.Fatalf("state after begin=%q err=%v", state, err)
		}
		return machine.Trigger(c, "submit")
	})
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}

	inspect := machine.Middleware()(func(c *framework.Context) error {
		value, exists, err := machine.Snapshot(c)
		if err != nil || !exists || value.State != stateDone || value.Data.Name != "Ada" {
			t.Fatalf("snapshot=%+v exists=%v err=%v", value, exists, err)
		}
		return nil
	})
	if err := inspect(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestMachineGuardFallbackAndTransitionError(t *testing.T) {
	store := session.NewMemory[Snapshot[testState, testData]](0)
	machine := New(session.New(store, session.ByUser), stateStart)
	if err := machine.Add(Rule[testState, testData]{
		From: stateStart, Event: "advance", To: stateName,
		Guard: func(*framework.Context, Snapshot[testState, testData]) bool { return false },
	}); err != nil {
		t.Fatal(err)
	}
	if err := machine.AddAny("advance", stateDone, nil, nil); err != nil {
		t.Fatal(err)
	}
	handler := machine.Middleware()(func(c *framework.Context) error {
		if err := machine.Trigger(c, "advance"); err != nil {
			return err
		}
		if state, _ := machine.State(c); state != stateDone {
			t.Fatalf("state=%q", state)
		}
		err := machine.Trigger(c, "missing")
		var transitionErr *TransitionError[testState]
		if !errors.Is(err, ErrNoTransition) || !errors.As(err, &transitionErr) {
			t.Fatalf("transition error=%v", err)
		}
		return nil
	})
	if err := handler(testContext(42)); err != nil {
		t.Fatal(err)
	}
}

func TestMachineActionAndDownstreamErrorsRollback(t *testing.T) {
	store := session.NewMemory[Snapshot[testState, testData]](0)
	machine := New(session.New(store, session.ByUser), stateStart)
	wantErr := errors.New("action failed")
	if err := machine.Add(Rule[testState, testData]{
		From: stateStart, Event: "fail", To: stateDone,
		Action: func(*framework.Context, *testData) error { return wantErr },
	}); err != nil {
		t.Fatal(err)
	}
	handler := machine.Middleware()(machine.Handle("fail"))
	if err := handler(testContext(42)); !errors.Is(err, wantErr) {
		t.Fatalf("error=%v", err)
	}
	if store.Len() != 0 {
		t.Fatal("failed action committed a session")
	}

	if err := machine.Add(Rule[testState, testData]{From: stateStart, Event: "next", To: stateDone}); err != nil {
		t.Fatal(err)
	}
	downstreamErr := errors.New("downstream failed")
	handler = machine.Middleware()(machine.Then("next", func(*framework.Context) error { return downstreamErr }))
	if err := handler(testContext(42)); !errors.Is(err, downstreamErr) {
		t.Fatalf("error=%v", err)
	}
	if store.Len() != 0 {
		t.Fatal("downstream error committed a transition")
	}
}

func TestMachineInitialFilterSetAndReset(t *testing.T) {
	store := session.NewMemory[Snapshot[testState, testData]](0)
	machine := New(session.New(store, session.ByUser), stateStart)
	handler := machine.Middleware()(func(c *framework.Context) error {
		if !machine.In(stateStart)(c) {
			t.Fatal("initial state filter did not match")
		}
		if err := machine.SetState(c, stateName); err != nil {
			return err
		}
		if !machine.In(stateName)(c) {
			t.Fatal("set state filter did not match")
		}
		return machine.Reset(c)
	})
	if err := handler(testContext(42)); err != nil {
		t.Fatal(err)
	}
	if store.Len() != 0 {
		t.Fatal("reset did not delete session")
	}
}

func testContext(userID int64) *framework.Context {
	update := &types.Update{Message: &types.Message{
		Chat: types.Chat{ID: 1},
		From: &types.User{ID: userID},
	}}
	return framework.NewContext(context.Background(), nil, update, "")
}
