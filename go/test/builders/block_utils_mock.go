package builders

import (
	"github.com/orbs-network/go-mock"
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
)

type MockBlockUtils struct {
	mock.Mock
}

func NewMockBlockUtils() *MockBlockUtils {
	return &MockBlockUtils{}
}

func CalculateBlockHash(block *lh.Block) lh.BlockHash {
	// TODO Do a real hash here
	// export const calculateBlockHash = (block: Block): Buffer => createHash("sha256").update(stringify(block.header)).digest(); // .digest("base64");
	return lh.BlockHash("0123456789ABCDEF")
}

func (*MockBlockUtils) CalculateBlockHash(block *lh.Block) lh.BlockHash {
	return CalculateBlockHash(block)
}
