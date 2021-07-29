package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"text/template"

	ap "github.com/Sorrow446/go-atomicparsley"
	arg "github.com/alexflint/go-arg"
	"github.com/dustin/go-humanize"
)

var client = &http.Client{Transport: &myTransport{}}

func (t *myTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(
		"User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	)
	req.Header.Add(
		"Referer", "https://app.napster.com/",
	)
	return http.DefaultTransport.RoundTrip(req)
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Downloaded += uint64(n)
	percentage := float64(wc.Downloaded) / float64(wc.Total) * float64(100)
	wc.Percentage = int(percentage)
	fmt.Printf("\r%d%%, %s/%s", wc.Percentage, humanize.Bytes(wc.Downloaded), humanize.Bytes(wc.Total))
	return n, nil
}

func getScriptDir() (string, error) {
	var (
		ok    bool
		err   error
		fname string
	)
	if filepath.IsAbs(os.Args[0]) {
		_, fname, _, ok = runtime.Caller(0)
		if !ok {
			return "", errors.New("Failed to get script filename.")
		}
	} else {
		fname, err = os.Executable()
		if err != nil {
			return "", err
		}
	}
	scriptDir := filepath.Dir(fname)
	return scriptDir, nil
}

func readTxtFile(path string) ([]string, error) {
	var lines []string
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return lines, nil
}

func contains(lines []string, value string) bool {
	for _, line := range lines {
		if strings.EqualFold(line, value) {
			return true
		}
	}
	return false
}

func processUrls(urls []string) ([]string, error) {
	var processed []string
	var txtPaths []string
	for _, url := range urls {
		if strings.HasSuffix(url, ".txt") && !contains(txtPaths, url) {
			txtLines, err := readTxtFile(url)
			if err != nil {
				return nil, err
			}
			for _, txtLine := range txtLines {
				if !contains(processed, txtLine) {
					processed = append(processed, txtLine)
				}
			}
			txtPaths = append(txtPaths, url)
		} else {
			if !contains(processed, url) {
				processed = append(processed, url)
			}
		}
	}
	return processed, nil
}

func readConfig() (*Config, error) {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var obj Config
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func parseArgs() *Args {
	var args Args
	arg.MustParse(&args)
	return &args
}

func parseCfg() (*Config, error) {
	cfg, err := readConfig()
	if err != nil {
		return nil, err
	}
	args := parseArgs()
	if args.Format != -1 {
		cfg.Format = args.Format
	}
	if !(cfg.Format >= 1 && cfg.Format <= 3) {
		return nil, errors.New("Format must be between 1 and 3.")
	}
	if args.OutPath != "" {
		cfg.OutPath = args.OutPath
	}
	if cfg.OutPath == "" {
		cfg.OutPath = "Napster downloads"
	}
	cfg.Urls, err = processUrls(args.Urls)
	if err != nil {
		errString := fmt.Sprintf("Failed to process URLs.%s", err)
		return nil, errors.New(errString)
	}
	return cfg, nil
}

func getMetadata(albumId string) (*AlbumMetadata, error) {
	_url := "http://direct-ns.rhapsody.com/metadata/data/methods/getAlbum.xml"
	req, err := http.NewRequest(http.MethodGet, _url, nil)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("developerKey", "8I1E4E1C2G5F1I8F")
	query.Set("cobrandId", "60301:105")
	query.Set("filterRightsKey", "2")
	query.Set("albumId", albumId)
	req.URL.RawQuery = query.Encode()
	do, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, errors.New(do.Status)
	}
	var obj AlbumMetadata
	err = xml.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func resolveShortcuts(artistShortcut, albumShortcut string) (string, error) {
	_url := "https://direct.rhapsody.com/metadata/data/methods/getIdByShortcut.js"
	req, err := http.NewRequest(http.MethodGet, _url, nil)
	if err != nil {
		return "", err
	}
	query := url.Values{}
	query.Set("albumShortcut", albumShortcut)
	query.Set("artistShortcut", artistShortcut)
	query.Set("developerKey", "5C8F8G9G8B4D0E5J")
	req.URL.RawQuery = query.Encode()
	do, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", errors.New(do.Status)
	}
	var obj Resolve
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return "", err
	}
	return obj.ID, nil
}

func checkUrl(url string) (string, string) {
	regexString := `^https://app.napster.com/artist/([a-z\d-]+)/album/([a-z\d-]+)$`
	regex := regexp.MustCompile(regexString)
	matches := regex.FindAllStringSubmatch(url, -1)
	if matches != nil {
		return matches[0][1], matches[0][2]
	}
	return "", ""
}

func makeDir(path string) error {
	err := os.Mkdir(path, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	return nil
}

func fileExists(path string) (bool, error) {
	f, err := os.Stat(path)
	if err == nil {
		return !f.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func init() {
	fmt.Println(`                                                                  
 _____             _              ____                _           _         
|   | |___ ___ ___| |_ ___ ___   |    \ ___ _ _ _ ___| |___ ___ _| |___ ___ 
| | | | .'| . |_ -|  _| -_|  _|  |  |  | . | | | |   | | . | .'| . | -_|  _|
|_|___|__,|  _|___|_| |___|_|    |____/|___|_____|_|_|_|___|__,|___|___|_|  
		  |_| 
	`)
	scriptDir, err := getScriptDir()
	if err != nil {
		panic(err)
	}
	err = os.Chdir(scriptDir)
	if err != nil {
		panic(err)
	}
}

func sanitize(filename string) string {
	regex := regexp.MustCompile(`[\/:*?"><|]`)
	sanitized := regex.ReplaceAllString(filename, "_")
	return sanitized
}

func parseFileMeta(meta *LiteTrackPlaybackInfos, format int) (string, string) {
	resolve := map[int]int{
		1: 64,
		2: 192,
	}
	chosenIndex := 0
	if format == 3 {
		sort.Slice(meta.LiteTrackPlaybackInfo, func(x, y int) bool {
			return meta.LiteTrackPlaybackInfo[x].TrackPlaybackFormat.BitRate >
				meta.LiteTrackPlaybackInfo[y].TrackPlaybackFormat.BitRate
		})
	} else {
		resolvedFormat := resolve[format]
		for i, track := range meta.LiteTrackPlaybackInfo {
			if track.TrackPlaybackFormat.BitRate == resolvedFormat {
				chosenIndex = i
				break
			}
		}
	}
	specs := fmt.Sprintf("%d Kbps %s",
		meta.LiteTrackPlaybackInfo[chosenIndex].TrackPlaybackFormat.BitRate,
		meta.LiteTrackPlaybackInfo[chosenIndex].TrackPlaybackFormat.Format,
	)
	originalUrl := meta.LiteTrackPlaybackInfo[chosenIndex].MediaUrl
	url := "https://napster-mmd.lldns.net/" + filepath.Base(originalUrl)
	return url, specs
}

func downloadCover(albumId, coverPath string) error {
	url := "https://direct.rhapsody.com/imageserver/images/" + albumId + "/600x600.jpg"
	f, err := os.OpenFile(coverPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := client.Get(url)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return errors.New(req.Status)
	}
	_, err = io.Copy(f, req.Body)
	return err
}

func parseAlbumMeta(meta *AlbumMetadata) map[string]string {
	parsedMeta := map[string]string{
		"album":       meta.Name,
		"albumArtist": meta.PrimaryArtist.Name,
		"copyright":   meta.Copyright,
		"genre":       meta.PrimaryStyle,
		"year":        meta.ReleaseYear,
	}
	return parsedMeta
}

func parseTrackMeta(trackMeta *LiteTrackMetadata, albMeta map[string]string, trackNum, trackTotal int) map[string]string {
	albMeta["artist"] = trackMeta.DisplayArtistName
	albMeta["title"] = trackMeta.Name
	albMeta["tracknum"] = fmt.Sprintf("%d/%d", trackNum, trackTotal)
	albMeta["trackNum"] = strconv.Itoa(trackNum)
	albMeta["trackNumPad"] = fmt.Sprintf("%02d", trackNum)
	albMeta["trackTotal"] = strconv.Itoa(trackNum)
	return albMeta
}

func parseTemplate(templateText string, tags map[string]string) string {
	var buffer bytes.Buffer
	for {
		err := template.Must(template.New("").Parse(templateText)).Execute(&buffer, tags)
		if err == nil {
			break
		}
		fmt.Println("Failed to parse template. Default will be used instead.")
		templateText = "{{ .trackNumPad}}. {{ .title }}"
		buffer.Reset()
	}
	return buffer.String()
}

func downloadTrack(trackPath, url string) error {
	f, err := os.Create(trackPath)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Range", "bytes=0-")
	do, err := client.Do(req)
	if err != nil {
		return err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK && do.StatusCode != http.StatusPartialContent {
		return errors.New(do.Status)
	}
	counter := &WriteCounter{Total: uint64(do.ContentLength)}
	_, err = io.Copy(f, io.TeeReader(do.Body, counter))
	fmt.Println("")
	return err
}

func writeTags(trackPath, coverPath string, tags map[string]string) error {
	if coverPath != "" {
		tags["artwork"] = coverPath
	}
	err := ap.WriteTags(trackPath, tags)
	return err
}

func main() {
	var coverPath string
	cfg, err := parseCfg()
	if err != nil {
		errString := fmt.Sprintf("Failed to parse config file. %s", err)
		panic(errString)
	}
	err = makeDir(cfg.OutPath)
	if err != nil {
		errString := fmt.Sprintf("Failed to make output folder. %s", err)
		panic(errString)
	}
	albumTotal := len(cfg.Urls)
	for albumNum, url := range cfg.Urls {
		fmt.Printf("Album %d of %d:\n", albumNum+1, albumTotal)
		artist, album := checkUrl(url)
		if artist == "" {
			fmt.Println("Invalid URL:", url)
			return
		}
		albumId, err := resolveShortcuts(artist, album)
		if err != nil {
			fmt.Println("Failed to resolve shortcuts.\n", err)
			return
		}
		meta, err := getMetadata(albumId)
		if err != nil {
			fmt.Println("Failed to get metadata.\n", err)
			continue
		}
		parsedAlbMeta := parseAlbumMeta(meta)
		albFolder := parsedAlbMeta["albumArtist"] + " - " + parsedAlbMeta["album"]
		fmt.Println(albFolder)
		albumPath := filepath.Join(cfg.OutPath, sanitize(albFolder))
		err = makeDir(albumPath)
		if err != nil {
			fmt.Println("Failed to make album folder.\n", err)
			continue
		}
		coverPath = filepath.Join(albumPath, "cover.jpg")
		err = downloadCover(albumId, coverPath)
		if err != nil {
			fmt.Println("Failed to get cover.\n", err)
			coverPath = ""
		}
		trackTotal := len(meta.TrackMetas.LiteTrackMetadata)
		for trackNum, track := range meta.TrackMetas.LiteTrackMetadata {
			trackNum++
			parsedMeta := parseTrackMeta(&track, parsedAlbMeta, trackNum, trackTotal)
			trackFname := parseTemplate(cfg.TrackTemplate, parsedMeta)
			trackPath := filepath.Join(albumPath, sanitize(trackFname)+".m4a")
			exists, err := fileExists(trackPath)
			if err != nil {
				fmt.Println("Failed to check if track already exists locally.\n", err)
				continue
			}
			if exists {
				fmt.Println("Track already exists locally.")
				continue
			}
			url, specs := parseFileMeta(&track.LiteTrackPlaybackInfos, cfg.Format)
			fmt.Printf("Downloading track %d of %d: %s - %s\n", trackNum, trackTotal, parsedMeta["title"], specs)
			err = downloadTrack(trackPath, url)
			if err != nil {
				fmt.Println("Failed to download track.\n", err)
				continue
			}
			err = writeTags(trackPath, coverPath, parsedMeta)
			if err != nil {
				fmt.Println("Failed to write tags.\n", err)
			}
		}
	}
}
