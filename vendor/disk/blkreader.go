package disk

const BlockSize = 512

type Block [BlockSize]byte

type BlockReader interface {
    ReadBlock(index int64) (*Block, error)
    ReadBlocks(index, count int64) ([]byte, error)
}
