package mkt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHappyPath(t *testing.T) {
	m := NewMerkleTree([]string{"a", "b", "c", "d"})
	require.NotEmpty(t, m)

	m.PrintAllProofs()
	m.PrintTree()

	proof, err := m.GetProof("a")
	require.NoError(t, err)

	t.Run("VerifyProof", func(t *testing.T) {
		hash := GetProofHash("a", proof)
		require.NotEmpty(t, hash)

		isValid := VerifyProof("a", hash, proof)
		require.True(t, isValid)
	})
}
