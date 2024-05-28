package botClient

import (
	"afho_backend/utils"
	"context"
	"errors"
	"fmt"
	"io"
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
	VBR:              false,
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
	DurationString string
	requesterId    string
	RequestedBy    string
	Duration       time.Duration
}

type Queue struct {
	Tracks       utils.Collection[Track]
	SeekPosition time.Duration
	sync.RWMutex
	Playing              bool
	forceStopCurrentSong bool
}

func NewQueue() *Queue {
	utils.Logger.Debug("Creating new Queue")
	return &Queue{
		Playing: false,
		Tracks:  utils.NewCollection([]Track{}),
	}
}

type MusicHandler struct {
	Queue          *Queue
	Stream         *dca.StreamingSession
	YoutubeService *youtubeApi.Service
	YoutubeClient  *youtube.Client
	EncodeSession  *dca.EncodeSession
	channel        chan error
	Paused         bool
	Speaking       bool
	Changing       bool
}

func (handler *MusicHandler) Init(client *BotClient) {
	utils.Logger.Debug("Initialising Music Handler")
	youtubeClient := youtube.Client{}
	utils.Logger.Debug("Creating Youtube Service")
	service, err := youtubeApi.NewService(context.Background(), option.WithAPIKey(client.Config.YTApiKey))
	if err != nil {
		utils.Logger.Error("Could not launch youtube service!")
		utils.Logger.Error(err.Error())
	}

	handler.Queue = NewQueue()
	handler.YoutubeClient = &youtubeClient
	handler.YoutubeService = service
}

func checkForChannel(discordClient *BotClient, requesterId string) error {
	utils.Logger.Debug("Checking if user is in a voice channel")
	_, err := discordClient.CacheHandler.VoiceConnectedMembers.Get(func(t *discordgo.Member) bool {
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

	if client.VoiceHandler.voiceConnection == nil {
		_, err = HandleJoin(client, client.Config.GuildID, requesterId)
		if err != nil {
			utils.Logger.Error(err.Error())
			return "Could not join voice channel", err
		}
	}

	isYoutubeSong := utils.SongRegex.MatchString(input)
	isYoutubePlaylist := utils.PlaylistRegex.MatchString(input)

	utils.Logger.Debug("Checking input type", "isYoutubeSong", isYoutubeSong, "isYoutubePlaylist", isYoutubePlaylist)

	var returnValue string

	// extract playlist id from url
	var playlistID string
	if isYoutubePlaylist {
		tmp := utils.PlaylistRegex.FindStringSubmatch(input)
		for i, name := range utils.PlaylistRegex.SubexpNames() {
			if i != 0 && name != "" {
				if tmp[i] == "" {
					continue
				}
				playlistID = tmp[i]
			}
		}
	}

	utils.Logger.Debug("Playlist ID", playlistID)

	if onTop && isYoutubePlaylist {
		utils.Logger.Debug("On top and is playlist!")
		return "Cannot add playlist on top of queue!", errors.New("cannot add playlist on top of queue")
	}

	if !isYoutubeSong && isYoutubePlaylist {
		err := handler.AddPlayList(input, requester, requesterId, playlistID)
		if err != nil {
			return err.Error(), err
		}
		returnValue = "Playlist added to queue!"
	} else if isYoutubeSong && !isYoutubePlaylist {
		track, err := handler.AddSong(input, requester, requesterId, onTop)
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
		track, err := handler.AddSongSearch(input, requester, requesterId, onTop)
		if err != nil {
			return err.Error(), err
		}
		returnValue = fmt.Sprintf("Track %v by %v added to queue!", track.Title, track.Author)
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
		utils.Logger.Warn("Could not search for video!")
		utils.Logger.Warn(err.Error())
		return Track{}, errors.New("could not search for video")
	}

	var id string
	for _, item := range response.Items {
		id = item.Id.VideoId
	}

	if id == "" {
		utils.Logger.Info("Video not found")
		return Track{}, errors.New("video not found")
	}

	vid, err := handler.YoutubeClient.GetVideo(id)
	if err != nil {
		utils.Logger.Warn(err.Error())
		utils.Logger.Warn("Could not get video info!")
		return Track{}, errors.New("could not get video info")
	}

	if handler.Queue == nil {
		handler.Queue = NewQueue()
	}

	track := Track{
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
	id, err := youtube.ExtractVideoID(input)
	if err != nil {
		utils.Logger.Warn(err.Error())
		utils.Logger.Warn("Could not extract ID from input!")
		return Track{}, errors.New("could not extract ID from input")
	}

	vid, err := handler.YoutubeClient.GetVideo(id)
	if err != nil {
		utils.Logger.Warn(err.Error())
		utils.Logger.Warn("Could not get video info!")
		return Track{}, errors.New("could not get video info")
	}

	if handler.Queue == nil {
		handler.Queue = NewQueue()
	}

	track := Track{
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
	utils.Logger.Debug("Adding playlist")

	playlist, err := handler.YoutubeClient.GetPlaylist(input)
	if err != nil {
		utils.Logger.Warn("Unable to get playlist!")
		return errors.New("unable to get playlist")
	}

	if handler.Queue == nil {
		handler.Queue = NewQueue()
		// utils.Logger.Debug("Queue: ", handler.Queue)
	}

	wg := sync.WaitGroup{}
	lock := sync.Mutex{}
	length := len(playlist.Videos)
	tracks := make([]Track, length)
	for i, vid := range playlist.Videos {
		utils.Logger.Debug("Getting video info", vid.ID)
		wg.Add(1)
		go func(index int, vidID string) {
			video, err := handler.YoutubeClient.GetVideo(vidID)
			if err != nil {
				utils.Logger.Error(err.Error())
				wg.Done()
				return
			}

			track := Track{
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

	utils.Logger.Debug("Added Tracks", tracks)
	handler.Queue.Tracks.Insert(tracks...)
	return nil
}

func (handler *MusicHandler) AddSongAndPlayList(input string, requester string, playlistID string, requesterId string, baseIndex int) error {
	return handler.AddPlayList(input, requester, requesterId, playlistID)
}

// ---------------------------- Queue Helpers ----------------------------
func (handler *MusicHandler) Shuffle(client *BotClient) {
	utils.Logger.Debug("Shuffling queue")
	handler.Queue.Tracks.Shuffle(1, len(handler.Queue.Tracks.Data), 3)
}

func (handler *MusicHandler) ClearQueue(client *BotClient) {
	utils.Logger.Debug("Clearing queue")
	if len(handler.Queue.Tracks.Data) <= 0 {
		return
	}
	handler.Queue.Tracks.Data = []Track{handler.Queue.Tracks.Data[0]}
}

func (handler *MusicHandler) Remove(client *BotClient, index int) {
	utils.Logger.Debug("Removing track", index)
	handler.Queue.Tracks.RemoveItemAtIndex(index)
}

func (handler *MusicHandler) Move(client *BotClient, from int, to int) {
	utils.Logger.Debug("Moving track", from, to)
	if from < 1 || from > len(handler.Queue.Tracks.Data) || to < 1 || to > len(handler.Queue.Tracks.Data) {
		return
	}
	handler.Queue.Tracks.Data[from], handler.Queue.Tracks.Data[to] = handler.Queue.Tracks.Data[to], handler.Queue.Tracks.Data[from]
}

// ---------------------------- Queue Handlers ----------------------------
func (handler *MusicHandler) DCA(client *BotClient) {
	utils.Logger.Debug("DCA")
	opts := &baseOpts
	opts.StartTime = int(handler.Queue.SeekPosition.Abs().Seconds())
	err := opts.Validate()
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	formats := handler.Queue.Tracks.Data[0].Video.Formats.WithAudioChannels()
	stream, _, err := handler.YoutubeClient.GetStream(handler.Queue.Tracks.Data[0].Video, &formats[0])
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	if !handler.Speaking {
		err = client.VoiceHandler.voiceConnection.Speaking(true)
		if err != nil {
			utils.Logger.Error(err.Error())
			return
		}
	}
	handler.Queue.Playing = true

	encodeSession, err := dca.EncodeMem(stream, opts)
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	done := make(chan error)
	handler.Stream = dca.NewStream(encodeSession, client.VoiceHandler.voiceConnection, done)
	handler.channel = done
	handler.EncodeSession = encodeSession

	go func() {
		for range done {
			err := <-done
			if err != nil && err != io.EOF {
				utils.Logger.Error(err)
				handler.Queue.forceStopCurrentSong = true
			}
		}
	}()
}

func (handler *MusicHandler) shiftQueue(client *BotClient) {
	utils.Logger.Debug("Track finished/Stopped")
	handler.EncodeSession.Cleanup()
	err := client.VoiceHandler.voiceConnection.Speaking(false)
	if err != nil {
		utils.Logger.Error(err)
	}
	handler.Queue.forceStopCurrentSong = false
	handler.Queue.Playing = false
	handler.Queue.SeekPosition = 0
	handler.channel = nil
	handler.Queue.Tracks.Shift(1)
}

func (handler *MusicHandler) HandleQueue(client *BotClient) {
	previous_time := time.Duration.Milliseconds(0)
	for {
		if handler.Queue == nil {
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		handler.Queue.RLock()
		if handler.Queue.Playing || handler.Paused {
			currentPos := (handler.Stream.PlaybackPosition() + handler.Queue.SeekPosition)
			timeLeft := handler.Queue.Tracks.Data[0].Duration - currentPos
			utils.Logger.Debugf("time left: %vs\n", timeLeft)
			if (timeLeft.Milliseconds() == previous_time && !handler.Paused && !handler.Changing) || timeLeft.Milliseconds() < 1*time.Second.Milliseconds() || handler.Queue.forceStopCurrentSong {
				handler.Changing = true
				handler.shiftQueue(client)
			}

			if currentPos > 1*time.Millisecond {
				handler.Changing = false
			}

			previous_time = timeLeft.Milliseconds()
		} else if len(handler.Queue.Tracks.Data) > 0 {
			handler.Changing = true
			handler.DCA(client)
			handler.Changing = false
		} else {
			if handler.Speaking && client.VoiceHandler.voiceConnection != nil {
				err := client.VoiceHandler.voiceConnection.Speaking(false)
				if err != nil {
					utils.Logger.Error(err)
				}
			}
		}
		handler.Queue.RUnlock()

		time.Sleep(200 * time.Millisecond)
	}
}

// ---------------------------- Player Handlers ----------------------------
func (handler *MusicHandler) Stop(client *BotClient) {
	utils.Logger.Debug("Stopping")
	handler.ClearQueue(client)
	handler.Queue.Tracks.Data[0].Duration = 0
	handler.Queue.forceStopCurrentSong = true
}

func (handler *MusicHandler) SetPause(pause bool) {
	utils.Logger.Debug("Setting pause", pause)
	handler.Stream.SetPaused(pause)
	handler.Paused = pause
}

func (handler *MusicHandler) Seek(client *BotClient, position time.Duration) {
	utils.Logger.Debug("Seeking", position)
	if (handler.Queue.Tracks.Data[0].Duration - position).Milliseconds() < 1*time.Second.Milliseconds() {
		return
	}

	handler.Changing = true

	handler.EncodeSession.Cleanup()
	handler.Queue.SeekPosition = position

	handler.DCA(client)

	handler.Changing = false
}

func (handler *MusicHandler) Skip(client *BotClient) {
	utils.Logger.Debug("Skipping")
	handler.channel <- nil
	handler.Queue.Tracks.Data[0].Duration = 0
}

func (handler *MusicHandler) SkipTo(client *BotClient, index int) {
	utils.Logger.Debug("Skipping to", index)
	handler.Queue.Tracks.Shift(index - 1)
	handler.channel <- nil
	handler.Queue.Tracks.Data[0].Duration = 0
}

func (handler *MusicHandler) SetFilters(client *BotClient, filters string) {
	utils.Logger.Debug("Setting filters", filters)
	baseOpts.AudioFilter = filters

	handler.Changing = true
	handler.EncodeSession.Cleanup()
	handler.Queue.SeekPosition = handler.Queue.SeekPosition + handler.Stream.PlaybackPosition()

	handler.DCA(client)
	handler.Changing = false
}
