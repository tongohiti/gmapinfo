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
    fmt.Printf("Image file:   %s\n", imagefile)

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

    fmt.Printf("Map name:     %s\n", hdr.MapName)
    fmt.Printf("Map version:  %v\n", hdr.MapVersion)
    fmt.Printf("Map date:     %v\n", hdr.MapDate)
    fmt.Printf("Timestamp:    %v\n", hdr.CreateDate)

    fmt.Printf("_BlockSize:   %d\n", hdr.BlockSize)
    fmt.Printf("_ClustBlks:   %d\n", hdr.ClusterBlocks)
    fmt.Printf("_ClustSize:   %d\n", hdr.ClusterSize)
    fmt.Printf("_NumClust:    %d\n", hdr.NumClusters)

    fmt.Printf("_FTabOffset:  0x%X (%d blocks)\n", hdr.FileTableBlock*hdr.BlockSize, hdr.FileTableBlock)

    partSectors := hdr.PartitionTable[0].NumSectors
    partBytes := partSectors * 512
    partClusters := partSectors / hdr.ClusterBlocks
    fmt.Printf("_Partition 0: %d bytes, %d blocks, %d clusters\n", partBytes, partSectors, partClusters)
    if partSectors%hdr.ClusterBlocks != 0 {
        fmt.Println("!! Non-whole number of clusters in partition - bad image file?")
    }

    bytes := imgfile.SizeBytes()
    blocks, clusters := bytes/int64(hdr.BlockSize), bytes/int64(hdr.ClusterSize)
    fmt.Printf("_File size:   %d bytes, %d blocks, %d clusters\n", bytes, blocks, clusters)
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

    fmt.Printf("_Header size: %d bytes\n", firstentry.Size)
    if firstentry.Size%hdr.BlockSize != 0 {
        fmt.Println("!! Non-whole number of blocks in first entry - bad image file?")
    }
    if firstentry.Size < hdr.FileTableBlock*hdr.BlockSize {
        fmt.Println("!! Too small data size in first entry - bad image file?")
    }
    fatblocks := firstentry.Size/hdr.BlockSize - hdr.FileTableBlock
    fmt.Printf("_Data start:  0x%X\n", firstentry.Size)
    fmt.Printf("_Num entries: %d (0x%[1]X)\n", fatblocks)

    // Read whole file table
    filetable, err := imgfile.ReadBlocks(int64(hdr.FileTableBlock)+1, int64(fatblocks)-1)
    if err != nil {
        return err
    }

    files, err := img.DecodeFileTable(filetable)
    if err != nil {
        return err
    }

    fmt.Printf("_Num files:   %d (0x%[1]X)\n", len(files))

    printFunc := func(descr *SubfileDescription) {
        var prefix string
        if descr.Nested {
            prefix = "    >>"
        } else {
            prefix = "  >"
        }
        fmt.Printf("%s %s, %d bytes", prefix, descr.Name, descr.Size)
        if descr.Attrs {
            fmt.Printf(", %v, locked=%t", descr.Date, descr.Locked)
        }
        if descr.MapId != 0 {
            fmt.Printf(", MapID %d (0x%[1]X)", descr.MapId)
        }
        fmt.Println()
    }

    for i := range files {
        entry := &files[i]
        err := describeSubfile(imgfile, entry, hdr.ClusterBlocks, printFunc)
        if err != nil {
            return err
        }
    }

    fmt.Printf("Total %d subfiles.\n", len(files))

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

type SubfileDescription struct {
    Name   string
    Size   uint32
    Date   img.Timestamp
    MapId  uint32
    Attrs  bool
    Locked bool
    Nested bool
}

type PrintFunc func(*SubfileDescription)

func describeSubfile(imgfile disk.BlockReader, entry *img.FileEntry, clusterblocks uint32, print PrintFunc) error {
    firstblock := int64(entry.FAT[0]) * int64(clusterblocks)
    data, err := imgfile.ReadBlock(firstblock)
    if err != nil {
        return err
    }

    hdr, err := img.DecodeSubfileCommonHeader(data[:])
    missingCommonHeader := err == img.ErrBadSignature
    if err != nil && !missingCommonHeader {
        return err
    }

    var descr SubfileDescription
    descr.Name = entry.Name
    descr.Size = entry.Size

    if missingCommonHeader {
        descr.Attrs = false
        print(&descr)
        return nil
    } else {
        descr.Attrs = true
        descr.Date = hdr.CreateDate
        descr.Locked = hdr.Locked
    }

    if hdr.Format == "TRE" {
        mapId, err := img.ReadTreMapId(data[:])
        if err == nil {
            descr.MapId = mapId
        }
    }

    print(&descr)

    if hdr.Format == "GMP" {
        gmpdirectory, err := img.ReadGmpDirectory(imgfile, entry, clusterblocks)
        if err != nil {
            return err
        }

        for _, e := range gmpdirectory {
            var descr SubfileDescription
            descr.Nested = true
            descr.Attrs = true
            descr.Name = e.Format
            descr.Size = e.Length
            descr.Date = e.SubfileHeader.CreateDate
            descr.Locked = e.SubfileHeader.Locked

            if e.Format == "TRE" {
                mapId, err := img.ReadTreMapId(e.RawHeader)
                if err == nil {
                    descr.MapId = mapId
                }
            }

            print(&descr)
        }
    }

    return nil
}
