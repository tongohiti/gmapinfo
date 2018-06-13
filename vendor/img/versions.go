package img

import "fmt"

type Version struct {
    Maj, Min uint8
}

func (v Version) String() string {
    return fmt.Sprintf("%d.%d", v.Maj, v.Min)
}
