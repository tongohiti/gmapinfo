package img

import "fmt"

type Date struct {
    Month, Year int
}

type Timestamp struct {
    Year, Month, Day int
    HH, MM, SS       int
}

func (d Date) String() string {
    return fmt.Sprintf("%04d/%02d", d.Year, d.Month)
}

func (t Timestamp) String() string {
    return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year, t.Month, t.Day, t.HH, t.MM, t.SS)
}

func convertDate(rawyear, rawmonth uint8) Date {
    var d Date
    d.Month = int(rawmonth) + 1
    if rawyear >= 0x63 {
        d.Year = int(rawyear) + 1900
    } else {
        d.Year = int(rawyear) + 2000
    }
    return d
}

func convertTimestamp(rawyear uint16, rawmonth, rawday uint8, rawhour, rawminute, rawsecond uint8) Timestamp {
    var t Timestamp
    t.Year = int(rawyear)
    t.Month = int(rawmonth) + 1
    t.Day = int(rawday)
    t.HH = int(rawhour)
    t.MM = int(rawminute)
    t.SS = int(rawsecond)
    return t
}
