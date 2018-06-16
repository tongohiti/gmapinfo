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

    // Read image header
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

    // Read zero pages between header and file table
    nonzero := false
    for i := int64(1); i < int64(hdr.FileTableBlock); i++ {
        zeroes, err := imgfile.ReadBlock(i)
        if err != nil {
            panic(err)
        }
        for _, z := range zeroes {
            if z != 0 {
                nonzero = true
                break
            }
        }
        if nonzero {
            break
        }
    }
    if nonzero {
        fmt.Println("!! Non-zero data between image header and FAT - bad image file?")
    }

    // Read first (fake) file entry
    firstentryblk, err := imgfile.ReadBlock(int64(hdr.FileTableBlock))
    if err != nil {
        panic(err)
    }

    firstentry, err := img.DecodeFileEntry(firstentryblk[:])
    if err != nil {
        panic(err)
    }

    fmt.Printf("Entry0:      %v\n", *firstentry)
    if firstentry.Size%hdr.BlockSize != 0 {
        fmt.Println("!! Non-whole number of blocks in first entry - bad image file?")
    }
    if firstentry.Size < hdr.FileTableBlock*hdr.BlockSize {
        fmt.Println("!! Too small data size in first entry - bad image file?")
    }
    fatblocks := firstentry.Size/hdr.BlockSize - hdr.FileTableBlock
    fmt.Printf("Num entries: %d (0x%[1]X)\n", fatblocks)
    fmt.Printf("Data start:  0x%X\n", firstentry.Size)

    // Read whole file table
    for i := uint32(1); i < fatblocks; i++ {
        entryblk, err := imgfile.ReadBlock(int64(hdr.FileTableBlock + i))
        if err != nil {
            panic(err)
        }

        entry, err := img.DecodeFileEntry(entryblk[:])
        if err != nil {
            panic(err)
        }

        fmt.Printf("Entry[%04d]: %v\n", i, entry)
    }
}
