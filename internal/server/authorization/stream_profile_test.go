package authorization

import (
	"testing"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/stretchr/testify/require"
)

func TestStreamKeyWithUnderscore(t *testing.T) {
	t.Setenv(environment.StreamProfilePath, t.TempDir())

	_, err := CreateProfile("my_stream")
	require.NoError(t, err)

	require.True(t, IsProfileReserved("my_stream"))
	require.False(t, IsProfileReserved("my"))

	_, err = CreateProfile("my")
	require.NoError(t, err)

	require.NoError(t, ResetProfileToken("my_stream"))

	removed, err := RemoveProfile("my_stream")
	require.NoError(t, err)
	require.True(t, removed)

	require.False(t, IsProfileReserved("my_stream"))
	require.True(t, IsProfileReserved("my"))
}
