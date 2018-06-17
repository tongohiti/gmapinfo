package img

import (
    "bytes"
    "encoding/binary"
)

const SubfileCommonHeaderSize = 21

type SubfileHeader struct {
    HeaderSize int
    Format     string
    CreateDate Timestamp
    Locked     bool
}

type rawSubfileHeader struct {
    HeaderSize   uint16   // 0x00
    FormatID     [10]byte // 0x02
    _            byte     // 0x0C
    Flag         byte     // 0x0D
    CreateYear   uint16   // 0x0E
    CreateMonth  uint8    // 0x10
    CreateDay    uint8    // 0x11
    CreateHour   uint8    // 0x12
    CreateMinute uint8    // 0x13
    CreateSecond uint8    // 0x14
}

func DecodeSubfileCommonHeader(hdrbytes []byte) (*SubfileHeader, error) {
    var rawhdr rawSubfileHeader

    r := bytes.NewReader(hdrbytes)
    e := binary.Read(r, binary.LittleEndian, &rawhdr)
    if e != nil {
        return nil, e
    }

    signature := string(rawhdr.FormatID[:])
    if signature[:6] != "GARMIN" || signature[6] != ' ' {
        return nil, ErrBadSignature
    }

    format := signature[7:]
    if ok := format[0] >= 'A' && format[0] <= 'Z' &&
        format[1] >= 'A' && format[1] <= 'Z' &&
        format[2] >= 'A' && format[2] <= 'Z';
        !ok {
        return nil, ErrBadSignature
    }

    var header SubfileHeader

    header.HeaderSize = int(rawhdr.HeaderSize)
    header.Format = format
    header.CreateDate = convertTimestamp(rawhdr.CreateYear, rawhdr.CreateMonth, rawhdr.CreateDay, rawhdr.CreateHour, rawhdr.CreateMinute, rawhdr.CreateSecond)
    header.Locked = rawhdr.Flag != 0

    return &header, nil
}
