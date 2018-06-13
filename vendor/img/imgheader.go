package img

import (
    "bytes"
    "encoding/binary"
    "errors"
)

type Header struct {
    MapName        string
    MapVersion     Version
    MapDate        Date
    CreateDate     Timestamp
    PartitionTable [4]Partition
}

type Partition struct {
    Empty                   bool
    StartCHS, EndCHS        CHS
    StartLBA, EndLBA        uint32
    StartSector, NumSectors uint32
}

// File header (512 bytes) of an ".img" file
type rawHeader struct {
    _              [8]byte         // 0x000
    VersionMaj     uint8           // 0x008
    VersionMin     uint8           // 0x009
    ExpiryMonth    uint8           // 0x00A
    ExpiryYear     uint8           // 0x00B
    _              uint16          // 0x00C
    MapSourceFlag  uint8           // 0x00E
    Checksum       uint8           // 0x00F
    Signature      [6]byte         // 0x010
    BlockSize      uint16          // 0x016
    NumSectors     uint16          // 0x018
    NumHeads       uint16          // 0x01A
    NumCylinders   uint16          // 0x01C
    _              uint16          // 0x01E
    Unknown1       [25]uint8       // 0x020
    CreateYear     uint16          // 0x039 (unaligned!)
    CreateMonth    uint8           // 0x03B
    CreateDay      uint8           // 0x03C
    CreateHour     uint8           // 0x03D
    CreateMinute   uint8           // 0x03E
    CreateSecond   uint8           // 0x03F
    FileTableBlock uint8           // 0x040
    Garmin         [6]byte         // 0x041
    _              uint16          // 0x047
    MapDescr1      [20]byte        // 0x049
    Heads          uint16          // 0x05D (unaligned!)
    Sectors        uint16          // 0x05F (unaligned!)
    Exp1, Exp2     uint8           // 0x061, 0x062
    NumClusters    uint16          // 0x063 (unaligned!)
    MapDescr2      [30]byte        // 0x065
    Dead           uint16          // 0x083
    Unk3           uint16          // 0x085
    Unk4           uint8           // 0x087
    _              [310]byte       // 0x088
    PartitionTable [4]rawPartition // 0x1BE
    EndSignature   uint16          // 0x1FE
}

// Partition table entry
type rawPartition struct {
    BootStatus uint8
    StartCHS   [3]uint8
    Type       uint8
    EndCHS     [3]uint8
    StartLBA   uint32
    NumSectors uint32
}

var ErrBadSignature = errors.New("bad file signature")

func DecodeHeader(hdrbytes []byte) (*Header, error) {
    var rawhdr rawHeader

    r := bytes.NewReader(hdrbytes)
    e := binary.Read(r, binary.LittleEndian, &rawhdr)
    if e != nil {
        return nil, e
    }

    signature := string(rawhdr.Signature[:])
    if signature != "DSKIMG" && signature != "DSDIMG" {
        return nil, ErrBadSignature
    }

    garmin := string(rawhdr.Garmin[:])
    if garmin != "GARMIN" {
        return nil, ErrBadSignature
    }

    if rawhdr.Dead != 0xDEAD {
        return nil, ErrBadSignature
    }

    if rawhdr.EndSignature != 0xAA55 {
        return nil, ErrBadSignature
    }

    var header Header

    header.MapName = convertMapName(&rawhdr.MapDescr1, &rawhdr.MapDescr2)
    header.MapVersion = Version{rawhdr.VersionMaj, rawhdr.VersionMin}
    header.MapDate = convertDate(rawhdr.ExpiryYear, rawhdr.ExpiryMonth)
    header.CreateDate = convertTimestamp(rawhdr.CreateYear, rawhdr.CreateMonth, rawhdr.CreateDay, rawhdr.CreateHour, rawhdr.CreateMinute, rawhdr.CreateSecond)

    geometry := CHS{C: rawhdr.NumCylinders, H: rawhdr.NumHeads, S: rawhdr.NumSectors}
    for i := range header.PartitionTable {
        header.PartitionTable[i] = convertPartitionDescr(rawhdr.PartitionTable[i], geometry)
    }

    return &header, nil
}

func convertMapName(part1 *[20]byte, part2 *[30]byte) string {
    name := make([]byte, 0, 50)
    name = append(name, part1[:]...)
    name = append(name, part2[:]...)
    name = bytes.TrimRight(name, " ")
    return string(name)
}

func convertPartitionDescr(rp rawPartition, diskgeom CHS) Partition {
    var part Partition
    part.StartCHS = decodeCHS(rp.StartCHS)
    part.EndCHS = decodeCHS(rp.EndCHS)
    part.StartSector = rp.StartLBA
    part.NumSectors = rp.NumSectors
    part.Empty = part.StartCHS.isZero() && part.EndCHS.isZero()
    if !part.Empty {
        part.StartLBA = part.StartCHS.toLBA(diskgeom)
        part.EndLBA = part.EndCHS.toLBA(diskgeom)
    }
    return part
}
