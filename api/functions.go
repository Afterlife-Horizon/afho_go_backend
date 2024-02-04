package api

import (
	"afho__backend/botClient"
	"afho__backend/utils"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func GetUserAvatar(discordClient *botClient.BotClient, userID string) string {
	var member, err4 = discordClient.CacheHandler.Members.Get(func(t *discordgo.Member) bool {
		return t.User.ID == userID
	})
	if err4 != nil {
		utils.Logger.Error(err4.Error())
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
	defer statement.Close()
	if err != nil {
		panic(err.Error())
	}

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
	defer statement.Close()
	if err != nil {
		panic(err.Error())
	}

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
	defer statement.Close()
	if err != nil {
		panic(err.Error())
	}

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
	defer statement.Close()
	if err != nil {
		panic(err.Error())
	}

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
	var statement, err = db.Prepare("SELECT user_id, video_id, url, name, thumbnail, type, date_added FROM Favorites, Videos WHERE user_id = ? AND video_id = id")
	defer statement.Close()
	if err != nil {
		utils.Logger.Error(err.Error())
		return nil, errors.New("internal error")
	}

	var rows, err2 = statement.Query(userId)
	if err2 != nil {
		return nil, errors.New("internal error")
	}

	var result []Fav
	for rows.Next() {
		var tmp = Fav{}
		err3 := rows.Scan(&tmp.User_id, &tmp.Id, &tmp.Url, &tmp.Name, &tmp.Thumbnail, &tmp.Type, &tmp.DateAdded)
		if err3 != nil {
			utils.Logger.Error(err3.Error())
			return nil, errors.New("internal error")
		}

		result = append(result, tmp)
	}

	return result, nil
}

func AddFavDB(discordClient *botClient.BotClient, db *sql.DB, userId string, url string) (Fav, error) {
	var songRegex, _ = regexp.Compile(`https?:\/\/(www.youtube.com|youtube.com)\/watch\?v=(?P<videoID>[^#\&\?]*)(&list=(?:[^#\&\?]*))?`)
	var playlistRegex, _ = regexp.Compile(`https?:\/\/(?:www.youtube.com|youtube.com)\/(?:playlist\?list=(?P<listID>[^#\&\?]*)|watch\?v=(?:[^#\&\?]*)&list=(?P<listID2>[^#\&\?]*))`)

	var isYoutubeSong = songRegex.MatchString(url)
	var isYoutubePlaylist = playlistRegex.MatchString(url)

	if !isYoutubeSong && !isYoutubePlaylist {
		return Fav{}, errors.New("invalid url")
	}

	var video Fav
	var err2 error
	if isYoutubeSong {
		var videoId string
		if isYoutubePlaylist {
			var tmp = songRegex.FindStringSubmatch(url)
			for i, name := range songRegex.SubexpNames() {
				if i != 0 && name != "" {
					if tmp[i] == "" {
						continue
					}
					videoId = tmp[i]
				}
			}
		}
		var videoURL = "https://www.youtube.com/watch?v=" + videoId

		fmt.Println(videoURL)

		video, err2 = addFavoriteDB(discordClient, db, videoURL)
	} else {
		video, err2 = addFavoritePlaylistDB(discordClient, db, url)
	}
	if err2 != nil {
		return Fav{}, err2
	}

	var statement, err = db.Prepare("INSERT INTO Favorites (user_id, video_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE user_id = user_id")
	defer statement.Close()
	if err != nil {
		utils.Logger.Error(err.Error())
		return Fav{}, errors.New("internal error")
	}

	_, err3 := statement.Exec(userId, video.Id)
	if err3 != nil {
		utils.Logger.Error(err3.Error())
		return Fav{}, errors.New("internal error")
	}

	return Fav{
		User_id:   userId,
		Id:        video.Id,
		Url:       url,
		Name:      video.Name,
		Thumbnail: video.Thumbnail,
		Type:      video.Type,
	}, nil
}

func addFavoritePlaylistDB(discordClient *botClient.BotClient, db *sql.DB, url string) (Fav, error) {
	var statement2, err = db.Prepare("INSERT INTO Videos (id, url, name, thumbnail, type) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = id")
	defer statement2.Close()
	if err != nil {
		utils.Logger.Error(err.Error())
		return Fav{}, errors.New("internal error")
	}

	var playlist, err2 = discordClient.MusicHandler.YoutubeClient.GetPlaylist(url)
	if err2 != nil {
		utils.Logger.Error(err2.Error())
		return Fav{}, errors.New("internal error")
	}

	var _, err3 = statement2.Exec(playlist.ID, url, playlist.Title, utils.GetMaxResThumbnail(playlist.Videos[0].Thumbnails).URL, "playlist")
	if err3 != nil {
		utils.Logger.Error(err3.Error())
		return Fav{}, errors.New("internal error")
	}
	return Fav{
		Id:        playlist.ID,
		Url:       url,
		Name:      playlist.Title,
		Thumbnail: utils.GetMaxResThumbnail(playlist.Videos[0].Thumbnails).URL,
		Type:      "playlist",
		DateAdded: time.Now().Local().String(),
	}, nil
}

func addFavoriteDB(discordClient *botClient.BotClient, db *sql.DB, url string) (Fav, error) {
	var statement2, err = db.Prepare("INSERT INTO Videos (id, url, name, thumbnail, type) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = id")
	defer statement2.Close()
	if err != nil {
		utils.Logger.Error(err.Error())
		return Fav{}, errors.New("internal error")
	}

	var video, err2 = discordClient.MusicHandler.YoutubeClient.GetVideo(url)
	if err2 != nil {
		utils.Logger.Error(err2.Error())
		return Fav{}, errors.New("internal error")
	}

	var _, err3 = statement2.Exec(video.ID, url, video.Title, utils.GetMaxResThumbnail(video.Thumbnails).URL, "video")
	if err3 != nil {
		utils.Logger.Error(err3.Error())
		return Fav{}, errors.New("internal error")
	}
	return Fav{
		Id:        video.ID,
		Url:       url,
		Name:      video.Title,
		Thumbnail: utils.GetMaxResThumbnail(video.Thumbnails).URL,
		Type:      "playlist",
		DateAdded: time.Now().Local().String(),
	}, nil
}

func RemoveFavoriteDB(discordClient *botClient.BotClient, db *sql.DB, userId string, videoId string) error {
	var statement, err = db.Prepare("DELETE FROM Favorites WHERE user_id = ? AND video_id = ?")
	defer statement.Close()
	if err != nil {
		utils.Logger.Error(err.Error())
		return errors.New("internal error")
	}

	_, err2 := statement.Exec(userId, videoId)
	if err2 != nil {
		utils.Logger.Error(err2.Error())
		return errors.New("internal error")
	}

	return nil
}
