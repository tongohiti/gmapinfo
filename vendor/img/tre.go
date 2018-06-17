package img

import (
    "bytes"
    "encoding/binary"
    "errors"
)

var ErrBadHeader   = errors.New("bad file header")

func ReadTreMapId(hdrbytes []byte) (uint32, error) {
    const MapIdOffset = 0x74

    if len(hdrbytes) < MapIdOffset + 4 {
        return 0, ErrBadHeader
    }

    var hdrsize uint16
    r := bytes.NewReader(hdrbytes)
    e := binary.Read(r, binary.LittleEndian, &hdrsize)
    if e != nil {
        return 0, e
    }
    if hdrsize < MapIdOffset + 4 {
        return 0, ErrBadHeader
    }

    idbytes := hdrbytes[MapIdOffset:]
    var mapId uint32
    r = bytes.NewReader(idbytes)
    e = binary.Read(r, binary.LittleEndian, &mapId)
    return mapId, e
}
