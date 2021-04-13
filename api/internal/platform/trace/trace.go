package trace

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
)

type ctxKey struct{}

const (
	header = "Trace-Id"
)

func New(ctx context.Context, trace string) context.Context {
	return context.WithValue(ctx, ctxKey{}, trace)
}

func Ctx(ctx context.Context) (context.Context, string) {
	var trace string
	if t, ok := ctx.Value(ctxKey{}).(string); ok {
		trace = t
	} else {
		trace = uuid.New().String()
	}
	l := log.Logger.With().Str("traceId", trace).Logger()
	ctx = l.WithContext(ctx)
	return New(ctx, trace), trace
}

func Job(ctx context.Context, job *beanstalk.Job) (context.Context, string) {
	var j traceJob
	err := json.Unmarshal(job.Body, &j)
	if err == nil && j.Trace != "" {
		ctx = New(ctx, j.Trace)
		job.Body = j.Body
	} else {
		log.Ctx(ctx).Warn().Msg("failed to parse trace from job")
	}
	return Ctx(ctx)
}

func WriteHeader(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx, traceID := Ctx(r.Context())
		h(w, r.WithContext(ctx), ps)
		w.Header().Set(header, traceID)
	}
}

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type Beanstalkd struct {
	producer Producer
}

func NewBeanstalkd(producer Producer) *Beanstalkd {
	return &Beanstalkd{producer: producer}
}

func (b *Beanstalkd) Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error) {
	ctx, trace := Ctx(ctx)
	traceBody, err := json.Marshal(traceJob{
		Trace: trace,
		Body:  body,
	})
	if err == nil {
		body = traceBody
	} else {
		log.Ctx(ctx).Warn().Msg("failed to marshal job with trace")
	}
	return b.producer.Put(ctx, tube, body, params)
}

type traceJob struct {
	Trace string `json:"traceId"`
	Body  json.RawMessage
}
