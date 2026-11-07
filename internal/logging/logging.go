package logging

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"time"
)

const (
	alphabet        = "0123456789abcdefghjkmnpqrstvwxyz"
	requestIdLength = 10
)

type requestIDKey struct{}
type timeStartKey struct{}

type ContextHandler struct {
	slog.Handler
}

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if requestID, ok := ctx.Value(requestIDKey{}).(string); ok {
		r.AddAttrs(slog.String("request_id", requestID))
	}
	return h.Handler.Handle(ctx, r)
}

var LogLevel = &slog.LevelVar{}

func SetLevel(loglevel int64) {
	LogLevel.Set(slog.Level(loglevel))
}

func ResetLevel() {
	LogLevel.Set(slog.LevelInfo)
}

func (h ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ContextHandler{h.Handler.WithAttrs(attrs)}
}

func (h ContextHandler) WithGroup(name string) slog.Handler {
	return ContextHandler{h.Handler.WithGroup(name)}
}

func StartLog(r *http.Request) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, requestIDKey{}, generateRequestID())
	ctx = context.WithValue(ctx, timeStartKey{}, time.Now())

	return r.WithContext(ctx)

}

func EndLog(r *http.Request, status int) {
	ctx := r.Context()
	start, ok := ctx.Value(timeStartKey{}).(time.Time)
	if !ok {
		// fallback so don't panic if middleware missed
		slog.InfoContext(ctx, "request completed")
		return
	}
	slog.InfoContext(ctx, "request completed",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Int("status", status),
		slog.Duration("duration", time.Since(start)),
	)
}

func SetRequestID(r *http.Request, reqID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), requestIDKey{}, reqID))
}

func GetRequestID(r *http.Request) string {
	v, ok := r.Context().Value(requestIDKey{}).(string)
	if !ok {
		return ""
	}
	return v
}

func init() {
	LogLevel.Set(slog.LevelInfo)
	baseHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: LogLevel})
	ctxHandler := ContextHandler{baseHandler}
	logger := slog.New(ctxHandler)
	slog.SetDefault(logger)
}

func generateRequestID() string {
	b := make([]byte, requestIdLength)
	for i := range b {
		b[i] = alphabet[rand.IntN(len(alphabet))]
	}
	return string(b)
}
