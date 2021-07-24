package web

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"time"

	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/confa/talk"
	"github.com/aromancev/confa/internal/confa/talk/clap"
	"github.com/aromancev/confa/internal/emails"
	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/user"
	"github.com/aromancev/confa/proto/queue"
	"github.com/google/uuid"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
)

func (r *mutationResolver) Login(ctx context.Context, address string) (string, error) {
	if err := email.Validate(address); err != nil {
		return "", newError(CodeBadRequest, "Invalid email.")
	}

	token, err := r.secretKey.Sign(auth.NewEmailClaims(address))
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to create email token.")
		return "", newError(CodeInternal, "")
	}

	msg, err := emails.Login(r.baseURL, address, token)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to render login email.")
		return "", newError(CodeInternal, "")
	}

	body, err := queue.Marshal(&queue.EmailJob{
		Emails: []*queue.Email{{
			FromName:  msg.FromName,
			ToAddress: msg.ToAddress,
			Subject:   msg.Subject,
			Html:      msg.HTML,
		}},
	}, trace.ID(ctx))
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to marshal email.")
		return "", newError(CodeInternal, "")
	}

	id, err := r.producer.Put(ctx, queue.TubeEmail, body, beanstalk.PutParams{TTR: 10 * time.Second})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to put email job.")
		return "", newError(CodeInternal, "")
	}
	log.Ctx(ctx).Info().Uint64("jobId", id).Msg("Email login job emitted.")
	return address, nil
}

func (r *mutationResolver) CreateSession(ctx context.Context, emailToken string) (*Token, error) {
	var claims auth.EmailClaims
	err := r.publicKey.Verify(emailToken, &claims)
	if err != nil {
		return nil, newError(CodeUnauthorized, "Wrong email token.")
	}

	usr, err := r.users.GetOrCreate(ctx, user.User{
		Idents: []user.Ident{
			{Platform: user.PlatformEmail,
				Value: claims.Address,
			},
		},
	})
	if err != nil {
		return nil, newError(CodeInternal, "")
	}

	sess, err := r.sessions.Create(ctx, usr.ID)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to create session.")
		return nil, newError(CodeInternal, "")
	}

	apiClaims := auth.NewAPIClaims(sess.Owner)
	access, err := r.secretKey.Sign(apiClaims)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to sign access token.")
		return nil, newError(CodeInternal, "")
	}

	auth.Ctx(ctx).SetSession(sess.Key)
	return &Token{
		Token:     access,
		ExpiresIn: int(apiClaims.ExpiresIn().Seconds()),
	}, nil
}

func (r *mutationResolver) CreateConfa(ctx context.Context, handle *string) (*Confa, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return nil, newError(CodeUnauthorized, "Invalid access token.")
	}

	var req confa.Confa
	if handle != nil {
		req.Handle = *handle
	}
	created, err := r.confas.Create(ctx, claims.UserID, req)
	switch {
	case errors.Is(err, confa.ErrValidation):
		return nil, newError(CodeBadRequest, err.Error())
	case errors.Is(err, confa.ErrDuplicateEntry):
		return nil, newError(CodeDuplicateEntry, err.Error())
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to create confa.")
		return nil, newError(CodeInternal, "")
	}

	return &Confa{
		ID:      created.ID.String(),
		OwnerID: created.Owner.String(),
		Handle:  created.Handle,
	}, nil
}

func (r *mutationResolver) CreateTalk(ctx context.Context, confaID string, handle *string) (*Talk, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return nil, newError(CodeUnauthorized, "Invalid access token.")
	}

	var req talk.Talk
	if handle != nil {
		req.Handle = *handle
	}
	var err error
	req.Confa, err = uuid.Parse(confaID)
	if err != nil {
		return nil, newError(CodeNotFound, "Confa not found.")
	}
	created, err := r.talks.Create(ctx, claims.UserID, req)
	switch {
	case errors.Is(err, talk.ErrValidation):
		return nil, newError(CodeBadRequest, err.Error())
	case errors.Is(err, talk.ErrDuplicateEntry):
		return nil, newError(CodeDuplicateEntry, err.Error())
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("failed to create talk.")
		return nil, newError(CodeInternal, "")
	}

	return &Talk{
		ID:        created.ID.String(),
		ConfaID:   created.Confa.String(),
		OwnerID:   created.Owner.String(),
		SpeakerID: created.Speaker.String(),
		Handle:    created.Handle,
	}, nil
}

func (r *mutationResolver) CreateClap(ctx context.Context, talkID string, value int) (string, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return "", newError(CodeUnauthorized, "Invalid access token.")
	}

	tID, err := uuid.Parse(talkID)
	if err != nil {
		return "", newError(CodeNotFound, "Talk not found.")
	}
	id, err := r.claps.CreateOrUpdate(ctx, claims.UserID, tID, uint(value))
	switch {
	case errors.Is(err, clap.ErrValidation):
		return "", newError(CodeBadRequest, err.Error())
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to create clap.")
		return "", newError(CodeInternal, "")
	}

	return id.String(), nil
}

func (r *queryResolver) Token(ctx context.Context) (*Token, error) {
	sessKey := auth.Ctx(ctx).Session()
	if sessKey == "" {
		return nil, newError(CodeUnauthorized, "Session is not present.")
	}
	sess, err := r.sessions.Fetch(ctx, sessKey)
	if err != nil {
		return nil, newError(CodeUnauthorized, "Invalid session.")
	}

	claims := auth.NewAPIClaims(sess.Owner)
	access, err := r.secretKey.Sign(claims)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to sign API token.")
		return nil, newError(CodeInternal, "")
	}
	return &Token{
		Token:     access,
		ExpiresIn: int(claims.ExpiresIn().Seconds()),
	}, nil
}

func (r *queryResolver) Confas(ctx context.Context, where ConfaInput, from string, limit int) (*Confas, error) {
	if limit < 0 || limit > batchLimit {
		limit = batchLimit
	}

	lookup := confa.Lookup{
		Limit: int64(limit),
	}
	var err error
	if from != "" {
		lookup.From, err = uuid.Parse(from)
		if err != nil {
			return &Confas{Limit: limit}, nil
		}
	}
	if where.ID != nil {
		lookup.ID, err = uuid.Parse(*where.ID)
		if err != nil {
			return &Confas{Limit: limit}, nil
		}
	}
	if where.OwnerID != nil {
		lookup.Owner, err = uuid.Parse(*where.OwnerID)
		if err != nil {
			return &Confas{Limit: limit}, nil
		}
	}
	if where.Handle != nil {
		lookup.Handle = *where.Handle
	}

	confas, err := r.confas.Fetch(ctx, lookup)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to fetch confa.")
		return nil, newError(CodeInternal, "")
	}

	if len(confas) == 0 {
		return &Confas{Limit: limit}, nil
	}

	res := &Confas{
		Items:    make([]*Confa, len(confas)),
		Limit:    limit,
		NextFrom: confas[len(confas)-1].ID.String(),
	}
	for i, c := range confas {
		res.Items[i] = &Confa{
			ID:      c.ID.String(),
			OwnerID: c.Owner.String(),
			Handle:  c.Handle,
		}
	}

	return res, nil
}

func (r *queryResolver) Talks(ctx context.Context, where TalkInput, from string, limit int) (*Talks, error) {
	if limit < 0 || limit > batchLimit {
		limit = batchLimit
	}
	lookup := talk.Lookup{
		Limit: int64(limit),
	}
	var err error
	if from != "" {
		lookup.From, err = uuid.Parse(from)
		if err != nil {
			return &Talks{Limit: limit}, nil
		}
	}
	if where.ID != nil {
		lookup.ID, err = uuid.Parse(*where.ID)
		if err != nil {
			return &Talks{Limit: limit}, nil
		}
	}
	if where.ConfaID != nil {
		lookup.Confa, err = uuid.Parse(*where.ConfaID)
		if err != nil {
			return &Talks{Limit: limit}, nil
		}
	}
	if where.OwnerID != nil {
		lookup.Owner, err = uuid.Parse(*where.OwnerID)
		if err != nil {
			return &Talks{Limit: limit}, nil
		}
	}
	if where.SpeakerID != nil {
		lookup.Speaker, err = uuid.Parse(*where.SpeakerID)
		if err != nil {
			return &Talks{Limit: limit}, nil
		}
	}
	if where.Handle != nil {
		lookup.Handle = *where.Handle
	}
	talks, err := r.talks.Fetch(ctx, lookup)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to fetch talk.")
		return nil, newError(CodeInternal, "")
	}

	if len(talks) == 0 {
		return &Talks{Limit: limit}, nil
	}

	res := &Talks{
		Items:    make([]*Talk, len(talks)),
		Limit:    limit,
		NextFrom: talks[len(talks)-1].ID.String(),
	}
	for i, t := range talks {
		res.Items[i] = &Talk{
			ID:        t.ID.String(),
			ConfaID:   t.Confa.String(),
			OwnerID:   t.Owner.String(),
			SpeakerID: t.Speaker.String(),
			Handle:    t.Handle,
		}
	}
	return res, nil
}

func (r *queryResolver) AggregateClaps(ctx context.Context, where ClapInput) (int, error) {
	var lookup clap.Lookup
	var err error
	if where.ConfaID != nil {
		lookup.Confa, err = uuid.Parse(*where.ConfaID)
		if err != nil {
			return 0, nil
		}
	}
	if where.SpeakerID != nil {
		lookup.Speaker, err = uuid.Parse(*where.SpeakerID)
		if err != nil {
			return 0, nil
		}
	}
	if where.TalkID != nil {
		lookup.Talk, err = uuid.Parse(*where.TalkID)
		if err != nil {
			return 0, nil
		}
	}
	claps, err := r.claps.Aggregate(ctx, lookup)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to aggregate claps.")
		return 0, newError(CodeInternal, "")
	}
	return int(claps), nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver } // nolint: gocritic
type queryResolver struct{ *Resolver }    // nolint: gocritic
