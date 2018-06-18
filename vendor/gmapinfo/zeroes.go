package gmapinfo

import "disk"

func readZeroes(imgfile disk.BlockReader, fileTableBlock uint32) (ok bool, err error) {
    for i := int64(1); i < int64(fileTableBlock); i++ {
        zeroes, err := imgfile.ReadBlock(i)
        if err != nil {
            return false, err
        }
        for _, z := range zeroes {
            if z != 0 {
                return false, nil
            }
        }
    }
    return true, nil
}
