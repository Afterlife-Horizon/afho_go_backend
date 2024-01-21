package api

import (
	"afho__backend/botClient"
	"afho__backend/utils"
	"database/sql"
	"log"
	"math"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func GetUserAvatar(discordClient *botClient.BotClient, userID string) string {
	var member, err4 = discordClient.CacheHandler.Members.Get(func(t *discordgo.Member) bool {
		return t.User.ID == userID
	})
	if err4 != nil {
		log.Printf(err4.Error())
	}

	var avatarURL = member.User.Avatar
	if strings.HasPrefix(avatarURL, "a_") {
		avatarURL = "https://cdn.discordapp.com/avatars/" + member.User.ID + "/" + avatarURL + ".gif"
	} else {
		avatarURL = "https://cdn.discordapp.com/avatars/" + member.User.ID + "/" + avatarURL + ".png"
	}

	return avatarURL
}

func xptoLvl(xp int) float64 {
	var exp float64 = 2
	return math.Floor(math.Pow(float64(xp)/exp, 1/exp)) + 1
}

func GetAdmins(discordClient *botClient.BotClient) []string {
	var admins = discordClient.CacheHandler.Members.Filter(func(member *discordgo.Member) bool {
		return member.Permissions&8 == 8
	})
	var adminNames []string = utils.Map[*discordgo.Member, string](admins, func(member *discordgo.Member) string {
		return member.User.Username
	}).Data

	return adminNames
}

func GetLevelsDb(db *sql.DB, discordClient *botClient.BotClient) []Level {
	var statement, err = db.Prepare("SELECT * FROM Levels")
	if err != nil {
		panic(err.Error())
	}
	defer statement.Close()

	var rows, err2 = statement.Query()
	if err2 != nil {
		panic(err2.Error())
	}

	var result []Level
	for rows.Next() {
		var tmp = Level{}
		err3 := rows.Scan(&tmp.User.User_id, &tmp.Xp)
		if err3 != nil {
			panic(err3.Error())
		}

		tmp.Lvl = int(xptoLvl(tmp.Xp))

		member, err4 := discordClient.CacheHandler.Members.Get(func(t *discordgo.Member) bool {
			return t.User.ID == tmp.User.User_id
		})
		if err4 != nil {
			panic(err4.Error())
		}

		tmp.User.DisplayAvatarURL = GetUserAvatar(discordClient, member.User.ID)

		result = append(result, tmp)
	}

	return result
}

// FormatTime formats a time.Duration into a string with the format HH:MM:SS
func FormatTime(t time.Duration) string {
	var hours = t / time.Hour
	var minutes = t / time.Minute % 60
	var seconds = t / time.Second % 60
	if hours == 0 && minutes == 0 {
		return seconds.String()
	} else if hours == 0 {
		return minutes.String() + ":" + seconds.String()
	}
	return hours.String() + ":" + minutes.String() + ":" + seconds.String()
}

func GetBrasilBoardDB(db *sql.DB, discordClient *botClient.BotClient) []BrasilBoard {
	var statement, err = db.Prepare("SELECT * FROM Bresil_count")
	if err != nil {
		panic(err.Error())
	}
	defer statement.Close()

	var rows, err2 = statement.Query()
	if err2 != nil {
		panic(err2.Error())
	}

	var result []BrasilBoard
	for rows.Next() {
		var tmp = BrasilBoard{}
		err3 := rows.Scan(&tmp.User.User_id, &tmp.BresilReceived, &tmp.BresilSent)
		if err3 != nil {
			panic(err3.Error())
		}

		member, err4 := discordClient.CacheHandler.Members.Get(func(t *discordgo.Member) bool {
			return t.User.ID == tmp.User.User_id
		})
		if err4 != nil {
			panic(err4.Error())
		}

		tmp.User.DisplayAvatarURL = GetUserAvatar(discordClient, member.User.ID)

		result = append(result, tmp)
	}

	return result
}

func GetTimesDB(discordClient *botClient.BotClient, db *sql.DB) []Time {
	var statement, err = db.Prepare("SELECT * FROM Time_connected")
	if err != nil {
		panic(err.Error())
	}
	defer statement.Close()

	var rows, err2 = statement.Query()
	if err2 != nil {
		panic(err2.Error())
	}

	var result []Time
	for rows.Next() {
		var tmp = Time{}
		err3 := rows.Scan(&tmp.User.User_id, &tmp.TimeSpent)
		if err3 != nil {
			panic(err3.Error())
		}

		member, err4 := discordClient.CacheHandler.Members.Get(func(t *discordgo.Member) bool {
			return t.User.ID == tmp.User.User_id
		})
		if err4 != nil {
			panic(err4.Error())
		}

		tmp.User.DisplayAvatarURL = GetUserAvatar(discordClient, member.User.ID)

		result = append(result, tmp)
	}

	return result
}
