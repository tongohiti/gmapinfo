package disk

const BlockSize = 512

type Block [BlockSize]byte

type BlockReader interface {
    ReadBlock(n int64) (*Block, error)
}
