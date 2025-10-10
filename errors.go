package bc

import "errors"

const errInvalidBlockHeaderLengthMsg = "block header should be 80 bytes long"

// Package-level error definitions for consistent error handling
var (
	// Block errors
	ErrBlockEmpty               = errors.New("block cannot be empty")
	ErrInvalidBlockHeaderLength = errors.New(errInvalidBlockHeaderLengthMsg)

	// BUMP errors
	ErrInsufficientBUMPData = errors.New("BUMP bytes do not contain enough data to be valid")
	ErrInvalidLeafHeight    = errors.New("there are no leaves at height which makes this invalid")
	ErrTxidNotInBUMP        = errors.New("the BUMP does not contain the txid")
	ErrEmptyMerkleTree      = errors.New("merkle tree is empty")
	ErrNoHashAtIndex        = errors.New("we do not have a hash for this index at height")

	// Merkle proof errors
	ErrIndexOutOfRange    = errors.New("index out of range for proof")
	ErrInvalidTransaction = errors.New("invalid transaction")
)
