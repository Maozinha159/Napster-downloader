package main

import "encoding/xml"

type myTransport struct{}

type WriteCounter struct {
	Total      uint64
	Downloaded uint64
	Percentage int
}

type Config struct {
	Urls          []string
	Format        int
	OutPath       string
	TrackTemplate string
}

type Args struct {
	Urls    []string `arg:"positional, required"`
	Format  int      `arg:"-f" default:"-1"`
	OutPath string   `arg:"-o"`
}

type AlbumMetadata struct {
	XMLName       xml.Name `xml:"AlbumMetadata"`
	Text          string   `xml:",chardata"`
	Copyright     string   `xml:"copyright"`
	DisplayName   string   `xml:"displayName"`
	PrimaryArtist struct {
		Text       string `xml:",chardata"`
		Shortcut   string `xml:"shortcut"`
		Name       string `xml:"name"`
		ArtistId   string `xml:"artistId"`
		RightFlags string `xml:"rightFlags"`
	} `xml:"primaryArtist"`
	AlbumId                  string         `xml:"albumId"`
	PrimaryArtistDisplayName string         `xml:"primaryArtistDisplayName"`
	DisplayableAlbumTypes    string         `xml:"displayableAlbumTypes"`
	AlbumArt70x70Url         string         `xml:"albumArt70x70Url"`
	Shortcut                 string         `xml:"shortcut"`
	NumberOfDiscs            string         `xml:"numberOfDiscs"`
	AlbumType                string         `xml:"albumType"`
	ReleaseYear              string         `xml:"releaseYear"`
	PrimaryArtistId          string         `xml:"primaryArtistId"`
	PrimaryStyle             string         `xml:"primaryStyle"`
	ReleaseDate              string         `xml:"releaseDate"`
	TrackMetas               TrackMetadatas `xml:"trackMetadatas"`
	Label                    string         `xml:"label"`
	OriginalReleaseDate      string         `xml:"originalReleaseDate"`
	Upccode                  string         `xml:"upccode"`
	RightFlags               string         `xml:"rightFlags"`
	AlbumArt162x162Url       string         `xml:"albumArt162x162Url"`
	NonDisplayableAlbumTypes string         `xml:"nonDisplayableAlbumTypes"`
	Pline                    string         `xml:"pline"`
	Name                     string         `xml:"name"`
}

type TrackMetadatas struct {
	Text              string              `xml:",chardata"`
	LiteTrackMetadata []LiteTrackMetadata `xml:"LiteTrackMetadata"`
}

type LiteTrackMetadata struct {
	Text                   string                 `xml:",chardata"`
	GenreId                string                 `xml:"genreId"`
	PlaybackSeconds        string                 `xml:"playbackSeconds"`
	PreviewURL             string                 `xml:"previewURL"`
	DisplayArtistName      string                 `xml:"displayArtistName"`
	LiteTrackPlaybackInfos LiteTrackPlaybackInfos `xml:"liteTrackPlaybackInfos"`
	TrackId                string                 `xml:"trackId"`
	AlbumId                string                 `xml:"albumId"`
	ArtistId               string                 `xml:"artistId"`
	RightFlags             string                 `xml:"rightFlags"`
	TrackIndex             string                 `xml:"trackIndex"`
	DisplayAlbumName       string                 `xml:"displayAlbumName"`
	DiscIndex              string                 `xml:"discIndex"`
	Name                   string                 `xml:"name"`
}

type LiteTrackPlaybackInfos struct {
	Text                  string                  `xml:",chardata"`
	LiteTrackPlaybackInfo []LiteTrackPlaybackInfo `xml:"LiteTrackPlaybackInfo"`
}

type LiteTrackPlaybackInfo struct {
	Text                string              `xml:",chardata"`
	MediaUrl            string              `xml:"mediaUrl"`
	TrackPlaybackFormat TrackPlaybackFormat `xml:"trackPlaybackFormat"`
}

type TrackPlaybackFormat struct {
	Text    string `xml:",chardata"`
	BitRate int    `xml:"bitRate"`
	Format  string `xml:"format"`
}

type Resolve struct {
	ID         string `json:"id"`
	ResultCode int    `json:"resultCode"`
	Tokens     struct {
		Artist string `json:"artist"`
		Album  string `json:"album"`
	} `json:"tokens"`
}
