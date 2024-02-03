package api

import (
	"afho__backend/botClient"
	"database/sql"
	"errors"
	"log"
	"math"
	"strings"

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

func XptoLvl(xp int) float64 {
	var exp float64 = 2
	return math.Floor(math.Pow(float64(xp)/exp, 1/exp)) + 1
}

func GetAdmins(discordClient *botClient.BotClient) []string {
	var admins []string
	for _, member := range discordClient.CacheHandler.Members.Data {
		for _, role := range member.Roles {
			if role == discordClient.Config.AdminRoleID {
				admins = append(admins, member.User.Username)
				break
			}
		}
	}

	return admins
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

		tmp.Lvl = int(XptoLvl(tmp.Xp))

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

func getBrasilBoardDB(db *sql.DB, discordClient *botClient.BotClient) []BrasilBoard {
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

func GetAchievementsDB(discordClient *botClient.BotClient, db *sql.DB) []APIAchievement {
	var statement, err = db.Prepare("SELECT user_id, achievement_name, Achievement_get.type, requirements FROM Achievement_get, Achievements WHERE Achievement_get.achievement_name = Achievements.name")
	if err != nil {
		panic(err.Error())
	}
	defer statement.Close()

	var rows, err2 = statement.Query()
	if err2 != nil {
		panic(err2.Error())
	}

	var tmpAll = make(map[string]APIAchievement, len(discordClient.CacheHandler.Members.Data))

	var result []APIAchievement
	for rows.Next() {
		var userId string
		var tmpAchievement = Achievement{}
		err3 := rows.Scan(&userId, &tmpAchievement.Name, &tmpAchievement.Type, &tmpAchievement.Requirements)
		if err3 != nil {
			panic(err3.Error())
		}

		member, err4 := discordClient.CacheHandler.Members.Get(func(t *discordgo.Member) bool {
			return t.User.ID == userId
		})
		if err4 != nil {
			panic(err4.Error())
		}

		if val, ok := tmpAll[userId]; ok {
			val.Achievements = append(val.Achievements, tmpAchievement)
			tmpAll[userId] = val
		} else {
			var tmp = APIAchievement{
				Id:       member.User.ID,
				Username: member.User.Username,
			}
			tmp.Achievements = append(tmp.Achievements, tmpAchievement)
			tmpAll[tmp.Username] = tmp
		}
	}

	for key, value := range tmpAll {
		if len(value.Achievements) > 0 {
			result = append(result, value)
		} else {
			delete(tmpAll, key)
		}
	}

	return result
}

func GetFavsDB(discordClient *botClient.BotClient, db *sql.DB, userId string) ([]Fav, error) {
	var statement, err = db.Prepare("SELECT user_id, video_id, url, name, thumbnail, type, date_added FROM Favorites, Videos WHERE user_id = ?")
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("internal error")
	}
	defer statement.Close()

	var rows, err2 = statement.Query(userId)
	if err2 != nil {
		return nil, errors.New("internal error")
	}

	var result []Fav
	for rows.Next() {
		var tmp = Fav{}
		err3 := rows.Scan(&tmp.User_id, &tmp.Id, &tmp.Url, &tmp.Name, &tmp.Thumbnail, &tmp.Type, &tmp.DateAdded)
		if err3 != nil {
			log.Println(err3.Error())
			return nil, errors.New("internal error")
		}

		result = append(result, tmp)
	}

	return result, nil
}
