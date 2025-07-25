package spv_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/bsv-blockchain/go-bc"
	"github.com/bsv-blockchain/go-bc/spv"
	"github.com/bsv-blockchain/go-bc/testing/data"
	"github.com/bsv-blockchain/go-bt/v2"
)

type mockBlockHeaderClient struct {
	blockHeaderFunc func(context.Context, string) (*bc.BlockHeader, error)
}

// BlockHeader is a mock implementation of the BlockHeader method for the
func (m *mockBlockHeaderClient) BlockHeader(ctx context.Context, blockHash string) (*bc.BlockHeader, error) {
	if m.blockHeaderFunc != nil {
		return m.blockHeaderFunc(ctx, blockHash)
	}

	return nil, errors.New("blockHeaderFunc in test is undefined")
}

// TestSPVEnvelope_VerifyPayment tests the VerifyPayment method of the SPV envelope.
func TestSPVEnvelope_VerifyPayment(t *testing.T) {
	t.Skip("this is failing due to bsv-blockchain/go-bt vs libsv/go-bt incompatibility")

	tests := map[string]struct {
		testFile string
		// setupOpts are passed to the NewVerifier func.
		setupOpts []spv.VerifyOpt
		// overrideOpts are passed to the VerifyPayment func to override the global settings.
		overrideOpts []spv.VerifyOpt
		exp          bool
		expErr       error
		expErrBinary error
	}{
		"valid ancestry passes": {
			exp:      true,
			testFile: "valid",
		},
		"ancestry without any proof fails": {
			exp:          false,
			testFile:     "invalid_missing_merkle_proof",
			expErr:       spv.ErrNoConfirmedTransaction,
			expErrBinary: spv.ErrProofOrInputMissing,
		},
		"ancestry without any proof passes if proof disabled": {
			exp:      true,
			testFile: "invalid_missing_merkle_proof",
			setupOpts: []spv.VerifyOpt{
				spv.NoVerifyProofs(),
			},
		},
		"ancestry without any proof passes if spv disabled": {
			exp:      true,
			testFile: "invalid_missing_merkle_proof",
			setupOpts: []spv.VerifyOpt{
				spv.NoVerifySPV(),
			},
		},
		"ancestry without any proof passes if spv overridden": {
			exp:      true,
			testFile: "invalid_missing_merkle_proof",
			overrideOpts: []spv.VerifyOpt{
				spv.NoVerifyProofs(),
			},
		},
		"valid ancestry with fee check supplied and valid fees passes": {
			exp:      true,
			testFile: "valid",
			overrideOpts: []spv.VerifyOpt{
				spv.VerifyFees(bt.NewFeeQuote()),
			},
		},
		"valid ancestry with fee check supplied and invalid fees fails": {
			exp:          false,
			testFile:     "valid",
			expErr:       spv.ErrFeePaidNotEnough,
			expErrBinary: spv.ErrFeePaidNotEnough,
			overrideOpts: []spv.VerifyOpt{
				spv.VerifyFees(bt.NewFeeQuote().AddQuote(bt.FeeTypeStandard, &bt.Fee{
					FeeType: bt.FeeTypeStandard,
					MiningFee: bt.FeeUnit{
						Satoshis: 10000000,
						Bytes:    1,
					},
				})),
			},
		},
		"wrong tx supplied as input in ancestry errs": {
			exp:          false,
			expErr:       spv.ErrNotAllInputsSupplied,
			expErrBinary: spv.ErrProofOrInputMissing,
			testFile:     "invalid_wrong_parent",
		},
		"tx with input missing from ancestry parents errors": {
			exp:          false,
			testFile:     "invalid_deep_parent_missing",
			expErr:       spv.ErrNotAllInputsSupplied,
			expErrBinary: spv.ErrProofOrInputMissing,
		},
		"valid ancestry with merkle proof supplied as hex passes": {
			exp:      true,
			testFile: "valid_merkle_proof_hex",
		},
		"ancestry with tx no inputs errs": {
			exp:          false,
			testFile:     "invalid_tx_missing_inputs",
			expErr:       spv.ErrNoTxInputsToVerify,
			expErrBinary: spv.ErrNoTxInputsToVerify,
		},
		"tx with input indexing out of bounds output errors": {
			exp:          false,
			testFile:     "invalid_tx_indexing_oob",
			expErr:       spv.ErrInputRefsOutOfBoundsOutput,
			expErrBinary: spv.ErrInputRefsOutOfBoundsOutput,
		},
		"tx with no inputs in multiple layer tx fails": {
			exp:          false,
			testFile:     "invalid_deep_tx_missing_inputs",
			expErr:       spv.ErrNoTxInputsToVerify,
			expErrBinary: spv.ErrNoTxInputsToVerify,
		},
		"ancestry with confirmed root errs": {
			exp:          false,
			testFile:     "invalid_confirmed_root",
			expErr:       spv.ErrTipTxConfirmed,
			expErrBinary: spv.ErrCannotCalculateFeePaid,
		},
		"nil initial payment errors": {
			exp:          false,
			expErr:       spv.ErrNilInitialPayment,
			expErrBinary: spv.ErrNilInitialPayment,
		},
		"ancestor, no parents, no spv, fee check should fail": {
			exp:          false,
			testFile:     "invalid_missing_parents",
			expErr:       spv.ErrCannotCalculateFeePaid,
			expErrBinary: spv.ErrCannotCalculateFeePaid,
			overrideOpts: []spv.VerifyOpt{
				spv.VerifyFees(bt.NewFeeQuote().AddQuote(bt.FeeTypeStandard, &bt.Fee{
					FeeType: bt.FeeTypeStandard,
					MiningFee: bt.FeeUnit{
						Satoshis: 0,
						Bytes:    10000,
					},
					RelayFee: bt.FeeUnit{},
				})),
				spv.NoVerifySPV(),
			},
		},
		"invalid merkle proof fails": {
			exp:          false,
			testFile:     "invalid_merkle_proof",
			expErr:       spv.ErrInvalidProof,
			expErrBinary: spv.ErrInvalidProof,
		},
		"wrong merkle proof supplied via hex with otherwise correct input errors": {
			exp:          false,
			testFile:     "invalid_wrong_merkle_proof_hex",
			expErr:       spv.ErrTxIDMismatch,
			expErrBinary: spv.ErrTxIDMismatch,
		},
		"wrong merkle proof supplied with otherwise correct input errors": {
			exp:          false,
			testFile:     "invalid_wrong_merkle_proof",
			expErr:       spv.ErrTxIDMismatch,
			expErrBinary: spv.ErrTxIDMismatch,
		},
		"valid multiple layer tx passes": {
			exp:      true,
			testFile: "valid_deep",
		},
		"single missing merkle proof in layered and branching tx errors": {
			exp:          false,
			testFile:     "invalid_deep_missing_merkle_proof",
			expErr:       spv.ErrNoConfirmedTransaction,
			expErrBinary: spv.ErrProofOrInputMissing,
		},
		"wrong merkle proof suppled with otherwise correct layered input errors": {
			exp:          false,
			testFile:     "invalid_deep_wrong_merkle_proof",
			expErr:       spv.ErrTxIDMismatch,
			expErrBinary: spv.ErrTxIDMismatch,
		},
		"invalid multiple layer tx false": {
			exp:          false,
			testFile:     "invalid_deep_merkle_proof_index",
			expErr:       spv.ErrInvalidProof,
			expErrBinary: spv.ErrInvalidProof,
		},
	}

	mch := &mockBlockHeaderClient{
		blockHeaderFunc: func(_ context.Context, hash string) (*bc.BlockHeader, error) {
			bb, err := data.BlockHeaderData.Load(hash)
			if err != nil {
				return nil, err
			}
			return bc.NewBlockHeaderFromStr(string(bb[:160]))
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testData := struct {
				Envelope    *spv.AncestryJSON `json:"data"`
				Description string            `json:"description"`
			}{}
			if test.testFile != "" {
				bb, err := data.SpvVerifyData.Load(test.testFile + ".json")
				require.NoError(t, err)
				require.NoError(t, json.NewDecoder(bytes.NewBuffer(bb)).Decode(&testData))
			}

			if test.testFile == "" {
				require.EqualError(t, errors.Cause(spv.ErrNilInitialPayment), test.expErr.Error())
				return
			}

			v, err := spv.NewPaymentVerifier(mch, test.setupOpts...)
			require.NoError(t, err, "expected no error when creating spv client")

			ancestryBytes, err := testData.Envelope.Bytes()
			require.NoError(t, err, "expected no error when creating binary from json")

			paymentBytes, err := hex.DecodeString(testData.Envelope.RawTx)
			require.NoError(t, err, "decoding hex rawtx failed")

			opts := append(test.setupOpts, test.overrideOpts...)
			paymentTx, err := bt.NewTxFromBytes(paymentBytes)
			if err != nil {
				require.NoError(t, err)
			}
			err = v.VerifyPayment(context.Background(), &spv.Payment{
				PaymentTx: paymentTx,
				Ancestry:  ancestryBytes,
			}, opts...)
			if test.expErrBinary != nil {
				require.Error(t, err)
				require.EqualError(t, errors.Cause(err), test.expErrBinary.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestVerifyAncestryBinary tests the VerifyPayment method of the SPV envelope with binary data.
func TestVerifyAncestryBinary(t *testing.T) {
	t.Skip("this is failing due to bsv-blockchain/go-bt vs libsv/go-bt incompatibility")

	tests := map[string]struct {
		testFile string
		// setupOpts are passed to the NewVerifier func.
		setupOpts []spv.VerifyOpt
		// overrideOpts are passed to the VerifyPayment func to override the global settings.
		overrideOpts []spv.VerifyOpt
		exp          bool
		expErr       error
		expErrBinary error
	}{
		"three txs all using eachothers outputs": {
			exp:      true,
			testFile: "valid_3_nested",
		},
		"1000 txs all using eachothers outputs": {
			exp:      true,
			testFile: "valid_1000_nested",
		},
	}

	mch := &mockBlockHeaderClient{
		blockHeaderFunc: func(_ context.Context, hash string) (*bc.BlockHeader, error) {
			bb, err := data.BlockHeaderData.Load(hash)
			if err != nil {
				return nil, err
			}
			return bc.NewBlockHeaderFromStr(string(bb[:160]))
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.testFile != "" {
				testDataJSON := struct {
					PaymentTx string `json:"paymentTx"`
					Ancestry  string `json:"ancestors"`
				}{}
				bb, err := data.SpvBinaryData.Load(test.testFile + ".json")
				require.NoError(t, err)
				require.NoError(t, json.NewDecoder(bytes.NewBuffer(bb)).Decode(&testDataJSON))

				v, err := spv.NewPaymentVerifier(mch, test.setupOpts...)
				require.NoError(t, err, "expected no error when creating spv client")

				paymentBytes, err := hex.DecodeString(testDataJSON.PaymentTx)
				require.NoError(t, err, "expected no error when creating binary from payemnt hex")

				ancestryBytes, err := hex.DecodeString(testDataJSON.Ancestry)
				require.NoError(t, err, "expected no error when creating binary from ancestry hex")

				opts := append(test.setupOpts, test.overrideOpts...)
				paymentTx, err := bt.NewTxFromBytes(paymentBytes)
				if err != nil {
					require.NoError(t, err)
				}
				err = v.VerifyPayment(context.Background(), &spv.Payment{
					PaymentTx: paymentTx,
					Ancestry:  ancestryBytes,
				}, opts...)
				if test.expErr != nil {
					require.Error(t, err)
					require.EqualError(t, errors.Cause(err), test.expErr.Error())
				} else {
					require.NoError(t, err)
				}
			}
		})
	}
}
