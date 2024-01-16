package botClient

import (
	"afho__backend/utils"
	"afho__backend/utils/commands"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	youtube "github.com/kkdai/youtube/v2"
	"google.golang.org/api/option"
	youtubeApi "google.golang.org/api/youtube/v3"
)

type Track struct {
	Video          *youtube.Video
	ID             string
	VideoURL       string
	Title          string
	Author         string
	Thumbnail      string
	Duration       time.Duration
	DurationString string
	requesterId    string
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
	stream         *dca.StreamingSession
	YoutubeService *youtubeApi.Service
	YoutubeClient  *youtube.Client
}

func (handler *MusicHandler) Init(client *BotClient) {
	youtubeClient := youtube.Client{}
	service, err := youtubeApi.NewService(context.Background(), option.WithAPIKey(client.Config.YTApiKey))
	if err != nil {
		log.Println("Could not launch youtube service!")
		log.Println(err.Error())
	}

	handler.YoutubeClient = &youtubeClient
	handler.YoutubeService = service
}

func (handler *MusicHandler) Add(client *BotClient, input string, requester string, requesterId string) string {
	var youtubRegex, _ = regexp.Compile(`/^(https?:\/\/)?(www\.)?(m\.|music\.)?(youtube\.com|youtu\.?be)\/.+$/gi`)
	var songRegex, _ = regexp.Compile(`/^.*(watch\?v=)([^#\&\?]*).*/gi`)
	var playlistRegex, _ = regexp.Compile(`/^.*(list=)([^#\&\?]*).*/gi`)

	var isYoutube = youtubRegex.MatchString(input)
	var isYoutubeSong = songRegex.MatchString(input)
	var isYoutubePlaylist = playlistRegex.MatchString(input)

	var returnValue string

	if !isYoutube {
		var track, err = handler.AddSongSearch(input, requester, requesterId)
		if err != nil {
			return err.Error()
		}
		returnValue = fmt.Sprintf("Track '%v' by '%v' added to queue!", track.Title, track.Author)
	} else if !isYoutubeSong && isYoutubePlaylist {
		handler.AddPlayList(input, requester)
	} else if isYoutubeSong && !isYoutubePlaylist {
		var track, err = handler.AddSong(input, requester, requesterId)
		if err != nil {
			return err.Error()
		}
		returnValue = fmt.Sprintf("Track '%v' by '%v' added to queue!", track.Title, track.Author)
	} else if isYoutubeSong && isYoutubePlaylist {
		handler.AddSongAndPlayList(input, requester)
	}

	commands.HandleJoin(client.Session, client.Config.GuildID, requesterId)

	if !handler.Queue.Playing {
		// Play sound in discord voice channel
		var voiceConnection, ok = client.Session.VoiceConnections[client.Config.GuildID]
		if !ok {
			returnValue = "Could not get voiceConnection"
		}

		handler.DCA(client, handler.Queue.tracks.Data[0].VideoURL, voiceConnection)

	}

	return returnValue
}

func (handler *MusicHandler) AddSongSearch(input string, requester string, requesterId string) (Track, error) {
	response, err := handler.YoutubeService.Search.List([]string{"id", "snippet"}).Q(input).MaxResults(1).Do()
	if err != nil {
		log.Println("Could not search for video!")
		log.Println(err.Error())
		return Track{}, errors.New("could not search for video")
	}

	var id string
	for _, item := range response.Items {
		id = item.Id.VideoId
	}

	if id == "" {
		log.Println("Video not found")
		return Track{}, errors.New("video not found")
	}

	vid, err := handler.YoutubeClient.GetVideo(id)
	if err != nil {
		log.Println(err.Error())
		log.Println("Could not get video info!")
		return Track{}, errors.New("could not get video info")
	}

	if handler.Queue == nil {
		handler.Queue = NewQueue()
	}

	var track = Track{
		Video:          vid,
		ID:             id,
		VideoURL:       "https://www.youtube.com/watch?v=" + id,
		Title:          vid.Title,
		Author:         vid.Author,
		Thumbnail:      getMaxResThumbnail(vid.Thumbnails).URL,
		Duration:       vid.Duration,
		DurationString: vid.Duration.String(),
		requesterId:    requesterId,
		RequestedBy:    requester,
	}

	handler.Queue.tracks.Insert(track)

	return track, nil
}
func (handler *MusicHandler) AddSong(input string, requester string, requesterId string) (Track, error) {
	var id, err = youtube.ExtractVideoID(input)
	if err != nil {
		log.Println(err.Error())
		log.Println("Could not extract ID from input!")
		return Track{}, errors.New("could not extract ID from input")
	}

	vid, err := handler.YoutubeClient.GetVideo(id)
	if err != nil {
		log.Println(err.Error())
		log.Println("Could not get video info!")
		return Track{}, errors.New("could not get video info")
	}

	if handler.Queue == nil {
		handler.Queue = NewQueue()
	}

	var track = Track{
		Video:          vid,
		ID:             id,
		VideoURL:       "https://www.youtube.com/watch?v=" + id,
		Title:          vid.Title,
		Author:         vid.Author,
		Thumbnail:      getMaxResThumbnail(vid.Thumbnails).URL,
		Duration:       vid.Duration,
		DurationString: vid.Duration.String(),
		requesterId:    requesterId,
		RequestedBy:    requester,
	}

	handler.Queue.tracks.Insert(track)

	return track, nil
}
func (handler *MusicHandler) AddPlayList(input string, requester string)        {}
func (handler *MusicHandler) AddSongAndPlayList(input string, requester string) {}

func getMaxResThumbnail(thumbnails youtube.Thumbnails) youtube.Thumbnail {
	var maxResThumbnail youtube.Thumbnail = youtube.Thumbnail{
		Width:  0,
		Height: 0,
	}
	for _, thumbnail := range thumbnails {
		if maxResThumbnail.Height <= thumbnail.Height && maxResThumbnail.Width <= thumbnail.Width {
			maxResThumbnail = thumbnail
		}
	}

	return maxResThumbnail
}

func (handler *MusicHandler) DCA(client *BotClient, url string, voiceConnection *discordgo.VoiceConnection) {
	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 96
	opts.Application = "lowdelay"

	formats := handler.Queue.tracks.Data[0].Video.Formats.WithAudioChannels()
	stream, _, err := handler.YoutubeClient.GetStream(handler.Queue.tracks.Data[0].Video, &formats[0])
	if err != nil {
		panic(err)
	}

	voiceConnection.Speaking(true)
	handler.Queue.Playing = true

	encodeSession, err := dca.EncodeMem(stream, opts)
	if err != nil {
		panic(err)
	}

	done := make(chan error)
	handler.stream = dca.NewStream(encodeSession, voiceConnection, done)

	go func() {
		for range done {
			err := <-done
			if err != nil && err != io.EOF {
				log.Println("FATA: An error occurred", err)
			}
			encodeSession.Cleanup()
			voiceConnection.Speaking(false)
			handler.Queue.Playing = false
			if len(handler.Queue.tracks.Data) == 1 {
				return
			}
			handler.Queue.tracks.Data = handler.Queue.tracks.Data[1:]
			fmt.Println(handler.Queue.tracks)

			handler.handleQueue(client, voiceConnection)
		}
	}()
}

func (handler *MusicHandler) handleQueue(client *BotClient, voiceConnection *discordgo.VoiceConnection) {
	if handler.Queue.Playing || len(handler.Queue.tracks.Data) <= 0 {
		return
	}

	handler.DCA(client, handler.Queue.tracks.Data[0].VideoURL, voiceConnection)
}
