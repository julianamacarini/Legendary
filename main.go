package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

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

type Subtitle struct {
	Id         string `json:"id"`
	Type       string `json:"type"`
	Attributes struct {
		SubtitleId        string `json:"subtitle_id"`
		Language          string `json:"language"`
		DownloadCount     int    `json:"download_count"`
		NewDownloadCount  int    `json:"new_download_count"`
		HearingImpaired   bool   `json:"hearing_impaired"`
		Hd                bool   `json:"hd"`
		Fps               int    `json:"fps"`
		Votes             int    `json:"votes"`
		Points            int    `json:"points"`
		Ratings           int    `json:"ratings"`
		FromTrusted       bool   `json:"from_trusted"`
		ForeignPartsOnly  bool   `json:"foreign_parts_only"`
		AiTranslated      bool   `json:"ai_translated"`
		MachineTranslated bool   `json:"machine_translated"`
		UploadDate        string `json:"upload_date"`
		Release           string `json:"release"`
		Comments          string `json:"comments"`
		LegacySubtitleId  int    `json:"legacy_subtitle_id"`
		Uploader          Uploader
		FeatureDetails    FeatureDetails
		Url               string     `json:"url"`
		RelatedLinks      []struct{} `json:"related_links"`
		Files             []struct{} `json:"files"`
	} `json:"attributes"`
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

func main() {
	filename := os.Args[1]

	apiKeyBytes, apiKeyErr := os.ReadFile("api.key")

	if apiKeyErr != nil {
		println(apiKeyErr)
		os.Exit(1)
	}

	apiKey := strings.TrimSpace(string(apiKeyBytes))

	println(filename, apiKey)

	client := &http.Client{}

	req, reqErr := http.NewRequest("GET", "https://api.opensubtitles.com/api/v1/utilities/guessit?filename="+filename, nil)

	if reqErr != nil {
		println(reqErr)
		os.Exit(1)
	}

	req.Header.Add("Api-Key", apiKey)

	res, resErr := client.Do(req)

	if resErr != nil {
		println(resErr)
		os.Exit(1)
	}

	// parse json FileInfo response
	fileInfo := &FileInfo{}
	parseErr := json.Unmarshal([]byte(res.Body), fileInfo)

	if parseErr != nil {
		println(parseErr)
		os.Exit(1)
	}

	// TODO: Query file info.
	// GET https://api.opensubtitles.com/api/v1/utilities/guessit?filename=Game.of.Thrones.S03E12.Bluray.1080p.x264.mkv

	// GET https://api.opensubtitles.com/api/v1/subtitles?imdb_id=tt0944947&season=3&episode=12&languages=en
}
