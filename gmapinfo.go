package main

import (
    "fmt"
    "os"
    "disk"
    "img"
)

func main() {
    fmt.Println("gmapinfo 0.0.1")

    imagefile := os.Args[1]
    fmt.Printf("Image file:  %s\n", imagefile)

    imgfile, err := disk.OpenImageFile(imagefile)
    if err != nil {
        panic(err)
    }
    defer imgfile.Close()

    hdrblock, err := imgfile.ReadBlock(0)
    if err != nil {
        panic(err)
    }

    hdr, err := img.DecodeHeader(hdrblock[:])
    if err != nil {
        panic(err)
    }

    fmt.Printf("Map name:    %s\n", hdr.MapName)
    fmt.Printf("Map version: %v\n", hdr.MapVersion)
    fmt.Printf("Map date:    %v\n", hdr.MapDate)
    fmt.Printf("Timestamp:   %v\n", hdr.CreateDate)
}
