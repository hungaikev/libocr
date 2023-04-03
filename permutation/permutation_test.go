package permutation

import (
	"testing"
)

// TestPermutation checks if the permutation is randomized.
func TestPermutation(t *testing.T) {
	n := 10
	key := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	perm1 := Permutation(n, key)
	perm2 := Permutation(n, key)
	for i := 0; i < n; i++ {
		if perm1[i] != perm2[i] {
			return
		}
	}
	t.Errorf("expected different permutations, got %v  and %v", perm1, perm2)
}

// TestPermutationLength checks if the length of the permutation is equal to the input length.
func TestPermutationLength(t *testing.T) {
	n := 10
	key := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	perm := Permutation(n, key)
	if len(perm) != n {
		t.Errorf("expected permutation length to be %d, got %d", n, len(perm))
	}
}

// TestPermutationDuplicates checks if the permutation contains no duplicates.
func TestPermutationDuplicates(t *testing.T) {
	n := 10
	key := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	perm := Permutation(n, key)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if perm[i] == perm[j] {
				t.Errorf("duplicate element found in permutation: %v", perm)
			}
		}
	}
}

// TestPermutationRange checks if the permutation only contains values in the range [0, n-1].
func TestPermutationRange(t *testing.T) {
	n := 10
	key := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	perm := Permutation(n, key)
	for i := 0; i < n; i++ {
		if perm[i] < 0 || perm[i] >= n {
			t.Errorf("out of range element found in permutation: %v", perm)
		}
	}
}

// TestCryptoRandSource checks if the Int63 method of the cryptoRandSource struct returns a random number within the range [0, 9223372036854775807].
func TestCryptoRandSource(t *testing.T) {
	key := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	crs := newCryptoRandSource(key)
	var generatedNums [10]int64
	for i := 0; i < 10; i++ {
		num := crs.Int63()
		if num < 0 || num > 9223372036854775807 {
			t.Errorf("generated number out of range: %v", num)
		}
		generatedNums[i] = num
	}
	// Check that the generated numbers are not all the same.
	if allSame(generatedNums) {
		t.Errorf("expected different generated numbers, got %v", generatedNums)
	}
}

// TestCryptoRandSourceConsistency checks if the same cryptoRandSource with the same key and seed returns the same sequence of random numbers.
func TestCryptoRandSourceConsistency(t *testing.T) {
	key := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	crs1 := newCryptoRandSource(key)
	crs2 := newCryptoRandSource(key)
	var generatedNums1 [10]int64
	var generatedNums2 [10]int64
	for i := 0; i < 10; i++ {
		num1 := crs1.Int63()
		num2 := crs2.Int63()
		if num1 != num2 {
			t.Errorf("expected same generated numbers, got %v and %v", num1, num2)
		}
		generatedNums1[i] = num1
		generatedNums2[i] = num2
	}
	// Check that the generated numbers are the same for both generators.
	for i := 0; i < 10; i++ {
		if generatedNums1[i] != generatedNums2[i] {
			t.Errorf("expected same generated numbers, got %v and %v", generatedNums1, generatedNums2)
		}
	}
}

// allSame checks if all the elements in a slice are the same.
func allSame(nums [10]int64) bool {
	for i := 1; i < len(nums); i++ {
		if nums[i] != nums[0] {
			return false
		}
	}
	return true
}
