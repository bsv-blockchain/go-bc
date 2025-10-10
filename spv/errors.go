package spv

import "github.com/pkg/errors"

var (
	// ErrNoTxInputs returns if an ancestry is attempted to be created from a transaction that has no inputs.
	ErrNoTxInputs = errors.New("provided tx has no inputs to build ancestry from")

	// ErrTxNotFound is returned when a transaction is not found in the store.
	ErrTxNotFound = errors.New("tx not found")

	// ErrPaymentNotVerified returns if a transaction in the tree provided was missed during verification.
	ErrPaymentNotVerified = errors.New("a tx was missed during validation")

	// ErrTipTxConfirmed returns if the tip transaction is already confirmed.
	ErrTipTxConfirmed = errors.New("tip transaction must be unconfirmed")

	// ErrNoConfirmedTransaction returns if a path from tip to beginning/anchor contains no confirmed transaction.
	ErrNoConfirmedTransaction = errors.New("not confirmed/anchored tx(s) provided")

	// ErrTxIDMismatch returns if they key value pair of a transactions input has a mismatch in txID.
	ErrTxIDMismatch = errors.New("input and proof ID mismatch")

	// ErrNotAllInputsSupplied returns if an unconfirmed transaction in ancestry contains inputs which are not
	// present in the parent ancestor.
	ErrNotAllInputsSupplied = errors.New("a tx input missing in parent ancestor")

	// ErrNoTxInputsToVerify returns if a transaction has no inputs.
	ErrNoTxInputsToVerify = errors.New("a tx has no inputs to verify")

	// ErrNilInitialPayment returns if a transaction has no inputs.
	ErrNilInitialPayment = errors.New("initial payment cannot be nil")

	// ErrInputRefsOutOfBoundsOutput returns if a transaction has no inputs.
	ErrInputRefsOutOfBoundsOutput = errors.New("tx input index into output is out of bounds")

	// ErrNoFeeQuoteSupplied is returned when VerifyFees is enabled but no bt.FeeQuote has been supplied.
	ErrNoFeeQuoteSupplied = errors.New("no bt.FeeQuote supplied for fee validation, supply the bt.FeeQuote using VerifyFees opt")

	// ErrFeePaidNotEnough returned when not enough fees have been paid.
	ErrFeePaidNotEnough = errors.New("not enough fees paid")

	// ErrCannotCalculateFeePaid returned when fee check is enabled but the tx has no parents.
	ErrCannotCalculateFeePaid = errors.New("no parents supplied in ancestry which means we cannot valdiate " +
		"fees, either ensure parents are supplied or remove fee check")

	// ErrInvalidProof is returned if the merkle proof validation fails.
	ErrInvalidProof = errors.New("invalid merkle proof, payment invalid")

	// ErrMissingOutput is returned when checking fees if an output in a parent tx is missing.
	ErrMissingOutput = errors.New("expected output used in payment tx missing")

	// ErrProofOrInputMissing returns if a path from tip to beginning/anchor is broken.
	ErrProofOrInputMissing = errors.New("break in the ancestry missing either a parent transaction or a proof")

	// ErrTriedToParseZeroBytes returns when we attempt to parse a slice of bytes of zero length which should be a mapi response.
	ErrTriedToParseZeroBytes = errors.New("there are no mapi response bytes to parse")

	// ErrUnsupporredVersion returns if another version of the binary format is being used - since we cannot guarantee we know how to parse it.
	ErrUnsupporredVersion = errors.New("we only support version 1 of the Ancestor Binary format")

	// ErrInvalidMerkleFlags returns if a merkle proof being verified uses something other than the one currently supported.
	ErrInvalidMerkleFlags = errors.New("invalid flags used in merkle proof")

	// ErrMissingTxidInProof returns if there's a missing txid in the proof.
	ErrMissingTxidInProof = errors.New("missing txid in proof")

	// ErrMissingRootInProof returns if there's a missing root in the proof.
	ErrMissingRootInProof = errors.New("missing root in proof")

	// ErrInvalidNodes returns if there is a * on the left hand side within the node array.
	ErrInvalidNodes = errors.New("invalid nodes")

	// ErrTxGetterUndefined is returned when txGetterFunc in test is undefined.
	ErrTxGetterUndefined = errors.New("txGetterFunc in test is undefined")

	// ErrMpGetterUndefined is returned when mpGetterFunc in test is undefined.
	ErrMpGetterUndefined = errors.New("mpGetterFunc in test is undefined")

	// ErrInvalidTarget is returned when target field is invalid.
	ErrInvalidTarget = errors.New("invalid target field")

	// ErrInvalidTargetType is returned when TargetType or target field is invalid.
	ErrInvalidTargetType = errors.New("invalid TargetType or target field")

	// ErrOnlyMerkleBranchSupported is returned when a non-merkle branch proof format is used.
	ErrOnlyMerkleBranchSupported = errors.New("only merkle branch supported in this version")

	// ErrOnlySingleProofSupported is returned when multiple proofs are provided.
	ErrOnlySingleProofSupported = errors.New("only single proof supported in this version")

	// ErrTxidMissing is returned when txid is missing in proof.
	ErrTxidMissing = errors.New("txid missing")

	// ErrMerkleRootMissing is returned when merkle root is missing in proof.
	ErrMerkleRootMissing = errors.New("merkleRoot missing")

	// ErrExpectingTxidButGotTx is returned when txid is expected but tx is provided.
	ErrExpectingTxidButGotTx = errors.New("expecting txid but got tx")

	// ErrExpectingTxButGotTxid is returned when tx is expected but txid is provided.
	ErrExpectingTxButGotTxid = errors.New("expecting tx but got txid")

	// ErrInvalidTxOrIDLength is returned when txOrId length is invalid.
	ErrInvalidTxOrIDLength = errors.New("invalid txOrId length - must be at least 64 chars (32 bytes)")

	// ErrInvalidTxLength is returned when tx length is invalid.
	ErrInvalidTxLength = errors.New("invalid tx length (should be greater than 32 bytes)")

	// ErrInvalidNodeType is returned when node type value is invalid.
	ErrInvalidNodeType = errors.New("invalid value in node type")
)
