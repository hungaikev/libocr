package permutation

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateRandomKey() [16]byte {
	var key [16]byte
	_, err := rand.Read(key[:])
	if err != nil {
		panic(err)
	}
	return key
}

func TestValidKey(t *testing.T) {
	n := 10
	key := generateRandomKey()

	result := Permutation(n, key)

	assert.Len(t, result, n)
	assert.ElementsMatch(t, result, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
}

func TestEmptyPermutation(t *testing.T) {
	n := 0
	key := generateRandomKey()

	result := Permutation(n, key)

	assert.Empty(t, result)
}

func TestSingleElementPermutation(t *testing.T) {
	n := 1
	key := generateRandomKey()

	result := Permutation(n, key)

	assert.Equal(t, []int{0}, result)
}

func TestConsistentPermutation(t *testing.T) {
	n := 5
	key := generateRandomKey()

	result1 := Permutation(n, key)
	result2 := Permutation(n, key)

	require.Len(t, result1, n)
	assert.Equal(t, result1, result2)
}

func TestDifferentKeys(t *testing.T) {
	n := 5
	key1 := generateRandomKey()
	key2 := generateRandomKey()

	result1 := Permutation(n, key1)
	result2 := Permutation(n, key2)

	require.Len(t, result1, n)
	require.Len(t, result2, n)
	assert.NotEqual(t, result1, result2)
}

func TestDifferentPermutationSizes(t *testing.T) {
	n1 := 5
	n2 := 7
	key := generateRandomKey()

	result1 := Permutation(n1, key)
	result2 := Permutation(n2, key)

	assert.Len(t, result1, n1)
	assert.Len(t, result2, n2)
}
