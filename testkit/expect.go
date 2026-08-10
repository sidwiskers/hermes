package testkit

import (
	"fmt"
	"strings"
)

// Expectation matches the Bot API requests produced by one Lab step. Use the
// constructors in this package rather than implementing this interface.
type Expectation interface {
	match([]Request) error
}

type expectationFunc func([]Request) error

func (expectation expectationFunc) match(requests []Request) error {
	return expectation(requests)
}

// Called expects at least one call to method.
func Called(method string) Expectation {
	return expectationFunc(func(requests []Request) error {
		for _, request := range requests {
			if request.Method == method {
				return nil
			}
		}
		return fmt.Errorf("expected %s call; observed %s", method, observedMethods(requests))
	})
}

// Sent expects a sendMessage request containing exactly text.
func Sent(text string) Expectation {
	return requestFieldEquals("sendMessage", "text", text, "sent text")
}

// Edited expects an editMessageText request containing exactly text.
func Edited(text string) Expectation {
	return requestFieldEquals("editMessageText", "text", text, "edited text")
}

// Answered expects an answerCallbackQuery request containing exactly text.
func Answered(text string) Expectation {
	return requestFieldEquals("answerCallbackQuery", "text", text, "callback answer")
}

// Acknowledged expects a silent answerCallbackQuery call.
func Acknowledged() Expectation { return Answered("") }

// Deleted expects at least one deleteMessage or deleteMessages call.
func Deleted() Expectation {
	return expectationFunc(func(requests []Request) error {
		for _, request := range requests {
			if request.Method == "deleteMessage" || request.Method == "deleteMessages" {
				return nil
			}
		}
		return fmt.Errorf("expected a message deletion; observed %s", observedMethods(requests))
	})
}

// Uploaded expects method to upload filename in multipart field.
func Uploaded(method, field, filename string) Expectation {
	return expectationFunc(func(requests []Request) error {
		for _, request := range requests {
			if request.Method != method {
				continue
			}
			file, ok := request.Files[field]
			if ok && file.Name == filename {
				return nil
			}
		}
		return fmt.Errorf("expected %s upload %s=%q; observed %s", method, field, filename, observedMethods(requests))
	})
}

// Matching expects at least one request for method to satisfy predicate.
func Matching(method string, predicate func(Request) bool) Expectation {
	return expectationFunc(func(requests []Request) error {
		if predicate == nil {
			return fmt.Errorf("matching predicate for %s is nil", method)
		}
		for _, request := range requests {
			if request.Method == method && predicate(request) {
				return nil
			}
		}
		return fmt.Errorf("expected matching %s call; observed %s", method, observedMethods(requests))
	})
}

// NoCalls expects the step to make no Bot API requests.
func NoCalls() Expectation {
	return expectationFunc(func(requests []Request) error {
		if len(requests) != 0 {
			return fmt.Errorf("expected no Bot API calls; observed %s", observedMethods(requests))
		}
		return nil
	})
}

func requestFieldEquals(method, field, expected, label string) Expectation {
	return expectationFunc(func(requests []Request) error {
		for _, request := range requests {
			if request.Method == method && requestString(request, field) == expected {
				return nil
			}
		}
		return fmt.Errorf("expected %s %q through %s; observed %s", label, expected, method, observedMethods(requests))
	})
}

func observedMethods(requests []Request) string {
	if len(requests) == 0 {
		return "no calls"
	}
	methods := make([]string, len(requests))
	for index, request := range requests {
		methods[index] = request.Method
	}
	return strings.Join(methods, ", ")
}
