package main

import (
    "fmt"
    "os"
    "encoding/hex"
    "disk"
    "img"
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

    hdrblock, err := imgfile.ReadBlock(0)
    if err != nil {
        panic(err)
    }

    fmt.Println(hex.Dump(hdrblock[:]))
    hdr, err := img.DecodeHeader(hdrblock[:])
    if err != nil {
        panic(err)
    }
    fmt.Printf("header = %v\n", hdr)
}
