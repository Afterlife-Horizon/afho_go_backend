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
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	youtube "github.com/kkdai/youtube/v2"
	"google.golang.org/api/option"
	youtubeApi "google.golang.org/api/youtube/v3"
)

var baseOpts = dca.EncodeOptions{
	Volume:           256,
	Channels:         2,
	FrameRate:        48000,
	FrameDuration:    20,
	Bitrate:          64,
	Application:      "lowdelay",
	CompressionLevel: 10,
	PacketLoss:       1,
	RawOutput:        true,
	VBR:              true,
	BufferedFrames:   100,
	Threads:          1,
	StartTime:        0,
}

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
	Playing      bool
	SeekPosition time.Duration
	Tracks       utils.Collection[Track]
	sync.RWMutex
}

func NewQueue() *Queue {
	return &Queue{
		Playing: false,
		Tracks:  utils.NewCollection[Track]([]Track{}),
	}
}

type MusicHandler struct {
	Queue          *Queue
	Stream         *dca.StreamingSession
	YoutubeService *youtubeApi.Service
	YoutubeClient  *youtube.Client
	EncodeSession  *dca.EncodeSession
	channel        chan error
	stop           chan bool
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
	// var youtubRegex, _ = regexp.Compile(`/https?\:\/\/(?:www\.youtube(?:\-nocookie)?\.com\/|m\.youtube\.com\/|youtube\.com\/)?(?:ytscreeningroom\?vi?=|youtu\.be\/|vi?\/|user\/.+\/u\/\w{1,2}\/|embed\/|watch\?(?:.*\&)?vi?=|\&vi?=|\?(?:.*\&)?vi?=)([^#\&\?\n\/<>\"']*)/`)
	var songRegex, _ = regexp.Compile(`https?:\/\/(www.youtube.com|youtube.com)\/watch\?v=(?P<videoID>[^#\&\?]*)(&list=(?:[^#\&\?]*))?`)
	var playlistRegex, _ = regexp.Compile(`https?:\/\/(?:www.youtube.com|youtube.com)\/(?:playlist\?list=(?P<listID>[^#\&\?]*)|watch\?v=(?:[^#\&\?]*)&list=(?P<listID2>[^#\&\?]*))`)

	// var isYoutube = youtubRegex.MatchString(input)
	var isYoutubeSong = songRegex.MatchString(input)
	var isYoutubePlaylist = playlistRegex.MatchString(input)

	var returnValue string

	// extract playlist id from url
	var playlistID string
	if isYoutubePlaylist {
		var tmp = playlistRegex.FindStringSubmatch(input)
		for i, name := range playlistRegex.SubexpNames() {
			if i != 0 && name != "" {
				if tmp[i] == "" {
					continue
				}
				playlistID = tmp[i]
			}
		}
	}

	// extract video id from url
	var videoID string
	if isYoutubeSong {
		var tmp = songRegex.FindStringSubmatch(input)
		for i, name := range songRegex.SubexpNames() {
			if i != 0 && name != "" {
				if tmp[i] == "" {
					continue
				}
				videoID = tmp[i]
			}
		}
	}

	fmt.Println(playlistID, videoID)

	fmt.Println(input, isYoutubeSong, isYoutubePlaylist)
	if !isYoutubeSong && isYoutubePlaylist {
		fmt.Println("playlist")
		handler.AddPlayList(input, requester, requesterId, playlistID)
		returnValue = "Playlist added to queue!"
	} else if isYoutubeSong && !isYoutubePlaylist {
		var track, err = handler.AddSong(input, requester, requesterId)
		if err != nil {
			return err.Error()
		}
		returnValue = fmt.Sprintf("Track %v by %v added to queue!", track.Title, track.Author)
	} else if isYoutubeSong && isYoutubePlaylist {
		handler.AddSongAndPlayList(input, requester, requesterId, playlistID, 0)
		returnValue = "Playlist added to queue!"
	} else {
		var track, err = handler.AddSongSearch(input, requester, requesterId)
		if err != nil {
			return err.Error()
		}
		returnValue = fmt.Sprintf("Track %v by %v added to queue!", track.Title, track.Author)
	}

	// fmt.Println(handler.Queue.tracks.Map(func(t Track) Track {
	// 	return Track{
	// 		Title:  t.Title,
	// 		Author: t.Author,
	// 	}
	// }).ToString())
	commands.HandleJoin(client.Session, client.Config.GuildID, requesterId)

	if !handler.Queue.Playing {
		var voiceConnection, ok = client.Session.VoiceConnections[client.Config.GuildID]
		if !ok {
			returnValue = "Could not get voiceConnection"
		}

		go handler.DCA(client, handler.Queue.Tracks.Data[0].VideoURL, voiceConnection)
	}

	return returnValue
}

func (handler *MusicHandler) AddOnTop(client *BotClient, input string, requester string, requesterId string) string {
	return "Not implemented yet"
	// // var youtubRegex, _ = regexp.Compile(`/^(https?:\/\/)?(www\.)?(m\.|music\.)?(youtube\.com|youtu\.?be)\/.+$/gi`)
	// var playlistRegex, _ = regexp.Compile(`/^.*(list=)([^#\&\?]*).*/gi`)

	// // var isYoutube = youtubRegex.MatchString(input)
	// var isYoutubePlaylist = playlistRegex.MatchString(input)

	// var returnValue string

	// if isYoutubePlaylist {
	// 	returnValue = "Cannot add playlist on top of queue"
	// 	return returnValue
	// }

	// commands.HandleJoin(client.Session, client.Config.GuildID, requesterId)

	// if !handler.Queue.Playing {
	// 	var voiceConnection, ok = client.Session.VoiceConnections[client.Config.GuildID]
	// 	if !ok {
	// 		returnValue = "Could not get voiceConnection"
	// 	}

	// 	go handler.DCA(client, handler.Queue.tracks.Data[0].VideoURL, voiceConnection)
	// }

	// returnValue = fmt.Sprintf("Track '%v' by '%v' added to queue!", handler.Queue.tracks.Data[0].Title, handler.Queue.tracks.Data[0].Author)

	// commands.HandleJoin(client.Session, client.Config.GuildID, requesterId)

	// return returnValue
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
		fmt.Println("Queue: ", handler.Queue)
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

	handler.Queue.Tracks.Insert(track)

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
		fmt.Println("Queue: ", handler.Queue)
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

	handler.Queue.Tracks.Insert(track)

	return track, nil
}
func (handler *MusicHandler) AddPlayList(input string, requester string, requesterId string, playlistID string) {
	response, err := handler.YoutubeService.Playlists.List([]string{"id", "snippet"}).Id(playlistID).Do()
	if err != nil {
		log.Println("Could not search for playlist!")
		log.Println(err.Error())
		return
	}

	var id string
	for _, item := range response.Items {
		id = item.Id
	}

	if id == "" {
		log.Println("Playlist not found")
		return
	}

	playlist, err := handler.YoutubeClient.GetPlaylist(id)
	if err != nil {
		log.Println(err.Error())
		log.Println("Could not get playlist info!")
		return
	}

	fmt.Println("Queue: ", handler.Queue)
	if handler.Queue == nil {
		handler.Queue = NewQueue()
		fmt.Println("Queue: ", handler.Queue)
	}

	var wg = sync.WaitGroup{}
	var lock = sync.Mutex{}
	length := len(playlist.Videos)
	var tracks = make([]Track, length)
	for i, vid := range playlist.Videos {
		wg.Add(1)
		go func(index int, vidID string) {
			video, err := handler.YoutubeClient.GetVideo(vidID)
			if err != nil {
				log.Println(err.Error())
				log.Println("Could not get video info!")
				wg.Done()
				return
			}

			var track = Track{
				Video:          video,
				ID:             id,
				VideoURL:       "https://www.youtube.com/watch?v=" + video.ID,
				Title:          video.Title,
				Author:         video.Author,
				Thumbnail:      getMaxResThumbnail(video.Thumbnails).URL,
				Duration:       video.Duration,
				DurationString: video.Duration.String(),
				requesterId:    requesterId,
				RequestedBy:    requester,
			}
			lock.Lock()
			tracks[index] = track
			lock.Unlock()
			wg.Done()
		}(i, vid.ID)
	}
	wg.Wait()

	handler.Queue.Tracks.Insert(tracks...)
}
func (handler *MusicHandler) AddSongAndPlayList(input string, requester string, playlistID string, requesterId string, baseIndex int) {
	handler.AddPlayList(input, requester, requesterId, playlistID)
}

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
	opts := &baseOpts
	opts.StartTime = int(handler.Queue.SeekPosition.Abs().Seconds())
	err := opts.Validate()
	if err != nil {
		panic(err)
	}

	formats := handler.Queue.Tracks.Data[0].Video.Formats.WithAudioChannels()
	stream, _, err := handler.YoutubeClient.GetStream(handler.Queue.Tracks.Data[0].Video, &formats[0])
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
	handler.Stream = dca.NewStream(encodeSession, voiceConnection, done)
	go func() {
		for range done {
			err := <-done
			if err != nil && err != io.EOF {
				log.Println("An error occured", err)
			}
		}
	}()

	var stop = make(chan bool)

	handler.channel = done
	handler.stop = stop
	var forceStopLoop = false
	handler.EncodeSession = encodeSession

	go func() {
		for range stop {
			handler.EncodeSession.Stop()
			handler.EncodeSession.Cleanup()
			voiceConnection.Speaking(false)
			handler.Queue.Playing = false
			forceStopLoop = true
		}
	}()

	go func() {
		for {
			var timeLeft = handler.Queue.Tracks.Data[0].Duration - (handler.Stream.PlaybackPosition() + handler.Queue.SeekPosition)
			// fmt.Println(timeLeft)
			if timeLeft.Milliseconds() < 1*time.Second.Milliseconds() {
				break
			}
			if forceStopLoop {
				return
			}
			time.Sleep(1 * time.Second)
		}
		encodeSession.Cleanup()
		voiceConnection.Speaking(false)
		handler.Queue.Playing = false
		handler.Queue.SeekPosition = 0
		handler.Queue.Tracks.Shift(1)
		handler.handleQueue(client, voiceConnection)
	}()
}

func (handler *MusicHandler) handleQueue(client *BotClient, voiceConnection *discordgo.VoiceConnection) {
	fmt.Println("handling queue: ", utils.Map[Track, string](&handler.Queue.Tracks, func(t Track) string {
		return fmt.Sprintf("%v - %v", t.Title, t.Author)
	}).ToString())

	if handler.Queue.Playing || len(handler.Queue.Tracks.Data) <= 0 {
		handler.channel = nil
		return
	}

	handler.DCA(client, handler.Queue.Tracks.Data[0].VideoURL, voiceConnection)
}

func (handler *MusicHandler) SetPause(pause bool) {
	handler.Stream.SetPaused(pause)
}

func (handler *MusicHandler) Seek(client *BotClient, position time.Duration) {
	if (handler.Queue.Tracks.Data[0].Duration - position).Milliseconds() < 1*time.Second.Milliseconds() {
		return
	}

	handler.stop <- true
	handler.EncodeSession.Stop()
	handler.EncodeSession.Cleanup()
	handler.Queue.SeekPosition = position

	var voiceConnection, ok = client.Session.VoiceConnections[client.Config.GuildID]
	if !ok {
		return
	}
	handler.DCA(client, handler.Queue.Tracks.Data[0].VideoURL, voiceConnection)
}

func (handler *MusicHandler) Skip(client *BotClient) {
	handler.channel <- nil
	handler.Queue.Tracks.Data[0].Duration = 0
}

func (handler *MusicHandler) Shuffle(client *BotClient) {
	handler.Queue.Tracks.Shuffle(1, len(handler.Queue.Tracks.Data), 3)
	fmt.Println(handler.Queue.Tracks.ToString())
}

func (handler *MusicHandler) Clear(client *BotClient) {
	handler.Queue.Tracks.Data = []Track{}
}

func (handler *MusicHandler) Remove(client *BotClient, index int) {
	handler.Queue.Tracks.RemoveItemAtIndex(index)
}

func (handler *MusicHandler) Move(client *BotClient, from int, to int) {
	if from < 1 || from > len(handler.Queue.Tracks.Data) || to < 1 || to > len(handler.Queue.Tracks.Data) {
		return
	}
	handler.Queue.Tracks.Data[from], handler.Queue.Tracks.Data[to] = handler.Queue.Tracks.Data[to], handler.Queue.Tracks.Data[from]
}
