package main

import (
    "fmt"
    "os"
    "encoding/hex"
)

func main() {
    fmt.Println("gmapinfo 0.0.1")
    imagefile := os.Args[1]
    fmt.Printf("Image file: %s\n", imagefile)
    bootsector, err := loadBootSector(imagefile)
    if err != nil {
        panic(err)
    }
    fmt.Println(hex.Dump(bootsector[:]))
}

func loadBootSector(imagefile string) (*[512]byte, error) {
    f, e := os.Open(imagefile)
    if e != nil {
        return nil, e
    }
    defer f.Close()

    var bootsector [512]byte
    f.Read(bootsector[:])

    var bootsignature uint16
    bootsignature = uint16(bootsector[510]) | (uint16(bootsector[511]) << 8)
    if bootsignature != 0xAA55 {
        return nil, fmt.Errorf("invalid image file (boot signature = 0x%04X)", bootsignature)
    }

    return &bootsector, nil
}
