package img

type CHS struct {
    C, H, S uint16
}

func (chs CHS) isZero() bool {
    return chs.C == 0 && chs.H == 0 && chs.S == 0
}

// LBA = (C × HPC + H) × SPT + (S - 1)
func (chs CHS) toLBA(diskGeometry CHS) uint32 {
    hpc := uint32(diskGeometry.H) // heads per cylinder
    spt := uint32(diskGeometry.S) // sectors per track
    c := uint32(chs.C)
    h := uint32(chs.H)
    s := uint32(chs.S)
    return (c*hpc+h)*spt + (s - 1)
}

func decodeCHS(rawchs [3]byte) CHS {
    var res CHS
    res.H = uint16(rawchs[0])
    res.S = uint16(rawchs[1] & 0x3F)
    res.C = uint16(rawchs[2]) | (uint16(rawchs[1]&0xC0) << 2)
    return res
}
