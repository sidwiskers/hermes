// Package fsm provides typed finite-state conversations for Hermes.
//
// A Machine uses a session.Manager for persistence and per-conversation
// serialization. Rules are explicit, ordered, and may include guards and
// actions. The state change is committed only after the selected action and
// downstream handler succeed under the session manager's commit policy.
package fsm
