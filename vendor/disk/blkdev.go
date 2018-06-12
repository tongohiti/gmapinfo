package disk

const BlockSize = 512

type Block [BlockSize]byte

type BlockDevice interface {
    ReadBlock(n int64) (*Block, error)
}
