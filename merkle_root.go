package bc

import (
	"encoding/hex"
	"fmt"
	"math"

	"github.com/bsv-blockchain/go-bt/v2"
	"github.com/bsv-blockchain/go-sdk/chainhash"
	crypto "github.com/bsv-blockchain/go-sdk/primitives/hash"
)

// TxsToTxIDs takes an array of transactions
// and returns instead an array of their
// corresponding transaction IDs.
func TxsToTxIDs(txs []string) ([]string, error) {
	txids := make([]string, 0, len(txs))

	for i, tx := range txs {
		t, err := bt.NewTxFromString(tx)
		if err != nil {
			return nil, fmt.Errorf("invalid transaction at index: %q", i)
		}
		txids = append(txids, t.TxID())
	}
	return txids, nil
}

// BuildMerkleRootFromCoinbase builds the merkle root of the block from the coinbase transaction hash (txid)
// and the merkle branches needed to work up the merkle tree and returns the merkle root byte array.
func BuildMerkleRootFromCoinbase(coinbaseHash []byte, merkleBranches []string) []byte {
	acc := coinbaseHash

	for i := 0; i < len(merkleBranches); i++ {
		branch, _ := hex.DecodeString(merkleBranches[i])
		concat := append(acc, branch...)
		hash := crypto.Sha256d(concat)
		acc = hash[:]
	}
	return acc
}

// BuildMerkleRoot builds the Merkle Root
// from a list of transactions.
func BuildMerkleRoot(txids []string) (string, error) {
	merkles, err := BuildMerkleTreeStore(txids)
	if err != nil {
		return "", err
	}
	return merkles[len(merkles)-1], nil
}

// BuildMerkleTreeStore creates a merkle tree from a slice of transaction IDs,
// stores it using a linear array, and returns a slice of the backing array.  A
// linear array was chosen as opposed to an actual tree structure since it uses
// about half as much memory.  The following describes a merkle tree and how it
// is stored in a linear array.
//
// A merkle tree is a tree in which every non-leaf node is the hash of its
// children nodes.  A diagram depicting how this works for bitcoin transactions
// where h(x) is a double sha256 follows:
//
//	         root = h1234 = h(h12 + h34)
//	        /                           \
//	  h12 = h(h1 + h2)            h34 = h(h3 + h4)
//	   /            \              /            \
//	h1 = h(tx1)  h2 = h(tx2)    h3 = h(tx3)  h4 = h(tx4)
//
// The above stored as a linear array is as follows:
//
//	[h1 h2 h3 h4 h12 h34 root]
//
// As the above shows, the merkle root is always the last element in the array.
//
// The number of inputs is not always a power of two which results in a
// balanced tree structure as above.  In that case, parent nodes with no
// children are also zero and parent nodes with only a single left node
// are calculated by concatenating the left node with itself before hashing.
// Since this function uses nodes that are pointers to the hashes, empty nodes
// will be nil.
//
// The additional bool parameter indicates if we are generating the merkle tree
// using witness transaction id's rather than regular transaction id's. This
// also presents an additional case wherein the wtxid of the coinbase transaction
// is the zeroHash.
//
// based off of bsvd:
// https://github.com/bitcoinsv/bsvd/blob/4c29707f717300d3eb92352081c3b0fec556881b/blockchain/merkle.go#L74
func BuildMerkleTreeStore(txids []string) ([]string, error) {
	// // Calculate how many entries are re?n array of that size.
	nextPoT := nextPowerOfTwo(len(txids))
	arraySize := nextPoT*2 - 1
	merkles := make([]string, arraySize)

	// Create the base transaction hashes and populate the array with them.
	copy(merkles, txids)

	// Start the array offset after the last transaction and adjusted to the
	// next power of two.
	offset := nextPoT
	for i := 0; i < arraySize-1; i += 2 {
		switch {
		// When there is no left child node, the parent is nil ("") too.
		case merkles[i] == "":
			merkles[offset] = ""

			// When there is no right child, the parent is generated by
			// hashing the concatenation of the left child with itself.
		case merkles[i+1] == "":
			newHash, err := MerkleTreeParentStr(merkles[i], merkles[i])
			if err != nil {
				return nil, err
			}
			merkles[offset] = newHash

			// The normal case sets the parent node to the double sha256
			// of the concatenation of the left and right children.
		default:
			// newHash := HashMerkleBranches(merkles[i], merkles[i+1])
			newHash, err := MerkleTreeParentStr(merkles[i], merkles[i+1])
			if err != nil {
				return nil, err
			}
			merkles[offset] = newHash
		}
		offset++
	}

	return merkles, nil
}

// BuildMerkleTreeStoreChainHash has the same functionality as BuildMerkleTreeStore but uses chainhash as a type to avoid string conversions.
func BuildMerkleTreeStoreChainHash(txids []*chainhash.Hash) []*chainhash.Hash {
	// // Calculate how many entries are re?n array of that size.
	nextPoT := nextPowerOfTwo(len(txids))
	arraySize := nextPoT*2 - 1
	merkles := make([]*chainhash.Hash, arraySize)

	// Create the base transaction hashes and populate the array with them.
	copy(merkles, txids)

	// Start the array offset after the last transaction and adjusted to the
	// next power of two.
	offset := nextPoT
	for i := 0; i < arraySize-1; i += 2 {
		switch {
		// When there is no left child node, the parent is nil ("") too.
		case merkles[i].IsEqual(nil):
			merkles[offset] = nil

			// When there is no right child, the parent is generated by
			// hashing the concatenation of the left child with itself.
		case merkles[i+1].IsEqual(nil):
			merkles[offset] = MerkleTreeParentBytes(merkles[i], merkles[i])

			// The normal case sets the parent node to the double sha256
			// of the concatenation of the left and right children.
		default:
			merkles[offset] = MerkleTreeParentBytes(merkles[i], merkles[i+1])
		}
		offset++
	}

	return merkles
}

// nextPowerOfTwo returns the next highest power of two from a given number if
// it is not already a power of two.  This is a helper function used during the
// calculation of a merkle tree.
func nextPowerOfTwo(n int) int {
	// Return the number if it's already a power of 2.
	if n&(n-1) == 0 {
		return n
	}

	// Figure out and return the next power of two.
	exponent := uint(math.Log2(float64(n))) + 1
	return 1 << exponent // 2^exponent
}
