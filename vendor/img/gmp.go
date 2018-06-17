package img

import (
    "bytes"
    "encoding/binary"
    "io"
)

type GmpSubfile struct {
    Offset uint32
    Length uint32
}

func DecodeGmpHeader(hdrbytes []byte, gmpsize uint32) ([]GmpSubfile, error) {
    commhdr, err := DecodeSubfileCommonHeader(hdrbytes)
    if err != nil {
        return nil, err
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
