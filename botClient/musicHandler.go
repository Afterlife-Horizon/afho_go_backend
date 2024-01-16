package botClient

import (
	"afho__backend/utils"
	"context"
	"log"
	"regexp"
	"time"

	"github.com/Andreychik32/ytdl"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Track struct {
	ID             string
	VideoURL       string
	Title          string
	Autor          string
	Thumbnail      string
	Duration       time.Duration
	DurationString string
	RequestedBy    string
}

type Queue struct {
	Playing bool
	tracks  utils.Collection[Track]
}

func NewQueue() *Queue {
	return &Queue{
		Playing: false,
		tracks:  utils.NewCollection[Track]([]Track{}),
	}
}

type MusicHandler struct {
	Queue          *Queue
	YoutubeService *youtube.Service
}

func (handler *MusicHandler) Init(client *BotClient) {
	service, err := youtube.NewService(context.Background(), option.WithAPIKey(client.Config.YTApiKey))
	if err != nil {
		log.Println("Could not launch youtube service!")
		log.Println(err.Error())
	}

	handler.YoutubeService = service
}

func (handler *MusicHandler) Add(input string, requester string) string {
	var youtubRegex, _ = regexp.Compile(`/^(https?:\/\/)?(www\.)?(m\.|music\.)?(youtube\.com|youtu\.?be)\/.+$/gi`)
	var songRegex, _ = regexp.Compile(`/^.*(watch\?v=)([^#\&\?]*).*/gi`)
	var playlistRegex, _ = regexp.Compile(`/^.*(list=)([^#\&\?]*).*/gi`)

	var isYoutube = youtubRegex.MatchString(input)
	var isYoutubeSong = songRegex.MatchString(input)
	var isYoutubePlaylist = playlistRegex.MatchString(input)

	if !isYoutube {
		handler.AddSongSearch(input, requester)
		return ""
	}

	if !isYoutubeSong && isYoutubePlaylist {
		handler.AddPlayList(input, requester)
		return ""
	}

	if isYoutubeSong && !isYoutubePlaylist {
		handler.AddSong(input, requester)
		return ""
	}

	if isYoutubeSong && isYoutubePlaylist {
		handler.AddSongAndPlayList(input, requester)
		return ""
	}

	// TODO: Play song in channel

	return ""
}

func (handler *MusicHandler) AddSongSearch(input string, requester string) {
	response, err := handler.YoutubeService.Search.List([]string{"id", "snippet", "contentDetails"}).Q(input).MaxResults(1).Do()
	if err != nil {
		log.Println("Could not search for video!")
		log.Println(err.Error())
		return
	}

	var id string
	for _, item := range response.Items {
		id = item.Id.VideoId
	}

	if id == "" {
		log.Println("Video not found")
		return
	}

	var videoURL = "https://www.youtube.com/watch?v=" + id
	vid, err := ytdl.GetVideoInfo(context.Background(), videoURL)
	if err != nil {
		log.Println("Coul not get video info!")
		log.Println(err.Error())
		return
	}

	if handler.Queue == nil {
		handler.Queue = NewQueue()
	}

	handler.Queue.tracks.Insert(Track{
		ID:             id,
		VideoURL:       videoURL,
		Title:          vid.Title,
		Autor:          vid.Artist,
		Thumbnail:      vid.GetThumbnailURL(ytdl.ThumbnailQualityMaxRes).String(),
		Duration:       vid.Duration,
		DurationString: vid.Duration.String(),
		RequestedBy:    requester,
	})
}
func (handler *MusicHandler) AddSong(input string, requester string)            {}
func (handler *MusicHandler) AddPlayList(input string, requester string)        {}
func (handler *MusicHandler) AddSongAndPlayList(input string, requester string) {}
