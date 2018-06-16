package main

import (
    "fmt"
    "flag"
    "os"
    "gmapinfo"
)

func main() {
    fmt.Println("gmapinfo 0.0.3")

    var params gmapinfo.Params
    flag.BoolVar(&params.Extract, "x", false, "extract subfiles")
    flag.BoolVar(&params.ZipOutput, "z", false, "pack extracted subfiles to zip file")
    flag.Parse()
    argc := len(flag.Args())
    ok := (argc == 1 && !params.Extract) || (argc == 2 && params.Extract)
    if !ok {
        os.Stdout.Sync()
        fmt.Fprintln(os.Stderr, "Bad arguments:", flag.Args())
        flag.Usage()
        os.Exit(2)
    }
    params.FileName = flag.Arg(0)
    if params.Extract {
        params.OutputName = flag.Arg(1)
    }

    err := gmapinfo.Run(params)
    if err != nil {
        os.Stdout.Sync()
        fmt.Fprintln(os.Stderr, "Error:", err.Error())
        os.Exit(1)
    }
}
