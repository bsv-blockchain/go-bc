package spv_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/bsv-blockchain/go-bc"
	"github.com/bsv-blockchain/go-bc/spv"
)

func TestPaymentVerifier_NewPaymentVerifier(t *testing.T) {
	tests := map[string]struct {
		bhc    bc.BlockHeaderChain
		expErr error
	}{
		"successful create": {
			bhc: &mockBlockHeaderClient{},
		},
		"undefined bhc errors": {
			expErr: errors.New("at least one blockchain header implementation should be returned"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := spv.NewPaymentVerifier(test.bhc)
			if test.expErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, test.expErr.Error())
			}
		})
	}
}

func TestMerkleProofVerifier_NewMerkleProofVerifier(t *testing.T) {
	tests := map[string]struct {
		bhc    bc.BlockHeaderChain
		expErr error
	}{
		"successful create": {
			bhc: &mockBlockHeaderClient{},
		},
		"undefined bhc errors": {
			expErr: errors.New("at least one blockchain header implementation should be returned"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := spv.NewMerkleProofVerifier(test.bhc)
			if test.expErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, test.expErr.Error())
			}
		})
	}
}
