package gmapinfo

import (
    "disk"
    "img"
    "fmt"
    "strings"
    "errors"
    "io"
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
        isGMP := strings.HasSuffix(name, ".GMP")

        var err error
        if isGMP {
            gmpdirectory, err := img.ReadGmpDirectory(imgfile, entry, clusterblocks)
            if err != nil {
                return err
            }
            for _, e := range gmpdirectory {
                subname := name + "/" + name[:len(name)-4] + "." + e.Format
                err = saveSubfile(subname, int64(e.Offset), int64(e.Length), entry.FAT, clustersize, int64(clusterblocks), imgfile, z)
                if err != nil {
                    break
                }
            }
        } else {
            err = saveSubfile(name, 0, int64(entry.Size), entry.FAT, clustersize, int64(clusterblocks), imgfile, z)
        }
        if err != nil {
            return err
        }
    }

    return nil
}

type progressFunc func(current, total int)

func saveSubfile(name string, offset, size int64, fat []uint16, clustersize, clusterblocks int64, imgfile disk.BlockReader, zw *zip.Writer) error {
    ze, err := zw.Create(name)
    if err != nil {
        return err
    }

    progressFunc := func(current, total int) {
        progress := (current + 1) * 100 / total
        fmt.Printf("Writing %s .. %d%%\r", name, progress)
        os.Stdout.Sync()
    }

    err = saveFileRegion(offset, size, fat, clustersize, clusterblocks, imgfile, ze, progressFunc)
    if err != nil {
        fmt.Printf("Writing %s .. FAILED! \n", name)
        return err
    }

    fmt.Printf("Writing %s .. OK! \n", name)
    return nil
}

func saveFileRegion(offset, size int64, fat []uint16, clustersize, clusterblocks int64, imgfile disk.BlockReader, target io.Writer, progress progressFunc) error {
    startcluster := offset / clustersize
    endcluster := (offset + size) / clustersize
    fatregion := fat[startcluster : endcluster+1]
    nclusters := len(fatregion)
    for i, cluster := range fatregion {
        progress(i, nclusters)
        data, err := imgfile.ReadBlocks(int64(cluster)*int64(clusterblocks), int64(clusterblocks))
        if err != nil {
            return err
        }

        isFirst := i == 0
        isLast := i == (nclusters - 1)
        var from, to int64
        if isFirst {
            from = offset % clustersize
        } else {
            from = 0
        }
        if isLast {
            to = (offset + size) % clustersize
        } else {
            to = clustersize
        }
        fragment := data[from:to]

        n, err := target.Write(fragment)
        if err != nil {
            return err
        }
        if n != len(fragment) {
            return ErrExtract
        }
    }
    return nil
}
