package bc_test

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/bsv-blockchain/go-bc"

	"github.com/stretchr/testify/require"
)

func TestBuildingMerklePathBinary(t *testing.T) {
	t.Parallel()

	// build example merkle path data.
	merklePath := bc.MerklePath{
		Index: 136,
		Path: []string{
			"6cf512411d03ab9b61643515e7aa9afd005bf29e1052ade95410b3475f02820c",
			"cd73c0c6bb645581816fa960fd2f1636062fcbf23cb57981074ab8d708a76e3b",
			"b4c8d919190a090e77b73ffcd52b85babaaeeb62da000473102aca7f070facef",
			"3470d882cf556a4b943639eba15dc795dffdbebdc98b9a98e3637fda96e3811e",
		},
	}

	// build binary path from it.
	merklePathBinary, err := merklePath.Bytes()
	if err != nil {
		t.Error(err)
		return
	}

	mp, _ := hex.DecodeString("88040c82025f47b31054e9ad52109ef25b00fd9aaae7153564619bab031d4112f56c3b6ea708d7b84a078179b53cf2cb2f0636162ffd60a96f81815564bbc6c073cdefac0f077fca2a10730400da62ebaebaba852bd5fc3fb7770e090a1919d9c8b41e81e396da7f63e3989a8bc9bdbefddf95c75da1eb3936944b6a55cf82d87034")
	if err != nil {
		t.Error(err)
		return
	}

	// assert binary path is expected.
	require.Equal(t, mp, merklePathBinary)
}

func TestDecodingMerklePathBinary(t *testing.T) {
	t.Parallel()

	merklePath, err := bc.NewMerklePathFromStr("88040c82025f47b31054e9ad52109ef25b00fd9aaae7153564619bab031d4112f56c3b6ea708d7b84a078179b53cf2cb2f0636162ffd60a96f81815564bbc6c073cdefac0f077fca2a10730400da62ebaebaba852bd5fc3fb7770e090a1919d9c8b41e81e396da7f63e3989a8bc9bdbefddf95c75da1eb3936944b6a55cf82d87034")
	if err != nil {
		t.Error(err)
		return
	}

	// data we are expecting to deserialize
	// merklePathData := bc.MerklePathData{
	// 	Index: 136,
	// 	Path: []string{"6cf512411d03ab9b61643515e7aa9afd005bf29e1052ade95410b3475f02820c",
	// 		"cd73c0c6bb645581816fa960fd2f1636062fcbf23cb57981074ab8d708a76e3b",
	// 		"b4c8d919190a090e77b73ffcd52b85babaaeeb62da000473102aca7f070facef",
	// 		"3470d882cf556a4b943639eba15dc795dffdbebdc98b9a98e3637fda96e3811e"},
	// }

	// assert binary path is expected.
	require.Equal(t, uint64(136), merklePath.Index)
	require.Len(t, merklePath.Path, 4)
	require.Equal(t, "6cf512411d03ab9b61643515e7aa9afd005bf29e1052ade95410b3475f02820c", merklePath.Path[0])
	require.Equal(t, "cd73c0c6bb645581816fa960fd2f1636062fcbf23cb57981074ab8d708a76e3b", merklePath.Path[1])
	require.Equal(t, "b4c8d919190a090e77b73ffcd52b85babaaeeb62da000473102aca7f070facef", merklePath.Path[2])
	require.Equal(t, "3470d882cf556a4b943639eba15dc795dffdbebdc98b9a98e3637fda96e3811e", merklePath.Path[3])
}

func TestGetMerklePath(t *testing.T) {
	txids := []string{
		"b6d4d13aa08bb4b6cdb3b329cef29b5a5d55d85a85c330d56fddbce78d99c7d6",
		"426f65f6a6ce79c909e54d8959c874a767db3076e76031be70942b896cc64052",
		"adc23d36cc457d5847968c2e4d5f017a6f12a2f165102d10d2843f5276cfe68e",
		"728714bbbddd81a54cae473835ae99eb92ed78191327eb11a9d7494273dcad2a",
		"e3aa0230aa81abd483023886ad12790acf070e2a9f92d7f0ae3bebd90a904361",
		"4848b9e94dd0e4f3173ebd6982ae7eb6b793de305d8450624b1d86c02a5c61d9",
		"912f77eefdd311e24f96850ed8e701381fc4943327f9cf73f9c4dec0d93a056d",
		"397fe2ae4d1d24efcc868a02daae42d1b419289d9a1ded3a5fe771efcc1219d9",
	}

	expected := "1a1e779cd7dfc59f603b4e88842121001af822b2dc5d3b167ae66152e586a6b0"

	merkles, err := bc.BuildMerkleTreeStore(txids)
	require.NoError(t, err)

	// build path for tx index 4.
	path := bc.GetTxMerklePath(4, merkles)
	// Safe conversion: Index is bounded by the number of transactions in the block
	pathIndex := int(path.Index) //nolint:gosec // Index bounded by transaction count
	root, err := bc.MerkleRootFromBranches("e3aa0230aa81abd483023886ad12790acf070e2a9f92d7f0ae3bebd90a904361", pathIndex, path.Path)
	require.NoError(t, err)
	require.Equal(t, expected, root)

	// build path for tx index 3.
	path = bc.GetTxMerklePath(3, merkles)
	root, err = path.CalculateRoot("728714bbbddd81a54cae473835ae99eb92ed78191327eb11a9d7494273dcad2a")
	require.NoError(t, err)
	require.Equal(t, expected, root)
}

func TestGetMerklePathOddPosition(t *testing.T) {
	txids := []string{
		"b6d4d13aa08bb4b6cdb3b329cef29b5a5d55d85a85c330d56fddbce78d99c7d6",
		"426f65f6a6ce79c909e54d8959c874a767db3076e76031be70942b896cc64052",
		"adc23d36cc457d5847968c2e4d5f017a6f12a2f165102d10d2843f5276cfe68e",
		"728714bbbddd81a54cae473835ae99eb92ed78191327eb11a9d7494273dcad2a",
		"e3aa0230aa81abd483023886ad12790acf070e2a9f92d7f0ae3bebd90a904361",
	}

	merkles, err := bc.BuildMerkleTreeStore(txids)
	require.NoError(t, err)

	// build path for tx index 4.
	path := bc.GetTxMerklePath(4, merkles)
	// Safe conversion: Index is bounded by the number of transactions in the block
	pathIndex := int(path.Index) //nolint:gosec // Index bounded by transaction count
	root, err := bc.MerkleRootFromBranches("e3aa0230aa81abd483023886ad12790acf070e2a9f92d7f0ae3bebd90a904361", pathIndex, path.Path)
	require.NoError(t, err)
	require.Equal(t, merkles[len(merkles)-1], root)
}

func TestGetMerklePathEmptyPath(t *testing.T) {
	txids := []string{
		"b6d4d13aa08bb4b6cdb3b329cef29b5a5d55d85a85c330d56fddbce78d99c7d6",
	}

	merkles, err := bc.BuildMerkleTreeStore(txids)
	require.NoError(t, err)

	// build path for tx index 4.
	path := bc.GetTxMerklePath(0, merkles)
	// Safe conversion: Index is bounded by the number of transactions in the block
	pathIndex := int(path.Index) //nolint:gosec // Index bounded by transaction count
	root, err := bc.MerkleRootFromBranches("b6d4d13aa08bb4b6cdb3b329cef29b5a5d55d85a85c330d56fddbce78d99c7d6", pathIndex, path.Path)
	require.NoError(t, err)
	require.Equal(t, merkles[len(merkles)-1], root)
	require.Equal(t, ([]string)(nil), path.Path)
	require.Equal(t, uint64(0), path.Index)
}

func TestGetMerklePathEmptyPathJson(t *testing.T) {
	txids := []string{
		"b6d4d13aa08bb4b6cdb3b329cef29b5a5d55d85a85c330d56fddbce78d99c7d6",
	}

	merkles, _ := bc.BuildMerkleTreeStore(txids)
	path := bc.GetTxMerklePath(0, merkles)
	js, err := json.Marshal(path)
	require.NoError(t, err)
	require.Equal(t, "{\"index\":0,\"path\":null}", string(js))
}
