package rpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/aromancev/confa/internal/proto/tracker"
	"github.com/aromancev/confa/tracker"
	"github.com/aromancev/confa/tracker/record"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	sdk "github.com/pion/ion-sdk-go"
	"github.com/twitchtv/twirp"
)

type Buckets struct {
	TrackRecords string
}

type Handler struct {
	connector *sdk.Connector
	runtime   *tracker.Runtime
	storage   *minio.Client
	emitter   *record.Beanstalk
	buckets   Buckets
}

func NewHandler(connector *sdk.Connector, runtime *tracker.Runtime, storage *minio.Client, emitter *record.Beanstalk, buckets Buckets) *Handler {
	return &Handler{
		connector: connector,
		runtime:   runtime,
		storage:   storage,
		buckets:   buckets,
		emitter:   emitter,
	}
}

func (h *Handler) Start(ctx context.Context, params *pb.StartParams) (*pb.Tracker, error) {
	var roomID uuid.UUID
	err := roomID.UnmarshalBinary(params.RoomId)
	if err != nil {
		return nil, fmt.Errorf("faield to unmarshal room id: %w", err)
	}
	expireAt := time.Now().Add(time.Duration(params.ExpireInMs) * time.Millisecond)

	var state tracker.State
	switch role := params.Role.Role.(type) {
	case *pb.Role_Record_:
		state, err = h.startRecording(ctx, roomID, expireAt, role.Record)
	default:
		return nil, fmt.Errorf("unsupported tracker role")
	}
	if err != nil {
		return nil, fmt.Errorf("faield to start tracker: %w", err)
	}

	return &pb.Tracker{
		RoomId:        params.RoomId,
		AlreadyExists: state.AlreadyExists,
		StartedAt:     state.StartedAt.UnixMilli(),
		ExpiresAt:     state.ExpiresAt.UnixMilli(),
	}, nil
}

func (h *Handler) Stop(ctx context.Context, params *pb.StopParams) (*pb.Tracker, error) {
	var roomID uuid.UUID
	err := roomID.UnmarshalBinary(params.RoomId)
	if err != nil {
		return nil, fmt.Errorf("faield to unmarshal room id: %w", err)
	}

	tr, err := h.runtime.StopTracker(
		ctx,
		roomID,
		roleRecord,
	)
	switch {
	case errors.Is(err, tracker.ErrNotFound):
		return nil, twirp.NewError(twirp.NotFound, "Tracker not found.")
	case err != nil:
		return nil, fmt.Errorf("faield to stop tracker: %w", err)
	}
	return &pb.Tracker{
		RoomId:        params.RoomId,
		AlreadyExists: tr.AlreadyExists,
		StartedAt:     tr.StartedAt.UnixMilli(),
		ExpiresAt:     tr.ExpiresAt.UnixMilli(),
	}, nil
}

func (h *Handler) startRecording(ctx context.Context, roomID uuid.UUID, expireAt time.Time, role *pb.Role_Record) (tracker.State, error) {
	var recordingID uuid.UUID
	err := recordingID.UnmarshalBinary(role.RecordingId)
	if err != nil {
		return tracker.State{}, fmt.Errorf("faield to unmarshal recording id: %w", err)
	}

	return h.runtime.StartTracker(
		ctx,
		roomID,
		roleRecord,
		expireAt,
		func(ctx context.Context, roomID uuid.UUID) (tracker.Tracker, error) {
			return record.NewTracker(ctx, h.storage, h.connector, h.emitter, h.buckets.TrackRecords, roomID, recordingID)
		},
	)
}

const (
	roleRecord = "record"
)
