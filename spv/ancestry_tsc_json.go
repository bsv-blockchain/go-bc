package spv

import (
	"encoding/hex"

	"github.com/bsv-blockchain/go-bc"
	"github.com/bsv-blockchain/go-bt/v2"
)

const (
	h = "header"
)

// TSCAncestriesJSON spec at https://tsc.bitcoinassociation.net/standards/transaction-ancestors/ eventually.
type TSCAncestriesJSON []TSCAncestryJSON

// TSCAncestryJSON is one of the serial objects within the overall list of ancestors.
//
// This JSON structure follows the TSC definition even though the other JSON
// structure used in ancestry_json.go is more useful for verification.
type TSCAncestryJSON struct {
	RawTx         string             `json:"rawtx,omitempty"`
	Proof         *bc.MerkleProof    `json:"proof,omitempty"`
	MapiResponses []*bc.MapiCallback `json:"mapiResponses,omitempty"`
}

// NewAncestryJSONFromBytes is a way to create the JSON format for Ancestry from the binary format.
func NewAncestryJSONFromBytes(b []byte) (TSCAncestriesJSON, error) {
	ancestry, err := parseAncestry(b)
	if err != nil {
		return nil, err
	}
	ancestors := make([]TSCAncestryJSON, 0)
	for _, ancestor := range ancestry {
		rawTx := ancestor.Tx.String()
		a := TSCAncestryJSON{
			RawTx:         rawTx,
			MapiResponses: ancestor.MapiResponses,
		}
		if ancestor.Proof != nil {
			mpb, err := parseBinaryMerkleProof(ancestor.Proof)
			if err != nil {
				return nil, err
			}
			a.Proof = &bc.MerkleProof{
				Index:     mpb.index,
				TxOrID:    mpb.txOrID,
				Target:    mpb.target,
				Nodes:     mpb.nodes,
				ProofType: flagProofType(mpb.flags),
			}
		}
		ancestors = append(ancestors, a)
	}
	return ancestors, nil
}

// Bytes take an AncestryJSON and return the serialized bytes.
func (j TSCAncestriesJSON) Bytes() ([]byte, error) {
	binaryTxContext := make([]byte, 0)

	// Binary format version 1.
	binaryTxContext = append(binaryTxContext, 1)

	// follow with the list of ancestors, including their proof or mapi responses if present.
	for _, ancestor := range j {
		rawTx, err := hex.DecodeString(ancestor.RawTx)
		if err != nil {
			return nil, err
		}
		length := bt.VarInt(uint64(len(rawTx)))
		binaryTxContext = append(binaryTxContext, flagTx)
		binaryTxContext = append(binaryTxContext, length.Bytes()...)
		binaryTxContext = append(binaryTxContext, rawTx...)
		if ancestor.Proof != nil {
			rawProof, err := ancestor.Proof.Bytes()
			if err != nil {
				return nil, err
			}
			length = bt.VarInt(uint64(len(rawProof)))
			binaryTxContext = append(binaryTxContext, flagProof)
			binaryTxContext = append(binaryTxContext, length.Bytes()...)
			binaryTxContext = append(binaryTxContext, rawProof...)
		}
		if len(ancestor.MapiResponses) > 0 {
			binaryTxContext = append(binaryTxContext, flagMapi)
			numOfMapiResponses := bt.VarInt(uint64(len(ancestor.MapiResponses)))
			binaryTxContext = append(binaryTxContext, numOfMapiResponses.Bytes()...)
			for _, mapiResponse := range ancestor.MapiResponses {
				mapiR, err := mapiResponse.Bytes()
				if err != nil {
					return nil, err
				}
				dataLength := bt.VarInt(uint64(len(mapiR)))
				binaryTxContext = append(binaryTxContext, dataLength.Bytes()...)
				binaryTxContext = append(binaryTxContext, mapiR...)
			}
		}
	}

	return binaryTxContext, nil
}

func flagProofType(flags byte) string {
	switch flags & targetTypeFlags {
	// if bits 1 and 2 of flags are NOT set, target should contain a block hash (32 bytes).
	// if bit 2 of flags is set, the target should contain a merkle root (32 bytes).
	case 0, 4:
		return "blockhash"
	// if bit 1 of flags is set, the target should contain a block header (80 bytes).
	case 2:
		return h
	default:
		return ""
	}
}
