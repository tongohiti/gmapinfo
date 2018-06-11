package disk

import (
    "fmt"
    "io"
    "os"
)

type ImageFile struct {
    f *os.File
    n int64
}

func (f *ImageFile) Close() {
    f.f.Close()
    f.f = nil
    f.n = 0
}

func (f *ImageFile) ReadSector(n int64) (*Sector, error) {
    _, e := f.f.Seek(n*SectorSize, io.SeekStart)
    if e != nil {
        return nil, e
    }

    var sector Sector
    r, e := io.ReadFull(f.f, sector[:])
    if e != nil {
        return nil, e
    }
    if r != SectorSize {
        return nil, fmt.Errorf("short read: %d/%d", r, SectorSize)
    }

    return &sector, nil
}

func OpenImageFile(filename string) (*ImageFile, error) {
    f, e := os.Open(filename)
    if e != nil {
        return nil, e
    }

    fs, e := f.Stat()
    if e != nil {
        f.Close()
        return nil, e
    }
    size := fs.Size()
    if (size % 512) != 0 {
        f.Close()
        return nil, fmt.Errorf("invalid image file size (%d, expected to be a multiple of 512)", size)
    }
    nsectors := size / 512

    return &ImageFile{f, nsectors}, nil
}
