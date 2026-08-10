package fleet

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

func (f *Fleet) openWebhookServer(
	address string,
	mounts []*mountedBot,
) (*http.Server, net.Listener, error) {
	handlers := make(map[string]http.Handler, len(mounts))
	for _, mounted := range mounts {
		if mounted.directReply {
			handlers[mounted.path] = mounted.bot.WebhookReplyHandler(mounted.webhook)
		} else {
			handlers[mounted.path] = mounted.bot.WebhookHandler(mounted.webhook)
		}
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, nil, err
	}
	server := &http.Server{
		Addr:              address,
		Handler:           exactWebhookMux(handlers),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return server, listener, nil
}

type exactWebhookMux map[string]http.Handler

func (mux exactWebhookMux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	handler := mux[request.URL.Path]
	if handler == nil {
		http.NotFound(writer, request)
		return
	}
	handler.ServeHTTP(writer, request)
}

func shutdownWebhookServer(server *http.Server, mounts []*mountedBot) error {
	if server == nil {
		return nil
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	shutdownErr := server.Shutdown(shutdownCtx)
	var closeErr error
	if shutdownErr != nil {
		closeErr = server.Close()
	}
	for _, mounted := range mounts {
		mounted.bot.Wait()
	}
	return errors.Join(shutdownErr, closeErr)
}
