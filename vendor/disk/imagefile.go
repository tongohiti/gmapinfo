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

func (f *ImageFile) ReadBlock(n int64) (*Block, error) {
    _, e := f.file.Seek(n*BlockSize, io.SeekStart)
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

    if n == 0 {
        f.xorbyte = block[0]
    }

    if f.xorbyte != 0 {
        xor(&block, f.xorbyte)
    }

    return &block, nil
}

func xor(block *Block, xorbyte byte) {
    for i := range block {
        block[i] ^= xorbyte
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
