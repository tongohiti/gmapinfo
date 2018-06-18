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

    describeImageFile(imagefile, imgfile.SizeBytes(), hdr)

    // Read zero pages between header and file table
    allzeroes, err := readZeroes(imgfile, hdr.FileTableBlock)
    if err != nil {
        return err
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

    fatblocks := firstentry.Size/hdr.BlockSize - hdr.FileTableBlock

    describeImageFileHeader(hdr, firstentry, fatblocks, !allzeroes)

    // Check that block size is actually 512 bytes, otherwise further code would read incorect blocks
    if hdr.BlockSize != disk.BlockSize {
        return fmt.Errorf("unsupported block size: %d", hdr.BlockSize) // if encountered (unlikely), need to rework BlockReader
    }

    // Read whole file table
    filetable, err := imgfile.ReadBlocks(int64(hdr.FileTableBlock)+1, int64(fatblocks)-1)
    if err != nil {
        return err
    }

    files, err := img.DecodeFileTable(filetable)
    if err != nil {
        return err
    }

    describeSubfiles(imgfile, hdr, files)

    if params.Extract {
        err := extractFiles(imgfile, hdr.ClusterBlocks, files, params.OutputName, params.ZipOutput)
        if err != nil {
            return err
        }
    }

    return nil
}

func describeImageFile(imageFileName string, imageFileSize int64, hdr *img.Header) {
    fmt.Printf("Image file:   %s\n", imageFileName)
    fmt.Printf("Map name:     %s\n", hdr.MapName)
    fmt.Printf("Map version:  %v\n", hdr.MapVersion)
    fmt.Printf("Map date:     %v\n", hdr.MapDate)
    fmt.Printf("Timestamp:    %v\n", hdr.CreateDate)

    fmt.Printf("_BlockSize:   %d\n", hdr.BlockSize)
    fmt.Printf("_ClustBlks:   %d\n", hdr.ClusterBlocks)
    fmt.Printf("_ClustSize:   %d\n", hdr.ClusterSize)
    fmt.Printf("_NumClust:    %d\n", hdr.NumClusters)

    fmt.Printf("_FTabOffset:  0x%X (%d blocks)\n", hdr.FileTableBlock*hdr.BlockSize, hdr.FileTableBlock)

    for i := range hdr.PartitionTable {
        part := &hdr.PartitionTable[i]
        if !part.Empty {
            partSize := SizeFromBlockCount(part.NumSectors, hdr.BlockSize, hdr.ClusterBlocks)
            fmt.Printf("_Partition %d: %v\n", i, partSize)
        } else {
            if i == 0 {
                fmt.Println("!! Partition 0 empty - bad image file?")
            }
        }
    }

    fileSize := SizeFromByteCount(imageFileSize, hdr.BlockSize, hdr.ClusterSize)
    fmt.Printf("_File size:   %v\n", fileSize)
}

func describeImageFileHeader(hdr *img.Header, firstEntry *img.FileEntry, fatBlocks uint32, unparsedHeaderData bool) {
    headerSize := SizeFromByteCount(int64(firstEntry.Size), hdr.BlockSize, hdr.ClusterSize)
    fmt.Printf("_Header size: %v\n", headerSize)

    if firstEntry.Size < hdr.FileTableBlock*hdr.BlockSize {
        fmt.Println("!! Insufficient data size specified in first entry - bad image file?")
    }
    if unparsedHeaderData {
        fmt.Println("!! Non-zero data between img file header and FAT - bad image file?")
    }

    fmt.Printf("_Data start:  0x%X\n", firstEntry.Size)
    fmt.Printf("_Num entries: %d (0x%[1]X)\n", fatBlocks)
}

func describeSubfiles(imgfile disk.BlockReader, hdr *img.Header, files []img.FileEntry) error {
    fmt.Println("Subfiles:")

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
