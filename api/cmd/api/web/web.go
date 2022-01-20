package web

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/google/uuid"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/confa/talk"
	"github.com/aromancev/confa/internal/platform/trace"
)

//go:generate gqlgen

type Code string

const (
	CodeBadRequest       = "BAD_REQUEST"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeDuplicateEntry   = "DUPLICATE_ENTRY"
	CodeNotFound         = "NOT_FOUND"
	CodePermissionDenied = "PERMISSION_DENIED"
	CodeUnknown          = "UNKNOWN_CODE"
)

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type Handler struct {
	router http.Handler
}

func NewHandler(resolver *Resolver) *Handler {
	r := http.NewServeMux()

	r.HandleFunc("/health", ok)
	r.Handle("/query",
		withHTTPAuth(
			handler.NewDefaultServer(
				NewExecutableSchema(Config{Resolvers: resolver}),
			),
		),
	)
	r.HandleFunc(
		"/rtc/ws/",
		withWSockAuthFunc(
			serveRTC(resolver.rooms, resolver.publicKey, resolver.upgrader, resolver.sfuConn, resolver.producer, resolver.eventWatcher),
		),
	)
	r.HandleFunc("/dev/", playground.Handler("API playground", "/api/query"))

	return &Handler{
		router: r,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, traceID := trace.Ctx(r.Context())
	w.Header().Set("Trace-Id", traceID)

	defer func() {
		if err := recover(); err != nil {
			log.Ctx(ctx).Error().Str("error", fmt.Sprint(err)).Msg("ServeHTTP panic")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	lw := newResponseWriter(w)
	r = r.WithContext(ctx)
	h.router.ServeHTTP(lw, r)

	lw.Event(ctx, r).Msg("Web served")
}

func ok(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

func newError(code Code, message string) *gqlerror.Error {
	return &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
}

func newInternalError() *gqlerror.Error {
	return &gqlerror.Error{
		Message: "internal system error",
		Extensions: map[string]interface{}{
			"code": CodeUnknown,
		},
	}
}

type responseWriter struct {
	http.ResponseWriter
	code int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, code: http.StatusOK}
}

func (w *responseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	panic("ResponseWriter does not implement http.Hijacker")
}

func (w *responseWriter) Event(ctx context.Context, r *http.Request) *zerolog.Event {
	var event *zerolog.Event
	if w.code >= http.StatusInternalServerError {
		event = log.Ctx(ctx).Error()
	} else {
		event = log.Ctx(ctx).Info()
	}
	return event.Str("method", r.Method).Int("code", w.code).Str("url", r.URL.String())
}

func newConfaLookup(input ConfaLookup, limit int, from *string) (confa.Lookup, error) {
	if limit <= 0 || limit > batchLimit {
		limit = batchLimit
	}

	lookup := confa.Lookup{
		Limit: int64(limit),
	}
	var err error
	if from != nil {
		lookup.From, err = uuid.Parse(*from)
		if err != nil {
			return confa.Lookup{}, err
		}
	}
	if input.ID != nil {
		lookup.ID, err = uuid.Parse(*input.ID)
		if err != nil {
			return confa.Lookup{}, err
		}
	}
	if input.OwnerID != nil {
		lookup.Owner, err = uuid.Parse(*input.OwnerID)
		if err != nil {
			return confa.Lookup{}, err
		}
	}
	if input.Handle != nil {
		lookup.Handle = *input.Handle
	}
	return lookup, nil
}

func newConfa(c confa.Confa) *Confa {
	return &Confa{
		ID:          c.ID.String(),
		OwnerID:     c.Owner.String(),
		Handle:      c.Handle,
		Title:       c.Title,
		Description: c.Description,
	}
}

func newTalkLookup(input TalkLookup, limit int, from *string) (talk.Lookup, error) {
	if limit < 0 || limit > batchLimit {
		limit = batchLimit
	}
	lookup := talk.Lookup{
		Limit: int64(limit),
	}
	var err error
	if from != nil {
		lookup.From, err = uuid.Parse(*from)
		if err != nil {
			return talk.Lookup{}, err
		}
	}
	if input.ID != nil {
		lookup.ID, err = uuid.Parse(*input.ID)
		if err != nil {
			return talk.Lookup{}, err
		}
	}
	if input.ConfaID != nil {
		lookup.Confa, err = uuid.Parse(*input.ConfaID)
		if err != nil {
			return talk.Lookup{}, err
		}
	}
	if input.OwnerID != nil {
		lookup.Owner, err = uuid.Parse(*input.OwnerID)
		if err != nil {
			return talk.Lookup{}, err
		}
	}
	if input.SpeakerID != nil {
		lookup.Speaker, err = uuid.Parse(*input.SpeakerID)
		if err != nil {
			return talk.Lookup{}, err
		}
	}
	if input.Handle != nil {
		lookup.Handle = *input.Handle
	}
	return lookup, nil
}

func newTalk(t talk.Talk) *Talk {
	return &Talk{
		ID:          t.ID.String(),
		ConfaID:     t.Confa.String(),
		OwnerID:     t.Owner.String(),
		SpeakerID:   t.Speaker.String(),
		RoomID:      t.Room.String(),
		Handle:      t.Handle,
		Title:       t.Title,
		Description: t.Description,
	}
}

const (
	batchLimit = 100
)
