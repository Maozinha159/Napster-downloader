# Napster-Downloader
Napster downloader written in Go.
![](https://i.imgur.com/FPInZVO.png)
[Windows binaries](https://github.com/Sorrow446/Napster-downloader/releases)

# Setup
Account not needed.    
Configure config file if needed.
|Option|Info|
| --- | --- |
|format|Download quality. 1 = 96 AAC PLUS, 2 = 192 AAC, 3 = best/320 AAC
|outPath|Where to download to. Path will be made if it doesn't already exist.
|trackTemplate|Track filename template. Vars: album, albumArtist, artist, copyright, genre, title, trackNum, trackNumPad, trackTotal, year

# Usage
Args take priority over the config file.

Download two albums:   
`np_dl_x64.exe https://app.napster.com/artist/deadmau5/album/tau-v1-v2-single https://app.napster.com/artist/deadmau5/album/random-album-title-awal-records`

Download a single album and from two text files:   
`np-dl_x64.exe.exe https://app.napster.com/artist/deadmau5/album/tau-v1-v2-single G:\1.txt G:\2.txt`

```
 _____             _              ____                _           _
|   | |___ ___ ___| |_ ___ ___   |    \ ___ _ _ _ ___| |___ ___ _| |___ ___
| | | | .'| . |_ -|  _| -_|  _|  |  |  | . | | | |   | | . | .'| . | -_|  _|
|_|___|__,|  _|___|_| |___|_|    |____/|___|_____|_|_|_|___|__,|___|___|_|
          |_|                                                                                                           

Usage: np_dl_x64.exe [--format FORMAT] [--outpath OUTPATH] URLS [URLS ...]

Positional arguments:
  URLS

Options:
  --format FORMAT, -f FORMAT [default: -1]
  --outpath OUTPATH, -o OUTPATH
  --help, -h             display this help and exit
```
