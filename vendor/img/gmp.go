package img

import (
    "disk"
    "bytes"
    "encoding/binary"
    "io"
)

type GmpSubfile struct {
    Offset uint32
    Length uint32
}

type GmpDirectoryEntry struct {
    GmpSubfile
    SubfileHeader
    RawHeader []byte
}

func DecodeGmpHeader(hdrbytes []byte, gmpsize uint32) ([]GmpSubfile, error) {
    commhdr, err := DecodeSubfileCommonHeader(hdrbytes)
    if err != nil {
        return nil, err
    }
    if commhdr.Format != "GMP" {
        return nil, ErrBadSignature
    }

    rawtable := hdrbytes[:commhdr.HeaderSize]
    rawtable = rawtable[SubfileCommonHeaderSize:]

    if len(rawtable)%4 != 0 {
        return nil, ErrBrokenFileTable
    }

    r := bytes.NewReader(rawtable)
    res := make([]GmpSubfile, 0, (len(rawtable)/4)-1)

    for {
        var entry GmpSubfile
        e := binary.Read(r, binary.LittleEndian, &entry.Offset)
        if e == io.EOF {
            break
        }
        if e != nil {
            return nil, e
        }
        if entry.Offset > 0 {
            res = append(res, entry)
        }
    }

    for i := range res {
        isLast := i == (len(res) - 1)
        offset := res[i].Offset
        var nextOffset uint32
        if isLast {
            nextOffset = gmpsize
        } else {
            nextOffset = res[i+1].Offset
        }
        if nextOffset < offset {
            return nil, ErrBrokenFileTable
        }
        length := nextOffset - offset
        res[i].Length = length
    }

    return res, nil
}

func ReadGmpDirectory(imgfile disk.BlockReader, gmpentry *FileEntry, clusterblocks uint32) ([]GmpDirectoryEntry, error) {
    gmpsize := gmpentry.Size
    gmpfat := gmpentry.FAT

    gmpstart := int64(gmpfat[0]) * int64(clusterblocks)
    gmpfirstblock, err := imgfile.ReadBlock(gmpstart)
    if err != nil {
        return nil, err
    }

    subfiles, err := DecodeGmpHeader(gmpfirstblock[:], gmpsize)
    if err != nil {
        return nil, err
    }

    res := make([]GmpDirectoryEntry, len(subfiles))

    clustersize := int64(clusterblocks) * disk.BlockSize
    for i := range subfiles {
        offset := int64(subfiles[i].Offset)
        startcluster := offset / clustersize

        readstart := int64(gmpfat[startcluster]) * int64(clusterblocks)
        readblocks := int64(clusterblocks)
        skip := offset % clustersize
        if clustersize-skip < SubfileCommonHeaderSize {
            readblocks++
        }

        subfiledata, err := imgfile.ReadBlocks(readstart, readblocks)
        if err != nil {
            return nil, err
        }
        subfiledata = subfiledata[skip:]

        subfilehdr, err := DecodeSubfileCommonHeader(subfiledata)
        if err != nil {
            return nil, err
        }

        hdrlen := subfilehdr.HeaderSize
        if hdrlen > len(subfiledata) {
            hdrlen = len(subfiledata)
        }

        res[i].GmpSubfile = subfiles[i]
        res[i].SubfileHeader = *subfilehdr
        res[i].RawHeader = subfiledata[:hdrlen]
    }

    return res, nil
}
