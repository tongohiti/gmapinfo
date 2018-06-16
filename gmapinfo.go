package main

import (
    "fmt"
    "os"
    "gmapinfo"
)

func main() {
    fmt.Println("gmapinfo 0.0.3")

    var params gmapinfo.Params
    params.FileName = os.Args[1]

    err := gmapinfo.Run(params)
    if err != nil {
        os.Stdout.Sync()
        fmt.Fprintln(os.Stderr, "Error:", err.Error())
        os.Exit(1)
    }
}
