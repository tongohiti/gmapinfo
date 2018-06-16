package disk

import (
    "fmt"
    "io"
    "os"
)

type ImageFile struct {
    file    *os.File
    blocks  int64
    xorbyte byte
}

func (f *ImageFile) Close() {
    f.file.Close()
    f.file = nil
    f.blocks = 0
}

func (f *ImageFile) ReadBlock(index int64) (*Block, error) {
    _, e := f.file.Seek(index*BlockSize, io.SeekStart)
    if e != nil {
        return nil, e
    }

    var block Block
    r, e := io.ReadFull(f.file, block[:])
    if e != nil {
        return nil, e
    }
    if r != BlockSize {
        return nil, fmt.Errorf("short read: %d/%d", r, BlockSize)
    }

    if index == 0 {
        f.xorbyte = block[0]
    }

    if f.xorbyte != 0 {
        xor(block[:], f.xorbyte)
    }

    return &block, nil
}

func (f *ImageFile) ReadBlocks(index, count int64) ([]byte, error) {
    _, e := f.file.Seek(index*BlockSize, io.SeekStart)
    if e != nil {
        return nil, e
    }

    size := count * BlockSize
    data := make([]byte, size)

    r, e := io.ReadFull(f.file, data)
    if e != nil {
        return nil, e
    }
    if int64(r) != size {
        return nil, fmt.Errorf("short read: %d/%d", r, size)
    }

    if index == 0 {
        f.xorbyte = data[0]
    }

    if f.xorbyte != 0 {
        xor(data, f.xorbyte)
    }

    return data, nil
}

func xor(data []byte, xorbyte byte) {
    for i := range data {
        data[i] ^= xorbyte
    }
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
    if (size % BlockSize) != 0 {
        f.Close()
        return nil, fmt.Errorf("invalid image file size (%d, expected to be a multiple of %d)", size, BlockSize)
    }
    nblocks := size / BlockSize

    return &ImageFile{f, nblocks, 0}, nil
}

func (f *ImageFile) SizeBytes() int64 {
    return f.blocks * BlockSize
}
