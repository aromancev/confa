package confa

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TestCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		pg, done := double.NewDocker("", migrate)
		defer done()
		crud := NewCRUD(pg, NewSQL())
		userID := uuid.New()
		request := Confa{
			Handle: "test",
		}

		confa, err := crud.Create(ctx, userID, request)
		require.NoError(t, err)
		fetchedConfa, err := crud.Fetch(ctx, confa.ID)
		require.NoError(t, err)
		require.Equal(t, confa, fetchedConfa)
	})
}
