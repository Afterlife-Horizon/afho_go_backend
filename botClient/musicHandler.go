package botClient

import (
	"afho__backend/utils"
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
	AudioFilter:      "",
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
	Paused         bool
}

func (handler *MusicHandler) Init(client *BotClient) {
	youtubeClient := youtube.Client{}
	service, err := youtubeApi.NewService(context.Background(), option.WithAPIKey(client.Config.YTApiKey))
	if err != nil {
		log.Println("Could not launch youtube service!")
		log.Println(err.Error())
	}

	handler.Queue = NewQueue()
	handler.YoutubeClient = &youtubeClient
	handler.YoutubeService = service
}

func checkForChannel(discordClient *BotClient, requesterId string) error {

	var _, err = discordClient.CacheHandler.VoiceConnectedMembers.Get(func(t *discordgo.Member) bool {
		return t.User.ID == requesterId
	})
	if err != nil {
		return errors.New("not in a voice channel")
	}

	return nil
}

func (handler *MusicHandler) Add(client *BotClient, input string, requester string, requesterId string, onTop bool) (string, error) {
	err := checkForChannel(client, requesterId)
	if err != nil {
		return err.Error(), err
	}
	go HandleJoin(client.Session, client.Config.GuildID, requesterId)

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

	if onTop && isYoutubePlaylist {
		return "Cannot add playlist on top of queue!", errors.New("cannot add playlist on top of queue")
	}

	if !isYoutubeSong && isYoutubePlaylist {
		err := handler.AddPlayList(input, requester, requesterId, playlistID)
		if err != nil {
			return err.Error(), err
		}
		returnValue = "Playlist added to queue!"
	} else if isYoutubeSong && !isYoutubePlaylist {
		var track, err = handler.AddSong(input, requester, requesterId, onTop)
		if err != nil {
			return err.Error(), err
		}
		returnValue = fmt.Sprintf("Track %v by %v added to queue!", track.Title, track.Author)
	} else if isYoutubeSong && isYoutubePlaylist {
		err := handler.AddSongAndPlayList(input, requester, requesterId, playlistID, 0)
		if err != nil {
			return err.Error(), err
		}
		returnValue = "Playlist added to queue!"
	} else {
		var track, err = handler.AddSongSearch(input, requester, requesterId, onTop)
		if err != nil {
			return err.Error(), err
		}
		returnValue = fmt.Sprintf("Track %v by %v added to queue!", track.Title, track.Author)
	}

	if !handler.Queue.Playing {
		var voiceConnection, ok = client.Session.VoiceConnections[client.Config.GuildID]
		if !ok {
			return "Could not get voiceConnection", errors.New("could not get voiceConnection")
		}

		go handler.DCA(client, handler.Queue.Tracks.Data[0].VideoURL, voiceConnection)
	}

	return returnValue, nil
}

func (handler *MusicHandler) AddOnTop(client *BotClient, input string, requester string, requesterId string) (string, error) {
	return handler.Add(client, input, requester, requesterId, true)
}

// ---------------------------- Add Helpers ----------------------------
func (handler *MusicHandler) AddSongSearch(input string, requester string, requesterId string, onTop bool) (Track, error) {
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
		Thumbnail:      utils.GetMaxResThumbnail(vid.Thumbnails).URL,
		Duration:       vid.Duration,
		DurationString: vid.Duration.String(),
		requesterId:    requesterId,
		RequestedBy:    requester,
	}

	if onTop {
		handler.Queue.Tracks.InsertAt(1, track)
		return track, nil
	}
	handler.Queue.Tracks.Insert(track)

	return track, nil
}
func (handler *MusicHandler) AddSong(input string, requester string, requesterId string, onTop bool) (Track, error) {
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
		Thumbnail:      utils.GetMaxResThumbnail(vid.Thumbnails).URL,
		Duration:       vid.Duration,
		DurationString: vid.Duration.String(),
		requesterId:    requesterId,
		RequestedBy:    requester,
	}

	if onTop {
		handler.Queue.Tracks.InsertAt(1, track)
		return track, nil
	}
	handler.Queue.Tracks.Insert(track)

	return track, nil
}
func (handler *MusicHandler) AddPlayList(input string, requester string, requesterId string, playlistID string) error {
	response, err := handler.YoutubeService.Playlists.List([]string{"id", "snippet"}).Id(playlistID).Do()
	if err != nil {
		log.Println("Could not search for playlist!")
		log.Println(err.Error())
		return errors.New("could not search for playlist")
	}

	var id string
	for _, item := range response.Items {
		id = item.Id
	}

	if id == "" {
		log.Println("Playlist not found")
		return errors.New("playlist not found")
	}

	playlist, err := handler.YoutubeClient.GetPlaylist(id)
	if err != nil {
		log.Println(err.Error())
		log.Println("Could not get playlist info!")
		return errors.New("could not get playlist info")
	}

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
				ID:             video.ID,
				VideoURL:       "https://www.youtube.com/watch?v=" + video.ID,
				Title:          video.Title,
				Author:         video.Author,
				Thumbnail:      utils.GetMaxResThumbnail(video.Thumbnails).URL,
				Duration:       video.Duration,
				DurationString: utils.FormatTime(video.Duration),
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
	return nil
}
func (handler *MusicHandler) AddSongAndPlayList(input string, requester string, playlistID string, requesterId string, baseIndex int) error {
	return handler.AddPlayList(input, requester, requesterId, playlistID)
}

// ---------------------------- Queue Helpers ----------------------------
func (handler *MusicHandler) Shuffle(client *BotClient) {
	handler.Queue.Tracks.Shuffle(1, len(handler.Queue.Tracks.Data), 3)
	fmt.Println(handler.Queue.Tracks.ToString())
}

func (handler *MusicHandler) ClearQueue(client *BotClient) {
	if len(handler.Queue.Tracks.Data) <= 0 {
		return
	}
	handler.Queue.Tracks.Data = []Track{handler.Queue.Tracks.Data[0]}
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

// ---------------------------- Queue Handlers ----------------------------
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

// ---------------------------- Player Handlers ----------------------------
func (handler *MusicHandler) Stop(client *BotClient) {
	handler.channel <- nil
	handler.Queue.Tracks.Data[0].Duration = 0
	handler.ClearQueue(client)
}
func (handler *MusicHandler) SetPause(pause bool) {
	handler.Stream.SetPaused(pause)
	handler.Paused = pause
	// handler.Queue.Playing = !pause
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

func (handler *MusicHandler) SkipTo(client *BotClient, index int) {
	handler.Queue.Tracks.Shift(index - 1)
	handler.channel <- nil
	handler.Queue.Tracks.Data[0].Duration = 0
}

func (handler *MusicHandler) SetFilters(client *BotClient, filters string) {
	baseOpts.AudioFilter = filters

	handler.stop <- true
	handler.EncodeSession.Stop()
	handler.EncodeSession.Cleanup()
	handler.Queue.SeekPosition = handler.Queue.SeekPosition + handler.Stream.PlaybackPosition()

	var voiceConnection, ok = client.Session.VoiceConnections[client.Config.GuildID]
	if !ok {
		return
	}
	handler.DCA(client, handler.Queue.Tracks.Data[0].VideoURL, voiceConnection)
}
