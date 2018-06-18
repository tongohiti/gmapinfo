package gmapinfo

import "fmt"

type SizeDescription struct {
    Bytes      int64
    Blocks     int64
    Clusters   int64
    BlockRem   int32
    ClusterRem int32
}

func (sz *SizeDescription) String() string {
    if sz.BlockRem == 0 && sz.ClusterRem == 0 {
        return fmt.Sprintf("%d bytes, %d blocks, %d clusters", sz.Bytes, sz.Blocks, sz.Clusters)
    } else { // Broken image file?
        return fmt.Sprintf("%d bytes, %d blocks (+%d bytes), %d clusters (+%d bytes) !! bad image file?", sz.Bytes, sz.Blocks, sz.BlockRem, sz.Clusters, sz.ClusterRem)
    }
}

func SizeFromBlockCount(blocks, blockSize, blocksInCluster uint32) *SizeDescription {
    var size SizeDescription

    size.Blocks = int64(blocks)
    size.Bytes = int64(blocks) * int64(blockSize)
    size.Clusters = int64(blocks) / int64(blocksInCluster)
    size.ClusterRem = int32(blocks%blocksInCluster) * int32(blockSize)

    return &size
}

func SizeFromByteCount(bytes int64, blockSize, clusterSize uint32) *SizeDescription {
    var size SizeDescription

    size.Bytes = bytes
    size.Blocks = bytes / int64(blockSize)
    size.Clusters = bytes / int64(clusterSize)
    size.BlockRem = int32(bytes % int64(blockSize))
    size.ClusterRem = int32(bytes % int64(clusterSize))

    return &size
}
