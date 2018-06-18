package gmapinfo

import (
    "io"
    "os"
    "path/filepath"
)

// Interface to create writable files in some location (in a directory or in zip archive, for example).
// Standard zip.Writer conforms to this interface.
type FileWriter interface {
    Create(name string) (io.Writer, error)
    Close() error
}

func newDirectoryWriter(path string) FileWriter {
    dw := DirectoryWriter{path: path}
    return &dw
}

type DirectoryWriter struct {
    path     string
    lastFile *os.File
}

func (dw *DirectoryWriter) Create(name string) (io.Writer, error) {
    e := dw.Close() // close previous file, if any
    if e != nil {
        return nil, e
    }

    fullname := filepath.Join(dw.path, name)
    dirname := filepath.Dir(fullname)

    if dirname != "." {
        e := os.MkdirAll(dirname, 777)
        if e != nil {
            return nil, e
        }
    }

    f, e := os.Create(fullname)
    if e != nil {
        return nil, e
    }
    dw.lastFile = f
    return f, nil
}

func (dw *DirectoryWriter) Close() error {
    if dw.lastFile != nil {
        e := dw.lastFile.Close()
        dw.lastFile = nil
        if e != nil {
            return e
        }
    }
    return nil
}
