package gmapinfo

import (
    "fmt"
    "disk"
    "img"
)

type Params struct {
    FileName   string
    Extract    bool
    ZipOutput  bool
    OutputName string
}

func Run(params Params) error {
    imagefile := params.FileName
    fmt.Printf("Image file:  %s\n", imagefile)

    imgfile, err := disk.OpenImageFile(imagefile)
    if err != nil {
        return err
    }
    defer imgfile.Close()

    // Read image header
    hdrblock, err := imgfile.ReadBlock(0)
    if err != nil {
        return err
    }

    hdr, err := img.DecodeHeader(hdrblock[:])
    if err != nil {
        return err
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
            return err
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
        return err
    }

    firstentry, err := img.DecodeFileEntry(firstentryblk[:])
    if err != nil {
        return err
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
    filetable, err := imgfile.ReadBlocks(int64(hdr.FileTableBlock)+1, int64(fatblocks)-1)
    if err != nil {
        return err
    }

    files, err := img.DecodeFileTable(filetable)
    if err != nil {
        return err
    }

    fmt.Printf("Num files:   %d (0x%[1]X)\n", len(files))

    for i := range files {
        entry := &files[i]
        fmt.Printf("Entry[%04d]: %v\n", i, *entry)
        err := describeSubfile(imgfile, entry, hdr.ClusterBlocks)
        if err != nil {
            return err
        }
    }

    if params.Extract {
        if hdr.BlockSize != disk.BlockSize {
            return fmt.Errorf("unsupported block size: %d", hdr.BlockSize)
        }

        err := extractFiles(imgfile, hdr.ClusterBlocks, files, params.OutputName, params.ZipOutput)
        if err != nil {
            return err
        }
    }

    return nil
}

func describeSubfile(imgfile disk.BlockReader, entry *img.FileEntry, clusterblocks uint32) error {
    firstblock := int64(entry.FAT[0]) * int64(clusterblocks)
    data, err := imgfile.ReadBlock(firstblock)
    if err != nil {
        return err
    }

    hdr, err := img.DecodeSubfileCommonHeader(data[:])
    if err == img.ErrBadSignature {
        fmt.Println("             Not a GARMIN common format.")
        return nil
    }
    if err != nil {
        return err
    }

    fmt.Printf("             %v\n", *hdr)

    return nil
}
