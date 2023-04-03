package permutation

import (
	"crypto/rand"
	"fmt"
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

func factorial(n int) int {
	if n == 0 {
		return 1
	}
	fact := 1
	for i := 1; i <= n; i++ {
		fact *= i
	}
	return fact
}

func TestSecureShuffling(t *testing.T) {
	n := 5
	iterations := 10000
	numPermutations := factorial(n)

	// Initialize a frequency map to count the occurrences of each permutation
	frequencyMap := make(map[string]int)

	for i := 0; i < iterations; i++ {
		key := generateRandomKey()
		result := Permutation(n, key)

		// Convert the result to a string for use as a map key
		keyString := fmt.Sprint(result)
		frequencyMap[keyString]++
	}

	// Perform a Chi-squared test
	// The expected frequency for each permutation is iterations / n!
	expectedFrequency := float64(iterations) / float64(numPermutations)
	var chiSquared float64
	for _, count := range frequencyMap {
		observed := float64(count)
		chiSquared += ((observed - expectedFrequency) * (observed - expectedFrequency)) / expectedFrequency
	}

	// Degrees of freedom: (n! - 1)
	// Using a significance level of 0.05 and 119 degrees of freedom,
	// the critical value for the Chi-squared test is approximately 146.6.
	// We expect the Chi-squared value to be lower than the critical value.
	criticalValue := 146.6
	assert.Less(t, chiSquared, criticalValue)
}
