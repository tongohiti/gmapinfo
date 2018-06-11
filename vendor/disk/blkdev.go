package disk

const SectorSize = 512

type Sector [SectorSize]byte

type BlockDevice interface {
    ReadSector(n int64) (*Sector, error)
}
