package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		panic("Usage: legendary <filename>")
	}

	filename := os.Args[1]

	apiKeyBytes, apiKeyErr := os.ReadFile("api.key")

	if apiKeyErr != nil {
		panic(apiKeyErr)
	}

	apiKey := strings.TrimSpace(string(apiKeyBytes))

	f := GetFileInfo(filename, apiKey)
	SearchSubtitles(f, filename, apiKey)
}

func SearchSubtitles(f *FileInfo, filename, apiKey string) {
	query := GetQueryValues(f, filename)
	subtitles := &SearchResponse{}
	DoRequest("https://api.opensubtitles.com/api/v1/subtitles", query, apiKey, subtitles)

	fmt.Printf("Found %d subtitles.\n", subtitles.TotalCount)

	for _, subtitle := range subtitles.Data {
		fmt.Printf("%s\t%d\t%t\n", subtitle.Attributes.Files[0].FileName, subtitle.Attributes.DownloadCount, subtitle.Attributes.HearingImpaired)
	}

	println("------")

	var bestSubtitle *Subtitle

	for _, subtitle := range subtitles.Data {
		if len(subtitle.Attributes.Files) != 1 {
			continue
		}

		file := subtitle.Attributes.Files[0]

		subFileName := strings.ToLower(file.FileName)

		if !strings.Contains(subFileName, f.ScreenSize) {
			continue
		}

		switch f.Source {
		case "Blu-Ray":
			if !strings.Contains(subFileName, "bluray") {
				continue
			}
		case "WEB-DL":
			if !strings.Contains(subFileName, "web-dl") {
				continue
			}
		case "WEBRip":
			if !strings.Contains(subFileName, "webrip") {
				continue
			}
		case "HDTV":
			if !strings.Contains(subFileName, "hdtv") {
				continue
			}
		}

		if bestSubtitle == nil {
			bestSubtitle = &subtitle
			continue
		}

		//  Prefer non-hearing impaired subtitles.
		if bestSubtitle.Attributes.HearingImpaired && !subtitle.Attributes.HearingImpaired {
			bestSubtitle = &subtitle
			continue
		}

		// Prefer subtitles with more downloads.
		if bestSubtitle.Attributes.DownloadCount < subtitle.Attributes.DownloadCount {
			bestSubtitle = &subtitle
			continue
		}
	}

	fmt.Printf("%s %d %t\n", bestSubtitle.Attributes.Files[0].FileName, bestSubtitle.Attributes.DownloadCount, bestSubtitle.Attributes.HearingImpaired)
}

func GetQueryValues(f *FileInfo, filename string) *url.Values {
	if f.Type == "episode" {
		return &url.Values{
			"type":               {f.Type},
			"query":              {filename},
			"season_number":      {strconv.Itoa(f.Season)},
			"episode_number":     {strconv.Itoa(f.Episode)},
			"languages":          {"en"},
			"foreign_parts_only": {"exclude"},
			"trusted_sources":    {"only"},
		}
	}

	if f.Type == "movie" {
		return &url.Values{
			"type":            {f.Type},
			"query":           {filename},
			"languages":       {"en"},
			"trusted_sources": {"only"},
		}
	}

	panic(fmt.Sprintf("Unknown file type: %s\n", f.Type))
}

func GetFileInfo(filename string, apiKey string) *FileInfo {
	fileInfo := &FileInfo{}
	query := url.Values{
		"filename": {filename},
	}
	DoRequest("https://api.opensubtitles.com/api/v1/utilities/guessit", &query, apiKey, fileInfo)

	return fileInfo
}

func DoRequest(url string, query *url.Values, apiKey string, v any) {
	client := &http.Client{}

	u := url + "?" + query.Encode()

	fmt.Printf("Requesting %s\n", strings.ToLower(u))

	req, reqErr := http.NewRequest("GET", u, nil)

	if reqErr != nil {
		panic(reqErr)
	}

	req.Header.Add("Api-Key", apiKey)

	res, resErr := client.Do(req)

	if resErr != nil {
		panic(resErr)
	}

	defer res.Body.Close()

	body, readErr := ioutil.ReadAll(res.Body)

	if readErr != nil && readErr != io.EOF {
		panic(readErr)
	}

	parseErr := json.Unmarshal(body, v)

	if parseErr != nil {
		panic(parseErr)
	}

	// s, e := json.MarshalIndent(v, "", "  ")

	// if e != nil {
	// 	panic(e)
	// }

	// println(string(s))
}

type FileInfo struct {
	Title      string `json:"title"`
	Season     int    `json:"season"`
	Episode    int    `json:"episode"`
	Source     string `json:"source"`
	ScreenSize string `json:"screen_size"`
	VideoCodec string `json:"video_codec"`
	Container  string `json:"container"`
	MimeType   string `json:"mimetype"`
	Type       string `json:"type"`
}

type SearchResponse struct {
	TotalPages int `json:"total_pages"`
	TotalCount int `json:"total_count"`
	PerPage    int `json:"per_page"`
	Page       int `json:"page"`
	Data       []Subtitle
}

type Subtitle struct {
	Id         string `json:"id"`
	Type       string `json:"type"`
	Attributes struct {
		SubtitleId        string  `json:"subtitle_id"`
		Language          string  `json:"language"`
		DownloadCount     int     `json:"download_count"`
		NewDownloadCount  int     `json:"new_download_count"`
		HearingImpaired   bool    `json:"hearing_impaired"`
		Hd                bool    `json:"hd"`
		Fps               float32 `json:"fps"`
		Votes             int     `json:"votes"`
		Points            int     `json:"points"`
		Ratings           float32 `json:"ratings"`
		FromTrusted       bool    `json:"from_trusted"`
		ForeignPartsOnly  bool    `json:"foreign_parts_only"`
		AiTranslated      bool    `json:"ai_translated"`
		MachineTranslated bool    `json:"machine_translated"`
		UploadDate        string  `json:"upload_date"`
		Release           string  `json:"release"`
		Comments          string  `json:"comments"`
		LegacySubtitleId  int     `json:"legacy_subtitle_id"`
		Uploader          Uploader
		FeatureDetails    FeatureDetails
		Url               string        `json:"url"`
		RelatedLinks      []RelatedLink `json:"related_links"`
		Files             []File        `json:"files"`
	} `json:"attributes"`
}

type RelatedLink struct {
	Label  string `json:"label"`
	Url    string `json:"url"`
	ImgUrl string `json:"img_url"`
}

type File struct {
	FileId   int    `json:"file_id"`
	CdNumber int    `json:"cd_number"`
	FileName string `json:"file_name"`
}

type Uploader struct {
	UploaderId int    `json:"uploader_id"`
	Name       string `json:"name"`
	Rank       string `json:"rank"`
}

type FeatureDetails struct {
	FeatureId   int    `json:"feature_id"`
	FeatureType string `json:"feature_type"`
	Year        int    `json:"year"`
	Title       string `json:"title"`
	MovieName   string `json:"movie_name"`
	ImdbId      int    `json:"imdb_id"`
	TmdbId      int    `json:"tmdb_id"`
}
