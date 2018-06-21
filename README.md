gmapinfo
========

This is a command line utility to display information about Garmin maps stored in `.img` format.


Usage
-----

`````
gmapinfo [flags] <img-file> [<output-file>]
  -f    overwrite existing files if necessary
  -s    show subfiles details
  -t    show more technical details
  -x    extract subfiles
  -z    pack extracted subfiles to zip file
`````

Some examples.

Just display basic map information:
`````
C:\>gmapinfo gmapbmap.img

Image file:   gmapbmap.img
Map name:     Worldwide Autoroute DEM Basemap,NR
Map version:  5.1
Map date:     2011/03
Timestamp:    2011-03-14 15:43:24
`````

Display basic information + some technical details (usually not interesting to an average user):
`````
C:\>gmapinfo -t gmapbmap.img

Image file:   gmapbmap.img
Map name:     Worldwide Autoroute DEM Basemap,NR
Map version:  5.1
Map date:     2011/03
Timestamp:    2011-03-14 15:43:24

Block size:      512
Blocks/cluster:  8
Cluster size:    4096
Num clusters:    16384
File table offs: 0x1000 (8 blocks)
Partition 0:     67108864 bytes, 131072 blocks, 16384 clusters
Image file size: 51003392 bytes, 99616 blocks, 12452 clusters
Header size:     36864 bytes, 72 blocks, 9 clusters
Data start:      0x9000
Num entries:     64
`````

Display information about subfiles included in map image file:
`````
C:\>gmapinfo -s gmapbmap.img

Image file:   gmapbmap.img
Map name:     Worldwide Autoroute DEM Basemap,NR
Map version:  5.1
Map date:     2011/03
Timestamp:    2011-03-14 15:43:24

Name          Size, bytes  Date/time            Locked?  Map ID
------------  -----------  -------------------  -------  --------
WX_AMR.RGN         124406  2010-04-28 13:23:58
WX_AMR.TRE           1148  2010-04-28 13:23:58           0x7DE0D7
WX_AMR.LBL         105272  2010-04-28 13:23:58
DCW_DEMT.RGN        71849  2010-04-30 09:19:08
DCW_DEMT.TRE         2260  2010-04-30 09:19:08           0x7DE0DB
DCW_DEMT.LBL         2657  2010-04-30 09:19:08
DCW_DEMT.DEM        19932  2010-04-30 09:19:08
F005701V.RGN     11829689  2011-03-14 14:56:22
F005701V.TRE       109562  2011-03-14 14:56:22           0x7F32C0
F005701V.LBL      1855268  2011-03-14 14:56:22
F005701V.NET      1247732  2011-03-14 14:56:22
F005701V.NOD      4441668  2011-03-14 14:56:22
F005701V.SRT          921  2011-03-14 14:56:22
F005701V.DEM     31126419  2011-03-14 14:56:22

Total 14 subfiles.
`````

Extract subfiles:
`````
C:\>gmapinfo -x gmapbmap.img C:\Temp\map-files

Image file:   gmapbmap.img
Map name:     Worldwide Autoroute DEM Basemap,NR
Map version:  5.1
Map date:     2011/03
Timestamp:    2011-03-14 15:43:24

Writing WX_AMR.RGN .. OK!
Writing WX_AMR.TRE .. OK!
Writing WX_AMR.LBL .. OK!
Writing DCW_DEMT.RGN .. OK!
Writing DCW_DEMT.TRE .. OK!
Writing DCW_DEMT.LBL .. OK!
Writing DCW_DEMT.DEM .. OK!
Writing F005701V.RGN .. OK!
Writing F005701V.TRE .. OK!
Writing F005701V.LBL .. OK!
Writing F005701V.NET .. OK!
Writing F005701V.NOD .. OK!
Writing F005701V.SRT .. OK!
Writing F005701V.DEM .. OK!
Done.
`````

Extract subfiles and place them in a `.zip` archive:
`````
C:\>gmapinfo -x -z gmapbmap.img C:\Temp\map-files.zip

Image file:   gmapbmap.img
Map name:     Worldwide Autoroute DEM Basemap,NR
Map version:  5.1
Map date:     2011/03
Timestamp:    2011-03-14 15:43:24

Writing WX_AMR.RGN .. OK!
Writing WX_AMR.TRE .. OK!
Writing WX_AMR.LBL .. OK!
Writing DCW_DEMT.RGN .. OK!
Writing DCW_DEMT.TRE .. OK!
Writing DCW_DEMT.LBL .. OK!
Writing DCW_DEMT.DEM .. OK!
Writing F005701V.RGN .. OK!
Writing F005701V.TRE .. OK!
Writing F005701V.LBL .. OK!
Writing F005701V.NET .. OK!
Writing F005701V.NOD .. OK!
Writing F005701V.SRT .. OK!
Writing F005701V.DEM .. OK!
Done.
`````


Compiling
---------

Use `go build gmapinfo.go`.


Disclaimer
----------

This software is provided "as is", without warranty of any kind. For more information, see [LICENSE](LICENSE).