package dlp

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	disgotypes "main/disgo-types"
	"os"
	"os/exec"
	"strings"
	"time"
)

type AudioService struct {
	MaxAudioDurationInSeconds int
	FileDirectory             string
}

func CreateService(maxAudioDuration int) *AudioService {
	err := os.Mkdir("media", os.ModePerm)
	if err != nil {
		log.Println(err)
	}

	return &AudioService{
		MaxAudioDurationInSeconds: maxAudioDuration,
		FileDirectory:             "media",
	}
}

func (as *AudioService) AudioServiceRunner(query string) (*disgotypes.Media, error) {
	timeout := make(chan bool, 1)
	result := make(chan mediaQueryResult, 1)
	log.Println("audio service runner started")

	go func() {
		time.Sleep(60 * time.Second)
		timeout <- true
	}()
	go func() {
		result <- as.queryAndDownload(query)
	}()

	select {
	case <-timeout:
		return nil, errors.New("request timed out")
	case result := <-result:
		log.Println(result)
		return result.Media, result.Error
	}
}

func (as *AudioService) queryAndDownload(query string) mediaQueryResult {
	log.Println("starting query and download")
	start := time.Now()

	ytDownloader, err := exec.LookPath("yt-dlp")
	if err != nil {
		return mediaQueryResult{Error: err}
	}

	args := []string{
		fmt.Sprintf("ytsearch10:%s", strings.ReplaceAll(query, "\"", "")),
		"--extract-audio",
		"--audio-format", "opus",
		"--no-playlist",
		"--match-filter", fmt.Sprintf("duration < %d & !is_live", as.MaxAudioDurationInSeconds),
		"--max-downloads", "1",
		"--output", fmt.Sprintf("%s/%d-%%(id)s.opus", as.FileDirectory, start.Unix()),
		"--quiet",
		"--ignore-errors", // Ignores unavailable videos,
		"--print-json",
		"--no-color",
		"--no-check-formats",
	}

	cmd := exec.Command(ytDownloader, args...)
	response, err := cmd.Output()
	log.Println(err)
	if err != nil && err.Error() != "exit status 101" {
		return mediaQueryResult{Error: err}
	}

	videoData := videoData{}
	err = json.Unmarshal(response, &videoData)
	if err != nil {
		log.Println("error unmarshalling")
		return mediaQueryResult{Error: err}
	}

	return mediaQueryResult{Media: &disgotypes.Media{
		Title:     videoData.Title,
		FilePath:  fmt.Sprintf("%s/%d-%s.opus", as.FileDirectory, start.Unix(), videoData.ID),
		Uploader:  videoData.Uploader,
		URL:       fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoData.ID),
		Thumbnail: videoData.Thumbnail,
		Duration:  int(videoData.Duration),
	}, Error: nil}
}

type mediaQueryResult struct {
	Media *disgotypes.Media
	Error error
}

type videoData struct {
	ID                   string      `json:"id"`
	Title                string      `json:"title"`
	Thumbnail            string      `json:"thumbnail"`
	Description          string      `json:"description"`
	Uploader             string      `json:"uploader"`
	UploaderID           string      `json:"uploader_id"`
	UploaderURL          string      `json:"uploader_url"`
	ChannelID            string      `json:"channel_id"`
	ChannelURL           string      `json:"channel_url"`
	Duration             int         `json:"duration"`
	ViewCount            int         `json:"view_count"`
	AverageRating        interface{} `json:"average_rating"`
	AgeLimit             int         `json:"age_limit"`
	WebpageURL           string      `json:"webpage_url"`
	Categories           []string    `json:"categories"`
	Tags                 []string    `json:"tags"`
	PlayableInEmbed      bool        `json:"playable_in_embed"`
	LiveStatus           interface{} `json:"live_status"`
	ReleaseTimestamp     interface{} `json:"release_timestamp"`
	CommentCount         interface{} `json:"comment_count"`
	LikeCount            int         `json:"like_count"`
	Channel              string      `json:"channel"`
	ChannelFollowerCount int         `json:"channel_follower_count"`
	UploadDate           string      `json:"upload_date"`
	Availability         string      `json:"availability"`
	OriginalURL          string      `json:"original_url"`
	WebpageURLBasename   string      `json:"webpage_url_basename"`
	WebpageURLDomain     string      `json:"webpage_url_domain"`
	Extractor            string      `json:"extractor"`
	ExtractorKey         string      `json:"extractor_key"`
	PlaylistCount        int         `json:"playlist_count"`
	Playlist             string      `json:"playlist"`
	PlaylistID           string      `json:"playlist_id"`
	PlaylistTitle        string      `json:"playlist_title"`
	PlaylistUploader     interface{} `json:"playlist_uploader"`
	PlaylistUploaderID   interface{} `json:"playlist_uploader_id"`
	NEntries             int         `json:"n_entries"`
	PlaylistIndex        int         `json:"playlist_index"`
	LastPlaylistIndex    int         `json:"__last_playlist_index"`
	PlaylistAutonumber   int         `json:"playlist_autonumber"`
	DisplayID            string      `json:"display_id"`
	Fulltitle            string      `json:"fulltitle"`
	DurationString       string      `json:"duration_string"`
	RequestedSubtitles   interface{} `json:"requested_subtitles"`
	Asr                  int         `json:"asr"`
	Filesize             int         `json:"filesize"`
	FormatID             string      `json:"format_id"`
	FormatNote           string      `json:"format_note"`
	SourcePreference     int         `json:"source_preference"`
	Fps                  interface{} `json:"fps"`
	AudioChannels        int         `json:"audio_channels"`
	Height               interface{} `json:"height"`
	HasDrm               bool        `json:"has_drm"`
	Tbr                  float64     `json:"tbr"`
	URL                  string      `json:"url"`
	Width                interface{} `json:"width"`
	Language             string      `json:"language"`
	LanguagePreference   int         `json:"language_preference"`
	Preference           interface{} `json:"preference"`
	Ext                  string      `json:"ext"`
	Vcodec               string      `json:"vcodec"`
	Acodec               string      `json:"acodec"`
	DynamicRange         interface{} `json:"dynamic_range"`
	Abr                  float64     `json:"abr"`
	Filename             string      `json:"filename"`
}
