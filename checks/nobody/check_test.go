package nobody

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafeas/voucher"
	vtesting "github.com/grafeas/voucher/testing"
)

func TestNobodyCheck(t *testing.T) {
	server := vtesting.NewTestDockerServer(t)

	auth := vtesting.NewAuth(server)

	nobodyCheck := new(check)
	nobodyCheck.SetAuth(auth)

	i := vtesting.NewTestReference(t)

	pass, err := nobodyCheck.Check(context.Background(), i)

	require.NoErrorf(t, err, "check failed with error: %s", err)
	assert.False(t, pass, "check passed when it should have failed")
}

func TestNobodyCheckWithNoAuth(t *testing.T) {
	i := vtesting.NewTestReference(t)

	nobodyCheck := new(check)

	// run check without setting up Auth.
	pass, err := nobodyCheck.Check(context.Background(), i)
	require.Equal(t, err, voucher.ErrNoAuth, "check should have failed due to lack of Auth, but didn't")
	assert.False(t, pass, "check passed when it should have failed due to no Auth")
}
