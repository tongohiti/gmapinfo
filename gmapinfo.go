package main

import (
    "fmt"
    "os"
    "disk"
    "img"
)

func main() {
    fmt.Println("gmapinfo 0.0.2")

    imagefile := os.Args[1]
    fmt.Printf("Image file:  %s\n", imagefile)

    imgfile, err := disk.OpenImageFile(imagefile)
    if err != nil {
        panic(err)
    }
    defer imgfile.Close()

    hdrblock, err := imgfile.ReadBlock(0)
    if err != nil {
        panic(err)
    }

    hdr, err := img.DecodeHeader(hdrblock[:])
    if err != nil {
        panic(err)
    }

    fmt.Printf("Map name:    %s\n", hdr.MapName)
    fmt.Printf("Map version: %v\n", hdr.MapVersion)
    fmt.Printf("Map date:    %v\n", hdr.MapDate)
    fmt.Printf("Timestamp:   %v\n", hdr.CreateDate)

    fmt.Printf("_BlockSize:  %d\n", hdr.BlockSize)
    fmt.Printf("_ClustBlks:  %d\n", hdr.ClusterBlocks)
    fmt.Printf("_ClustSize:  %d\n", hdr.ClusterSize)
    fmt.Printf("_NumClust:   %d\n", hdr.NumClusters)

    fmt.Printf("_FTabOffs:   %d blk = 0x%X\n", hdr.FileTableBlock, hdr.FileTableBlock*hdr.BlockSize)

    partSectors := hdr.PartitionTable[0].NumSectors
    partBytes := partSectors * 512
    partClusters := partSectors / hdr.ClusterBlocks
    fmt.Printf("_Partition0: %d bytes, %d blocks, %d clusters // %v\n", partBytes, partSectors, partClusters, hdr.PartitionTable[0])
    if partSectors%hdr.ClusterBlocks != 0 {
        fmt.Println("!! Non-whole number of clusters in partition - bad image file?")
    }

    bytes := imgfile.SizeBytes()
    blocks, clusters := bytes/int64(hdr.BlockSize), bytes/int64(hdr.ClusterSize)
    fmt.Printf("File size:   %d bytes, %d blocks, %d clusters\n", bytes, blocks, clusters)
    xblocks, xclusters := bytes%int64(hdr.BlockSize), bytes%int64(hdr.ClusterSize)
    if xblocks != 0 {
        fmt.Println("!! Non-whole number of blocks - bad image file?")
    }
    if xclusters != 0 {
        fmt.Println("!! Non-whole number of clusters - bad image file?")
    }
}
