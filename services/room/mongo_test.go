package room

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMongo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		rooms := NewMongo(dockerMongo(t))

		request := Room{
			ID:    uuid.New(),
			Owner: uuid.New(),
		}
		created, err := rooms.Create(ctx, request)
		require.NoError(t, err)
		assert.NotZero(t, created[0].CreatedAt)

		fetched, err := rooms.Fetch(ctx, Lookup{
			ID:    request.ID,
			Owner: request.Owner,
		})
		require.NoError(t, err)
		assert.Equal(t, created, fetched)
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		rooms := NewMongo(dockerMongo(t))

		room := Room{
			ID:    uuid.New(),
			Owner: uuid.New(),
		}
		created, err := rooms.Create(ctx, room)
		require.NoError(t, err)
		_, err = rooms.Create(
			ctx,
			Room{
				ID:    uuid.New(),
				Owner: uuid.New(),
			},
			Room{
				ID:    uuid.New(),
				Owner: uuid.New(),
			},
		)
		require.NoError(t, err)

		t.Run("by id", func(t *testing.T) {
			fetched, err := rooms.Fetch(ctx, Lookup{
				ID: room.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("by owner", func(t *testing.T) {
			fetched, err := rooms.Fetch(ctx, Lookup{
				Owner: room.Owner,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("with limit and offset", func(t *testing.T) {
			fetched, err := rooms.Fetch(ctx, Lookup{
				Limit: 1,
			})
			require.NoError(t, err)
			assert.Equal(t, 1, len(fetched))

			// 3 in total, skipped one.
			fetched, err = rooms.Fetch(ctx, Lookup{
				From: fetched[0].ID,
			})
			require.NoError(t, err)
			assert.Equal(t, 2, len(fetched))
		})
	})
}
