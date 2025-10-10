package spv_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/bsv-blockchain/go-bt/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bsv-blockchain/go-bc"
	"github.com/bsv-blockchain/go-bc/spv"
	"github.com/bsv-blockchain/go-bc/testing/data"
)

var (
	// Test errors for create_ancestry tests
	errBigBad                       = errors.New("big bad error")
	errBiggerBadder                 = errors.New("bigger badder error")
	errCloseButNoCigar              = errors.New("close but no cigar")
	errOhNo                         = errors.New("oh no")
	errCouldNotFindTx36327872       = errors.New("could not find tx 36327872f3d4fb62f8dc0e0a746715df21b59f98f8a18848d703193fa61e55cb: tx not found")
	errFailedToGetTx6915ee43        = errors.New("failed to get tx 6915ee43a9d12ffd63c064a83956a8d8dd296270054241bd379fb4ac1eda1347: big bad error")
	errFailedToGetMerkleProofA7a97c = errors.New("failed to get merkle proof for tx a7a97cb7650b8ff1ee294f86653664adfbaef112d8813784a28b1c735b550d25: bigger badder error")
	errCouldNotFindTxA8e76021       = errors.New("could not find tx a8e760210a8e3646ded829860745c60a2ade443d6998465dd0da5ae5f37b3b8e: tx not found")
	errFailedToGetTx8fff14aa        = errors.New("failed to get tx 8fff14aabcad6ebd1da2ce7737751e1e1e4c9c0d4c5b8abfef7e7914556d7965: close but no cigar")
	errFailedToGetMerkleProof3fa9a1 = errors.New("failed to get merkle proof for tx 3fa9a1a20c1f4dbd2c3dd5866749621ebe15c7290c13080ba0d45ab9c649cc15: oh no")
)

const (
	txNotDefinedForTestFmt    = "txid %s not defined for test: %w"
	proofNotDefinedForTestFmt = "merkle proof for tx %s not defined in test: %w"
)

type mockTxMerkleGetter struct {
	txStoreFunc func(context.Context, string) (*bt.Tx, error)
	mpStoreFunc func(context.Context, string) (*bc.MerkleProof, error)
}

func (m *mockTxMerkleGetter) Tx(ctx context.Context, txID string) (*bt.Tx, error) {
	if m.txStoreFunc == nil {
		return nil, spv.ErrTxGetterUndefined
	}

	return m.txStoreFunc(ctx, txID)
}

func (m *mockTxMerkleGetter) MerkleProof(ctx context.Context, txID string) (*bc.MerkleProof, error) {
	if m.mpStoreFunc == nil {
		return nil, spv.ErrMpGetterUndefined
	}

	return m.mpStoreFunc(ctx, txID)
}

type txOverride struct {
	tx  *bt.Tx
	err error
}

type proofOverride struct {
	proof *bc.MerkleProof
	err   error
}

var (
	fixturesOnce  sync.Once
	fixturesErr   error
	fixtureTxs    map[string]string
	fixtureProofs map[string]*bc.MerkleProof
)

func newTxFunc(t *testing.T, overrides map[string]*txOverride) func(context.Context, string) (*bt.Tx, error) {
	t.Helper()
	ensureFixtureData(t)

	return func(_ context.Context, txID string) (*bt.Tx, error) {
		if override, ok := overrides[txID]; ok {
			if override == nil {
				return nil, nil
			}
			return override.tx, override.err
		}

		raw, ok := fixtureTxs[txID]
		if !ok {
			return nil, fmt.Errorf(txNotDefinedForTestFmt, txID, spv.ErrNotAllInputsSupplied)
		}

		tx, err := bt.NewTxFromString(raw)
		if err != nil {
			return nil, err
		}

		return tx, nil
	}
}

func newMerkleProofFunc(t *testing.T, overrides map[string]*proofOverride) func(context.Context, string) (*bc.MerkleProof, error) {
	t.Helper()
	ensureFixtureData(t)

	return func(_ context.Context, txID string) (*bc.MerkleProof, error) {
		if override, ok := overrides[txID]; ok {
			if override == nil {
				return nil, nil
			}
			return override.proof, override.err
		}

		proof, ok := fixtureProofs[txID]
		if !ok {
			return nil, fmt.Errorf(proofNotDefinedForTestFmt, txID, spv.ErrNotAllInputsSupplied)
		}

		return proof, nil
	}
}

func ensureFixtureData(t *testing.T) {
	t.Helper()

	fixturesOnce.Do(func() {
		fixtureTxs, fixtureProofs, fixturesErr = loadFixtureData()
	})

	require.NoError(t, fixturesErr)
}

func loadFixtureData() (map[string]string, map[string]*bc.MerkleProof, error) {
	txs := make(map[string]string)
	proofs := make(map[string]*bc.MerkleProof)

	for _, file := range []string{"spvancestor_1.json", "spvancestor_2.json"} {
		bb, err := data.SpvCreateData.Load(file)
		if err != nil {
			return nil, nil, err
		}

		var env spv.AncestryJSON
		if err := json.NewDecoder(bytes.NewReader(bb)).Decode(&env); err != nil {
			return nil, nil, err
		}

		if err := collectFixtureData(&env, txs, proofs); err != nil {
			return nil, nil, err
		}
	}

	return txs, proofs, nil
}

func collectFixtureData(node *spv.AncestryJSON, txs map[string]string, proofs map[string]*bc.MerkleProof) error {
	if node == nil {
		return nil
	}

	if node.TxID != "" {
		if node.RawTx != "" {
			if _, exists := txs[node.TxID]; !exists {
				txs[node.TxID] = node.RawTx
			}
		}

		if existing, exists := proofs[node.TxID]; !exists || (existing == nil && node.Proof != nil) {
			proofs[node.TxID] = node.Proof
		}
	}

	for id, parent := range node.Parents {
		if parent == nil {
			continue
		}

		if parent.TxID == "" {
			parent.TxID = id
		}

		if err := collectFixtureData(parent, txs, proofs); err != nil {
			return err
		}
	}

	return nil
}

func TestSPVEnvelopeCreateEnvelope(t *testing.T) {
	tests := []struct {
		name           string
		tx             string
		txOverrides    map[string]*txOverride
		proofOverrides map[string]*proofOverride
		expFile        string
		expErr         error
	}{
		{
			name:    "valid ancestry created",
			tx:      "0200000005c931fb0af1eedc1fcb1c00171c80fcdef48564d3a8582dd530a9ba258497b1ea000000006a473044022011dd49f90eb34195e61712cef41d16ce9496807124b1e1c6205cb06ccfdbef0002203b2553032d166724d89a42a12b997cb1c4a78c4d8bbc72f94ab9ef4c3db36d38412103f6e8ebb2836f89aedbe712fa91eda827df597a01fe1a19fa1658bc1d28d1ad15feffffffc931fb0af1eedc1fcb1c00171c80fcdef48564d3a8582dd530a9ba258497b1ea010000006b483045022100f8a92d8e09a239d863b57c44fd2915558afbe05074c992b72639288f15a5a928022047399d49fcc7354140ecec229fe682383a9e9c59e43da368d35d211a06a17f2641210363d67968518ee1c0485b9b95544c0a9ec8c280b649b72c74fe86c299cd055a3bfeffffff250d555b731c8ba2843781d812f1aefbad643665864f29eef18f0b65b77ca9a7010000006b483045022100ece110a2ae06c67f3d4b25ad4ec2d7acf85fa1618dabef9d7eed5b4e5c50bbea02207716c2a9dd8d1ce64a56a3377eb35892db42dbba62058ebeb1b14ca33a8b98c241210241f2c990d7e0fe5c1c5e4508883b76b9786fcc67ccee1b8724eefa89a8f32981feffffff4713da1eacb49f37bd414205706229ddd8a85639a864c063fd2fd1a943ee1569010000006b483045022100831aac063f1b32f9e5645c5442a6b15cb620c6c49930f0194fa6acf6c864f48202203a8cc93d8bcd6f7d86cac0c7c99485cc1837fc87790f13f5bbf2f3b901e86be341210376c2519a09f7cfbcbe000f823c1e957cadace64296e907ae1ea36536313f0706feffffffcb551ea63f1903d74888a1f8989fb521df1567740a0edcf862fbd4f372783236010000006b483045022100f0fa4ab61952f5497d4dfe6b7a4e5e9b3305a5b94b0d7a6d1536927dfea2aa4202205eaa36045812bf5561c9e3a38fc92e6d89908637a5c0ec250a843f36355418c2412103f6e8ebb2836f89aedbe712fa91eda827df597a01fe1a19fa1658bc1d28d1ad15feffffff02bccdf505000000001976a914349256bff9dbe79285454d0e55d1a3163bd6dff888ac0084d717000000001976a914e2e4d329a79401e0a713210a4c615abdc540eda888ac68000000",
			expFile: "spvancestor_2",
		},
		{
			name: "creating an ancestry with tx that doesn't exist errors",
			tx:   "0200000005c931fb0af1eedc1fcb1c00171c80fcdef48564d3a8582dd530a9ba258497b1ea000000006a473044022011dd49f90eb34195e61712cef41d16ce9496807124b1e1c6205cb06ccfdbef0002203b2553032d166724d89a42a12b997cb1c4a78c4d8bbc72f94ab9ef4c3db36d38412103f6e8ebb2836f89aedbe712fa91eda827df597a01fe1a19fa1658bc1d28d1ad15feffffffc931fb0af1eedc1fcb1c00171c80fcdef48564d3a8582dd530a9ba258497b1ea010000006b483045022100f8a92d8e09a239d863b57c44fd2915558afbe05074c992b72639288f15a5a928022047399d49fcc7354140ecec229fe682383a9e9c59e43da368d35d211a06a17f2641210363d67968518ee1c0485b9b95544c0a9ec8c280b649b72c74fe86c299cd055a3bfeffffff250d555b731c8ba2843781d812f1aefbad643665864f29eef18f0b65b77ca9a7010000006b483045022100ece110a2ae06c67f3d4b25ad4ec2d7acf85fa1618dabef9d7eed5b4e5c50bbea02207716c2a9dd8d1ce64a56a3377eb35892db42dbba62058ebeb1b14ca33a8b98c241210241f2c990d7e0fe5c1c5e4508883b76b9786fcc67ccee1b8724eefa89a8f32981feffffff4713da1eacb49f37bd414205706229ddd8a85639a864c063fd2fd1a943ee1569010000006b483045022100831aac063f1b32f9e5645c5442a6b15cb620c6c49930f0194fa6acf6c864f48202203a8cc93d8bcd6f7d86cac0c7c99485cc1837fc87790f13f5bbf2f3b901e86be341210376c2519a09f7cfbcbe000f823c1e957cadace64296e907ae1ea36536313f0706feffffffcb551ea63f1903d74888a1f8989fb521df1567740a0edcf862fbd4f372783236010000006b483045022100f0fa4ab61952f5497d4dfe6b7a4e5e9b3305a5b94b0d7a6d1536927dfea2aa4202205eaa36045812bf5561c9e3a38fc92e6d89908637a5c0ec250a843f36355418c2412103f6e8ebb2836f89aedbe712fa91eda827df597a01fe1a19fa1658bc1d28d1ad15feffffff02bccdf505000000001976a914349256bff9dbe79285454d0e55d1a3163bd6dff888ac0084d717000000001976a914e2e4d329a79401e0a713210a4c615abdc540eda888ac68000000",
			txOverrides: map[string]*txOverride{
				"36327872f3d4fb62f8dc0e0a746715df21b59f98f8a18848d703193fa61e55cb": nil,
			},
			expErr: errCouldNotFindTx36327872,
		},
		{
			name: "error when getting tx is handled",
			tx:   "0200000005c931fb0af1eedc1fcb1c00171c80fcdef48564d3a8582dd530a9ba258497b1ea000000006a473044022011dd49f90eb34195e61712cef41d16ce9496807124b1e1c6205cb06ccfdbef0002203b2553032d166724d89a42a12b997cb1c4a78c4d8bbc72f94ab9ef4c3db36d38412103f6e8ebb2836f89aedbe712fa91eda827df597a01fe1a19fa1658bc1d28d1ad15feffffffc931fb0af1eedc1fcb1c00171c80fcdef48564d3a8582dd530a9ba258497b1ea010000006b483045022100f8a92d8e09a239d863b57c44fd2915558afbe05074c992b72639288f15a5a928022047399d49fcc7354140ecec229fe682383a9e9c59e43da368d35d211a06a17f2641210363d67968518ee1c0485b9b95544c0a9ec8c280b649b72c74fe86c299cd055a3bfeffffff250d555b731c8ba2843781d812f1aefbad643665864f29eef18f0b65b77ca9a7010000006b483045022100ece110a2ae06c67f3d4b25ad4ec2d7acf85fa1618dabef9d7eed5b4e5c50bbea02207716c2a9dd8d1ce64a56a3377eb35892db42dbba62058ebeb1b14ca33a8b98c241210241f2c990d7e0fe5c1c5e4508883b76b9786fcc67ccee1b8724eefa89a8f32981feffffff4713da1eacb49f37bd414205706229ddd8a85639a864c063fd2fd1a943ee1569010000006b483045022100831aac063f1b32f9e5645c5442a6b15cb620c6c49930f0194fa6acf6c864f48202203a8cc93d8bcd6f7d86cac0c7c99485cc1837fc87790f13f5bbf2f3b901e86be341210376c2519a09f7cfbcbe000f823c1e957cadace64296e907ae1ea36536313f0706feffffffcb551ea63f1903d74888a1f8989fb521df1567740a0edcf862fbd4f372783236010000006b483045022100f0fa4ab61952f5497d4dfe6b7a4e5e9b3305a5b94b0d7a6d1536927dfea2aa4202205eaa36045812bf5561c9e3a38fc92e6d89908637a5c0ec250a843f36355418c2412103f6e8ebb2836f89aedbe712fa91eda827df597a01fe1a19fa1658bc1d28d1ad15feffffff02bccdf505000000001976a914349256bff9dbe79285454d0e55d1a3163bd6dff888ac0084d717000000001976a914e2e4d329a79401e0a713210a4c615abdc540eda888ac68000000",
			txOverrides: map[string]*txOverride{
				"6915ee43a9d12ffd63c064a83956a8d8dd296270054241bd379fb4ac1eda1347": {
					err: errBigBad,
				},
			},
			expErr: errFailedToGetTx6915ee43,
		},
		{
			name: "error when getting merkle proof is handled",
			tx:   "0200000005c931fb0af1eedc1fcb1c00171c80fcdef48564d3a8582dd530a9ba258497b1ea000000006a473044022011dd49f90eb34195e61712cef41d16ce9496807124b1e1c6205cb06ccfdbef0002203b2553032d166724d89a42a12b997cb1c4a78c4d8bbc72f94ab9ef4c3db36d38412103f6e8ebb2836f89aedbe712fa91eda827df597a01fe1a19fa1658bc1d28d1ad15feffffffc931fb0af1eedc1fcb1c00171c80fcdef48564d3a8582dd530a9ba258497b1ea010000006b483045022100f8a92d8e09a239d863b57c44fd2915558afbe05074c992b72639288f15a5a928022047399d49fcc7354140ecec229fe682383a9e9c59e43da368d35d211a06a17f2641210363d67968518ee1c0485b9b95544c0a9ec8c280b649b72c74fe86c299cd055a3bfeffffff250d555b731c8ba2843781d812f1aefbad643665864f29eef18f0b65b77ca9a7010000006b483045022100ece110a2ae06c67f3d4b25ad4ec2d7acf85fa1618dabef9d7eed5b4e5c50bbea02207716c2a9dd8d1ce64a56a3377eb35892db42dbba62058ebeb1b14ca33a8b98c241210241f2c990d7e0fe5c1c5e4508883b76b9786fcc67ccee1b8724eefa89a8f32981feffffff4713da1eacb49f37bd414205706229ddd8a85639a864c063fd2fd1a943ee1569010000006b483045022100831aac063f1b32f9e5645c5442a6b15cb620c6c49930f0194fa6acf6c864f48202203a8cc93d8bcd6f7d86cac0c7c99485cc1837fc87790f13f5bbf2f3b901e86be341210376c2519a09f7cfbcbe000f823c1e957cadace64296e907ae1ea36536313f0706feffffffcb551ea63f1903d74888a1f8989fb521df1567740a0edcf862fbd4f372783236010000006b483045022100f0fa4ab61952f5497d4dfe6b7a4e5e9b3305a5b94b0d7a6d1536927dfea2aa4202205eaa36045812bf5561c9e3a38fc92e6d89908637a5c0ec250a843f36355418c2412103f6e8ebb2836f89aedbe712fa91eda827df597a01fe1a19fa1658bc1d28d1ad15feffffff02bccdf505000000001976a914349256bff9dbe79285454d0e55d1a3163bd6dff888ac0084d717000000001976a914e2e4d329a79401e0a713210a4c615abdc540eda888ac68000000",
			proofOverrides: map[string]*proofOverride{
				"a7a97cb7650b8ff1ee294f86653664adfbaef112d8813784a28b1c735b550d25": {
					err: errBiggerBadder,
				},
			},
			expErr: errFailedToGetMerkleProofA7a97c,
		},
		{
			name:    "ancestry needing multiple layers can be built",
			tx:      "02000000054e5113bf8330981b4f063d95d18540bc09016843347437ace0133f27c830a518000000006a47304402204455d7f4e1ccf6c8aa47e7b99349e02aadc80474e3c8b9b479d2d5aa98eb954402206a3ac9cc086bb03cb4a9db872fbb46d682f1036dd5e51d910484d62651eaad124121024d62a2e85bd2c310a2a43b2f17e4312381209276424fb634bd89325068997091feffffff531d1d4246046ad2e67636965b682509d0d021badf1ed95177b2e27a80e2020a000000006b483045022100a0882e9db7ff299178211b7223ee891119d22374c8e2ce21aec00daaad5e170102207af1d17b4d9ed4f9684e520f709239fded50f3ea7d5ec7a9e458ab00a2efffea412102e9fab6eb7648af9cd3719224062bfec10863ea9ba7a2caf26384705bc4387569feffffff531d1d4246046ad2e67636965b682509d0d021badf1ed95177b2e27a80e2020a010000006b483045022100b1d5b94ce9b5a1c8b531757aff4439bfb426dc439c3e35f6cdc3a08c9657ea0c02203c8490252d52db4557032042a5dcc8c3a2cdd9378dfd38f5ca03a9940ca055a2412103b05c23dcb06d192490d83c39bfbb515dfd0c0c172ae9314c91824b8936459132feffffffcd6b25094891d60413d5b50a4f6309574d648b9af6dbe782aa3510b18a0759fa010000006b483045022100bda388d23ccbea1beb3b1b55387600ae1b54759844beca4e9bcd405c2665b1a902206e536756978ffc3812c12f6eeeb82de351e13ad70d9e93c2358eb8a3c315330d4121024d62a2e85bd2c310a2a43b2f17e4312381209276424fb634bd89325068997091feffffff99995c0c19cda01cec64b7e20a3df37920bd3ad1ba6e666e337cb37d7829b354000000006a47304402205d6ebce84aeaf337c81894ffce89cc21a59a8659d07a599bb3371e5fee89153c02204432f8fae576e0aa47122d2dfbc914a6a34813fba86ecd0fcf0dd7ca65d199df4121034fc957d2e5b7299624148b914e830c763532983979c4cb3df29cca2751982d9afeffffff0282c3f505000000001976a9147fcac6c6eabf96ea58e561d3880f8e68bc58574088ac0027b929000000001976a9142d58ec9527c3df1ada4cac18cde85beb57cb199588ac68000000",
			expFile: "spvancestor_1",
		},
		{
			name: "missing tx multiple layers down causes error",
			tx:   "02000000054e5113bf8330981b4f063d95d18540bc09016843347437ace0133f27c830a518000000006a47304402204455d7f4e1ccf6c8aa47e7b99349e02aadc80474e3c8b9b479d2d5aa98eb954402206a3ac9cc086bb03cb4a9db872fbb46d682f1036dd5e51d910484d62651eaad124121024d62a2e85bd2c310a2a43b2f17e4312381209276424fb634bd89325068997091feffffff531d1d4246046ad2e67636965b682509d0d021badf1ed95177b2e27a80e2020a000000006b483045022100a0882e9db7ff299178211b7223ee891119d22374c8e2ce21aec00daaad5e170102207af1d17b4d9ed4f9684e520f709239fded50f3ea7d5ec7a9e458ab00a2efffea412102e9fab6eb7648af9cd3719224062bfec10863ea9ba7a2caf26384705bc4387569feffffff531d1d4246046ad2e67636965b682509d0d021badf1ed95177b2e27a80e2020a010000006b483045022100b1d5b94ce9b5a1c8b531757aff4439bfb426dc439c3e35f6cdc3a08c9657ea0c02203c8490252d52db4557032042a5dcc8c3a2cdd9378dfd38f5ca03a9940ca055a2412103b05c23dcb06d192490d83c39bfbb515dfd0c0c172ae9314c91824b8936459132feffffffcd6b25094891d60413d5b50a4f6309574d648b9af6dbe782aa3510b18a0759fa010000006b483045022100bda388d23ccbea1beb3b1b55387600ae1b54759844beca4e9bcd405c2665b1a902206e536756978ffc3812c12f6eeeb82de351e13ad70d9e93c2358eb8a3c315330d4121024d62a2e85bd2c310a2a43b2f17e4312381209276424fb634bd89325068997091feffffff99995c0c19cda01cec64b7e20a3df37920bd3ad1ba6e666e337cb37d7829b354000000006a47304402205d6ebce84aeaf337c81894ffce89cc21a59a8659d07a599bb3371e5fee89153c02204432f8fae576e0aa47122d2dfbc914a6a34813fba86ecd0fcf0dd7ca65d199df4121034fc957d2e5b7299624148b914e830c763532983979c4cb3df29cca2751982d9afeffffff0282c3f505000000001976a9147fcac6c6eabf96ea58e561d3880f8e68bc58574088ac0027b929000000001976a9142d58ec9527c3df1ada4cac18cde85beb57cb199588ac68000000",
			txOverrides: map[string]*txOverride{
				"a8e760210a8e3646ded829860745c60a2ade443d6998465dd0da5ae5f37b3b8e": nil,
			},
			expErr: errCouldNotFindTxA8e76021,
		},
		{
			name: "error getting tx multiple layers down is handled",
			tx:   "02000000054e5113bf8330981b4f063d95d18540bc09016843347437ace0133f27c830a518000000006a47304402204455d7f4e1ccf6c8aa47e7b99349e02aadc80474e3c8b9b479d2d5aa98eb954402206a3ac9cc086bb03cb4a9db872fbb46d682f1036dd5e51d910484d62651eaad124121024d62a2e85bd2c310a2a43b2f17e4312381209276424fb634bd89325068997091feffffff531d1d4246046ad2e67636965b682509d0d021badf1ed95177b2e27a80e2020a000000006b483045022100a0882e9db7ff299178211b7223ee891119d22374c8e2ce21aec00daaad5e170102207af1d17b4d9ed4f9684e520f709239fded50f3ea7d5ec7a9e458ab00a2efffea412102e9fab6eb7648af9cd3719224062bfec10863ea9ba7a2caf26384705bc4387569feffffff531d1d4246046ad2e67636965b682509d0d021badf1ed95177b2e27a80e2020a010000006b483045022100b1d5b94ce9b5a1c8b531757aff4439bfb426dc439c3e35f6cdc3a08c9657ea0c02203c8490252d52db4557032042a5dcc8c3a2cdd9378dfd38f5ca03a9940ca055a2412103b05c23dcb06d192490d83c39bfbb515dfd0c0c172ae9314c91824b8936459132feffffffcd6b25094891d60413d5b50a4f6309574d648b9af6dbe782aa3510b18a0759fa010000006b483045022100bda388d23ccbea1beb3b1b55387600ae1b54759844beca4e9bcd405c2665b1a902206e536756978ffc3812c12f6eeeb82de351e13ad70d9e93c2358eb8a3c315330d4121024d62a2e85bd2c310a2a43b2f17e4312381209276424fb634bd89325068997091feffffff99995c0c19cda01cec64b7e20a3df37920bd3ad1ba6e666e337cb37d7829b354000000006a47304402205d6ebce84aeaf337c81894ffce89cc21a59a8659d07a599bb3371e5fee89153c02204432f8fae576e0aa47122d2dfbc914a6a34813fba86ecd0fcf0dd7ca65d199df4121034fc957d2e5b7299624148b914e830c763532983979c4cb3df29cca2751982d9afeffffff0282c3f505000000001976a9147fcac6c6eabf96ea58e561d3880f8e68bc58574088ac0027b929000000001976a9142d58ec9527c3df1ada4cac18cde85beb57cb199588ac68000000",
			txOverrides: map[string]*txOverride{
				"8fff14aabcad6ebd1da2ce7737751e1e1e4c9c0d4c5b8abfef7e7914556d7965": {
					err: errCloseButNoCigar,
				},
			},
			expErr: errFailedToGetTx8fff14aa,
		},
		{
			name: "error getting merkle proof multiple layers down is handled",
			tx:   "02000000054e5113bf8330981b4f063d95d18540bc09016843347437ace0133f27c830a518000000006a47304402204455d7f4e1ccf6c8aa47e7b99349e02aadc80474e3c8b9b479d2d5aa98eb954402206a3ac9cc086bb03cb4a9db872fbb46d682f1036dd5e51d910484d62651eaad124121024d62a2e85bd2c310a2a43b2f17e4312381209276424fb634bd89325068997091feffffff531d1d4246046ad2e67636965b682509d0d021badf1ed95177b2e27a80e2020a000000006b483045022100a0882e9db7ff299178211b7223ee891119d22374c8e2ce21aec00daaad5e170102207af1d17b4d9ed4f9684e520f709239fded50f3ea7d5ec7a9e458ab00a2efffea412102e9fab6eb7648af9cd3719224062bfec10863ea9ba7a2caf26384705bc4387569feffffff531d1d4246046ad2e67636965b682509d0d021badf1ed95177b2e27a80e2020a010000006b483045022100b1d5b94ce9b5a1c8b531757aff4439bfb426dc439c3e35f6cdc3a08c9657ea0c02203c8490252d52db4557032042a5dcc8c3a2cdd9378dfd38f5ca03a9940ca055a2412103b05c23dcb06d192490d83c39bfbb515dfd0c0c172ae9314c91824b8936459132feffffffcd6b25094891d60413d5b50a4f6309574d648b9af6dbe782aa3510b18a0759fa010000006b483045022100bda388d23ccbea1beb3b1b55387600ae1b54759844beca4e9bcd405c2665b1a902206e536756978ffc3812c12f6eeeb82de351e13ad70d9e93c2358eb8a3c315330d4121024d62a2e85bd2c310a2a43b2f17e4312381209276424fb634bd89325068997091feffffff99995c0c19cda01cec64b7e20a3df37920bd3ad1ba6e666e337cb37d7829b354000000006a47304402205d6ebce84aeaf337c81894ffce89cc21a59a8659d07a599bb3371e5fee89153c02204432f8fae576e0aa47122d2dfbc914a6a34813fba86ecd0fcf0dd7ca65d199df4121034fc957d2e5b7299624148b914e830c763532983979c4cb3df29cca2751982d9afeffffff0282c3f505000000001976a9147fcac6c6eabf96ea58e561d3880f8e68bc58574088ac0027b929000000001976a9142d58ec9527c3df1ada4cac18cde85beb57cb199588ac68000000",
			proofOverrides: map[string]*proofOverride{
				"3fa9a1a20c1f4dbd2c3dd5866749621ebe15c7290c13080ba0d45ab9c649cc15": {
					err: errOhNo,
				},
			},
			expErr: errFailedToGetMerkleProof3fa9a1,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testTx, err := bt.NewTxFromString(test.tx)
			require.NoError(t, err)

			mock := &mockTxMerkleGetter{
				txStoreFunc: newTxFunc(t, test.txOverrides),
				mpStoreFunc: newMerkleProofFunc(t, test.proofOverrides),
			}

			c, err := spv.NewEnvelopeCreator(mock, mock)
			require.NoError(t, err)

			ancestry, err := c.CreateTxAncestry(context.Background(), testTx)
			if test.expErr == nil {
				require.NoError(t, err)
				assert.NotNil(t, ancestry)

				bb, readErr := data.SpvCreateData.Load(test.expFile + ".json")
				require.NoError(t, readErr)

				var env spv.AncestryJSON
				require.NoError(t, json.NewDecoder(bytes.NewReader(bb)).Decode(&env))
				require.Equal(t, env, *ancestry)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, test.expErr.Error())
			}
		})
	}
}
