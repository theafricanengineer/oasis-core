package api

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	curve25519 "github.com/oasislabs/ed25519/extra/x25519"
)

func Test_GenerateKeypair(t *testing.T) {
	pub, priv, err := GenerateKeyPair(rand.Reader)
	require.NoError(t, err, "GenerateKeyPair")

	var pubTmp [32]byte
	curve25519.ScalarBaseMult(&pubTmp, priv)
	require.EqualValues(t, pub, &pubTmp, "scalarBaseMult(priv) ?= pub")
}
