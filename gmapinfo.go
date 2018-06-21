package main

import (
    "fmt"
    "flag"
    "os"
    "path/filepath"
    "gmapinfo"
)

func main() {
    fmt.Println("gmapinfo 1.0.0")

    var params gmapinfo.Params
    flag.BoolVar(&params.ShowDetails, "t", false, "show more technical details")
    flag.BoolVar(&params.ShowSubfiles, "s", false, "show subfiles details")
    flag.BoolVar(&params.Extract, "x", false, "extract subfiles")
    flag.BoolVar(&params.ZipOutput, "z", false, "pack extracted subfiles to zip file")
    flag.BoolVar(&params.ForceOverwrite, "f", false, "overwrite existing files if necessary")
    flag.Usage = usage
    flag.Parse()
    argc := len(flag.Args())
    ok := (argc == 1 && !params.Extract) || (argc == 2 && params.Extract)
    if !ok {
        os.Stdout.Sync()
        fmt.Fprintln(os.Stderr, "Bad arguments")
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

func usage() {
    name := filepath.Base(os.Args[0])
    fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags] <img-file> [<output-file>]\n", name)
    flag.PrintDefaults()
}
