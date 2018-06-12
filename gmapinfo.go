package main

import (
    "fmt"
    "os"
    "encoding/hex"
    "disk"
)

func main() {
    fmt.Println("gmapinfo 0.0.1")

    imagefile := os.Args[1]
    fmt.Printf("Image file: %s\n", imagefile)

    imgfile, err := disk.OpenImageFile(imagefile)
    if err != nil {
        panic(err)
    }
    defer imgfile.Close()

    bootsector, err := loadBootSector(imgfile)
    if err != nil {
        panic(err)
    }
    fmt.Println(hex.Dump(bootsector[:]))
}

func loadBootSector(dev disk.BlockReader) (*disk.Block, error) {
    bootsector, err := dev.ReadBlock(0)
    if err != nil {
        return nil, err
    }

    var bootsignature uint16
    bootsignature = uint16(bootsector[510]) | (uint16(bootsector[511]) << 8)
    if bootsignature != 0xAA55 {
        return nil, fmt.Errorf("invalid boot sector (boot signature = 0x%04X)", bootsignature)
    }

    return bootsector, nil
}
