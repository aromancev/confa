package profile

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	batchLimit = 100
)

type Mongo struct {
	db *mongo.Database
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		db: db,
	}
}

func (m *Mongo) Create(ctx context.Context, requests ...Profile) ([]Profile, error) {
	if len(requests) == 0 {
		return nil, errors.New("trying to create zero objects")
	}
	if len(requests) > batchLimit {
		return nil, fmt.Errorf("trying to create more than %d", batchLimit)
	}

	now := mongoNow()
	docs := make([]interface{}, len(requests))
	for i := range requests {
		requests[i].CreatedAt = now
		if err := requests[i].Validate(); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrValidation, err)
		}
		docs[i] = requests[i]
	}

	_, err := m.db.Collection("profiles").InsertMany(ctx, docs)
	switch {
	case mongo.IsDuplicateKeyError(err):
		return nil, ErrDuplicateEntry
	case err != nil:
		return nil, err
	}
	return requests, nil
}

func (m *Mongo) CreateOrUpdate(ctx context.Context, request Profile) (Profile, error) {
	if err := request.Validate(); err != nil {
		return Profile{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}

	insert := bson.M{
		"_id":       request.ID,
		"ownerId":   request.Owner,
		"createdAt": mongoNow(),
	}
	update := make(bson.M)
	if request.Handle == "" {
		insert["handle"] = request.ID.String()
	} else {
		update["handle"] = request.Handle
	}
	if request.GivenName != nil {
		update["givenName"] = request.GivenName
	}
	if request.FamilyName != nil {
		update["familyName"] = request.FamilyName
	}
	if !request.AvatarThumbnail.IsEmpty() {
		update["avatarThumbnail"] = request.AvatarThumbnail
	}
	if request.AvatarID != uuid.Nil {
		update["avatarId"] = request.AvatarID
	}

	res := m.db.Collection("profiles").FindOneAndUpdate(
		ctx,
		bson.M{
			"ownerId": request.Owner,
		},
		bson.M{
			"$setOnInsert": insert,
			"$set":         update,
		},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	)
	if err := res.Err(); err != nil {
		return Profile{}, err
	}
	var upserted Profile
	if err := res.Decode(&upserted); err != nil {
		return Profile{}, err
	}
	return upserted, nil
}

func (m *Mongo) Fetch(ctx context.Context, lookup Lookup) ([]Profile, error) {
	if err := lookup.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidation, err)
	}

	if lookup.Limit > batchLimit || lookup.Limit == 0 {
		lookup.Limit = batchLimit
	}

	cur, err := m.db.Collection("profiles").Find(
		ctx,
		mongoFilter(lookup),
		&options.FindOptions{
			Sort:  bson.M{"_id": 1},
			Limit: &lookup.Limit,
		},
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var rooms []Profile
	for cur.Next(ctx) {
		var p Profile
		err := cur.Decode(&p)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, p)
	}

	return rooms, nil
}

func (m *Mongo) FetchOne(ctx context.Context, lookup Lookup) (Profile, error) {
	confas, err := m.Fetch(ctx, lookup)
	if err != nil {
		return Profile{}, err
	}
	if len(confas) == 0 {
		return Profile{}, ErrNotFound
	}
	if len(confas) > 1 {
		return Profile{}, ErrUnexpectedResult
	}
	return confas[0], nil
}

func mongoNow() time.Time {
	// Mongodb only stores milliseconds.
	return time.Now().UTC().Round(time.Millisecond)
}

func mongoFilter(l Lookup) bson.M {
	filter := make(bson.M)
	switch {
	case l.ID != uuid.Nil:
		filter["_id"] = l.ID
	case l.From.ID != uuid.Nil:
		filter["_id"] = bson.M{
			"$gt": l.From.ID,
		}
	}
	if len(l.Owners) != 0 {
		filter["ownerId"] = bson.M{
			"$in": l.Owners,
		}
	}
	if l.Handle != "" {
		filter["handle"] = l.Handle
	}
	return filter
}
