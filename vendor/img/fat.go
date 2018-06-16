package img

import (
    "bytes"
    "encoding/binary"
)

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

func DecodeFileEntry(rawbytes []byte) (*rawFileEntry, error) {
    var rawentry rawFileEntry

    r := bytes.NewReader(rawbytes)
    e := binary.Read(r, binary.LittleEndian, &rawentry)
    if e != nil {
        return nil, e
    }

    return &rawentry, nil
}
