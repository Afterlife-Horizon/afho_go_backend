package api

import (
	"afho__backend/botClient"
	"afho__backend/utils"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
)

func GetLevels(discordClient *botClient.BotClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var levels []Level = GetLevelsDb(discordClient.DB, discordClient)
		sort.Slice(levels, func(i, j int) bool {
			return levels[i].Xp > levels[j].Xp
		})
		c.JSON(200, levels)
	}
}

func Fetch(discordClient *botClient.BotClient) func(c *gin.Context) {
	return func(c *gin.Context) {

		var time time.Duration
		var queue []Track
		var paused bool
		if discordClient.MusicHandler.Queue != nil && discordClient.MusicHandler.Queue.Playing {
			time = discordClient.MusicHandler.Queue.SeekPosition + discordClient.MusicHandler.Stream.PlaybackPosition()
			queue = utils.Map[botClient.Track, Track](&discordClient.MusicHandler.Queue.Tracks, func(track botClient.Track) Track {
				return Track{
					Id:                track.ID,
					Title:             track.Title,
					Duration:          int(track.Duration.Seconds()),
					DurationFormatted: FormatTime(track.Duration),
					Requester:         track.RequestedBy,
				}
			}).Data
			paused = !discordClient.MusicHandler.Queue.Playing
		} else {
			time = 0
			queue = []Track{}
			paused = false
		}

		c.JSON(200, FetchResults{
			Admins:       GetAdmins(discordClient),
			Formatedprog: FormatTime(time),
			Prog:         time.Seconds(),
			Queue: []Queue{
				{
					Effects: Effects{},
					Paused:  paused,
					Tracks:  queue,
				},
			},
		})
	}
}

func ConnectedMembers(discordClient *botClient.BotClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var connectedMembers []connectedMembers = utils.Map[*discordgo.Member](discordClient.CacheHandler.VoiceConnectedMembers, func(member *discordgo.Member) connectedMembers {
			return connectedMembers{
				Username: member.User.Username,
			}
		}).Data

		c.JSON(200, ConnectedMembersResponse{
			Data: connectedMembers,
		})
	}
}

func GetBrasilBoard(discordClient *botClient.BotClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var brasilBoard []BrasilBoard = GetBrasilBoardDB(discordClient.DB, discordClient)

		sort.Slice(brasilBoard, func(i, j int) bool {
			return brasilBoard[i].BresilReceived > brasilBoard[j].BresilReceived
		})

		c.JSON(200, brasilBoard)
	}
}

func GetTimes(discordClient *botClient.BotClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var times []Time = GetTimesDB(discordClient, discordClient.DB)

		sort.Slice(times, func(i, j int) bool {
			return times[i].TimeSpent > times[j].TimeSpent
		})

		c.JSON(200, times)
	}
}
