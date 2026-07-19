package api

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

func isNilUnion(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func validateChatID(chatID any, method string) error {
	if chatID == nil {
		return fmt.Errorf("hermes: %s chat_id is required", method)
	}
	if value, ok := chatID.(string); ok && strings.TrimSpace(value) == "" {
		return fmt.Errorf("hermes: %s chat_id is required", method)
	}
	return nil
}

func callMessage(ctx context.Context, b *Client, method string, params any) (*Message, error) {
	var message Message
	if err := b.Call(ctx, method, params, &message); err != nil {
		return nil, err
	}
	return &message, nil
}

func callMessageOrBool(ctx context.Context, b *Client, method string, params any) (*Message, error) {
	var result messageOrBool
	if err := b.Call(ctx, method, params, &result); err != nil {
		return nil, err
	}
	return result.Message, nil
}
