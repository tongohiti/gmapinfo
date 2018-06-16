package gmapinfo

import (
    "disk"
    "img"
    "fmt"
    "errors"
    "os"
    "archive/zip"
)

var ErrExtract = errors.New("extract error")

func extractFiles(imgfile disk.BlockReader, clusterblocks uint32, files []img.FileEntry, outname string, zipout bool) error {
    if !zipout {
        return ErrExtract
    }

    f, e := os.Create(outname)
    if e != nil {
        return e
    }
    defer f.Close()

    z := zip.NewWriter(f)
    defer z.Close()

    clustersize := int64(clusterblocks) * disk.BlockSize
    for i := range files {
        entry := &files[i]
        name := entry.Name

        ze, err := z.Create(name)
        if err != nil {
            return err
        }

        size := int64(entry.Size)
        nclusters := len(entry.FAT)
        for i, cluster := range entry.FAT {
            progress := (i + 1) * 100 / nclusters
            fmt.Printf("Writing %s .. %d%%\r", name, progress)
            os.Stdout.Sync()
            data, err := imgfile.ReadBlocks(int64(cluster)*int64(clusterblocks), int64(clusterblocks))
            if err != nil {
                return err
            }
            if size < clustersize {
                data = data[:size]
            }
            size -= int64(len(data))
            ze.Write(data)
        }

        fmt.Printf("Writing %s .. OK! \n", name)
    }

    return nil
}
