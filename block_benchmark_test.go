package bc

import (
	"crypto/rand"
	"math"
	"testing"

	"github.com/bsv-blockchain/go-sdk/chainhash"
	"github.com/stretchr/testify/require"
)

// BenchmarkNewBUMPFromMerkleTreeAndIndex benchmarks the NewBUMPFromMerkleTreeAndIndex function
func BenchmarkNewBUMPFromMerkleTreeAndIndex(b *testing.B) {
	transactions := 100000
	// test how quickly we can calculate the BUMP Merkle Paths from a block of 100,000 random txids
	chainHashBlock := make([]*chainhash.Hash, 0)
	for i := 0; i < transactions; i++ {
		bytes := make([]byte, 32)
		_, _ = rand.Read(bytes)
		hash, err := chainhash.NewHash(bytes)
		if err != nil {
			b.Fatal(err)
		}
		chainHashBlock = append(chainHashBlock, hash)
	}
	merkles := BuildMerkleTreeStoreChainHash(chainHashBlock)

	b.ResetTimer()

	for idx := 0; idx < transactions; idx++ {
		bump, err := NewBUMPFromMerkleTreeAndIndex(850000, merkles, uint64(idx)) //nolint:gosec // G115: Safe conversion - idx is bounded by test
		require.NoError(b, err)
		totalHashes := 0
		for _, level := range bump.Path {
			totalHashes += len(level)
		}
		// number of levels plus the txid itself.
		l := int(math.Ceil(math.Log2(float64(transactions)))) + 1
		require.Equal(b, l, totalHashes)
	}
}

// BenchmarkExpandTargetFrom benchmarks the ExpandTargetFrom function
func BenchmarkExpandTargetFrom(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ExpandTargetFrom("182815ee")
	}
}
