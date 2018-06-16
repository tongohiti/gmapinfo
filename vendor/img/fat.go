package img

import (
    "bytes"
    "encoding/binary"
    "errors"
)

type FileEntry struct {
    Name string
    Size uint32
    FAT  []uint16
}

// File table entry
type rawFileEntry struct {
    Flag uint8
    Name [8]byte
    Ext  [3]byte
    Size uint32
    _    uint8
    Part uint8
    _    [14]byte
    FAT  [240]uint16
}

var (
    ErrUnknownFlagValue = errors.New("unknown file entry flag")
    ErrUnsupportedSplit = errors.New("not implemented: file split into 256 or more parts")
    ErrBrokenFAT        = errors.New("broken FAT")
    ErrBrokenFileTable  = errors.New("broken file table")
)

func DecodeFileEntry(rawbytes []byte) (*FileEntry, error) {
    var rawentry rawFileEntry

    r := bytes.NewReader(rawbytes)
    e := binary.Read(r, binary.LittleEndian, &rawentry)
    if e != nil {
        return nil, e
    }

    if rawentry.Flag == 0 {
        return nil, nil
    } else if rawentry.Flag != 1 {
        return nil, ErrUnknownFlagValue
    }

    if rawentry.Part == 0xFF { // part index may be larger than 8 bits
        return nil, ErrUnsupportedSplit
    }

    fat, ok := decodeFAT(rawentry.FAT)
    if !ok {
        return nil, ErrBrokenFAT
    }

    var entry FileEntry

    entry.Name = constructFileName(rawentry.Name[:], rawentry.Ext[:])
    entry.Size = rawentry.Size
    entry.FAT = fat

    return &entry, nil
}

func DecodeFileTable(rawbytes []byte) ([]FileEntry, error) {
    const EntrySize = 512
    if len(rawbytes)%EntrySize != 0 {
        return nil, ErrBrokenFileTable
    }
    n := len(rawbytes) / EntrySize

    res := make([]FileEntry, 0, n)

    var (
        rawentry  rawFileEntry
        preventry rawFileEntry
        entry     *FileEntry
    )

    r := bytes.NewReader(rawbytes)
    for i := 0; i < n; i++ {
        e := binary.Read(r, binary.LittleEndian, &rawentry)
        if e != nil {
            return nil, e
        }

        if rawentry.Flag == 0 {
            continue
        } else if rawentry.Flag != 1 {
            return nil, ErrUnknownFlagValue
        }

        if rawentry.Part == 0xFF { // part index may be larger than 8 bits
            return nil, ErrUnsupportedSplit
        }

        isnew := rawentry.Part == 0
        if i == 0 && !isnew {
            return nil, ErrBrokenFileTable
        }

        fat, ok := decodeFAT(rawentry.FAT)
        if !ok {
            return nil, ErrBrokenFAT
        }

        if isnew {
            n := len(res)
            res = res[:n+1]
            entry = &res[n]
        }

        if isnew {
            entry.Name = constructFileName(rawentry.Name[:], rawentry.Ext[:])
            entry.Size = rawentry.Size
            entry.FAT = fat
        } else {
            if rawentry.Name != preventry.Name || rawentry.Ext != preventry.Ext || rawentry.Size != 0 {
                return nil, ErrBrokenFileTable
            }
            entry.FAT = append(entry.FAT, fat...)
        }

        preventry = rawentry
    }

    return res, nil
}

func constructFileName(name, ext []byte) string {
    lname := strlen(name)
    lext := strlen(ext)
    var fullname [12]byte
    copy(fullname[0:], name[:lname])
    L := lname
    if lext > 0 {
        fullname[L] = '.'
        L++
        copy(fullname[L:], ext[:lext])
        L += lext
    }
    return string(fullname[:L])
}

func strlen(str []byte) int {
    i := bytes.IndexByte(str, ' ')
    if i < 0 {
        return len(str)
    } else {
        return i
    }
}

func decodeFAT(fat [240]uint16) ([]uint16, bool) {
    const NONE uint16 = 0xFFFF
    length := 0
    for length < 240 {
        if fat[length] == NONE {
            break
        }
        length++
    }
    for i := length; i < 240; i++ {
        if fat[i] != NONE {
            return fat[:length], false
        }
    }
    return fat[:length], true
}
