package api

import (
	"afho_backend/botClient"
	"afho_backend/utils"
	"database/sql"
	"errors"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func GetUserAvatar(discordClient *botClient.BotClient, userID string) string {
	member, err4 := discordClient.CacheHandler.Members.Get(func(t *discordgo.Member) bool {
		return t.User.ID == userID
	})
	if err4 != nil {
		utils.Logger.Error(err4.Error())
		return ""
	}

	avatarURL := member.User.Avatar
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
	statement, err := db.Prepare("SELECT * FROM Levels")
	if err != nil {
		utils.Logger.Error(err.Error())
		return []Level{}
	}
	defer statement.Close()

	rows, err2 := statement.Query()
	if err2 != nil {
		utils.Logger.Error(err2.Error())
		return []Level{}
	}

	var result []Level = make([]Level, 0, 100)
	for rows.Next() {
		tmp := Level{}
		err3 := rows.Scan(&tmp.User.User_id, &tmp.Xp)
		if err3 != nil {
			utils.Logger.Error(err3.Error())
			continue
		}

		tmp.Lvl = int(XptoLvl(tmp.Xp))

		member, err4 := discordClient.CacheHandler.Members.Get(func(t *discordgo.Member) bool {
			return t.User.ID == tmp.User.User_id
		})
		if err4 != nil {
			utils.Logger.Warn(err4.Error())
			continue
		}

		tmp.User.Username = member.User.Username
		tmp.User.DisplayAvatarURL = GetUserAvatar(discordClient, member.User.ID)

		result = append(result, tmp)
	}

	return result
}

func getBrasilBoardDB(db *sql.DB, discordClient *botClient.BotClient) []BrasilBoard {
	statement, err := db.Prepare("SELECT * FROM Bresil_count")
	if err != nil {
		utils.Logger.Error(err.Error())
		return []BrasilBoard{}
	}
	defer statement.Close()

	rows, err2 := statement.Query()
	if err2 != nil {
		utils.Logger.Error(err2.Error())
		return []BrasilBoard{}
	}

	var result []BrasilBoard = make([]BrasilBoard, 0, 100)
	for rows.Next() {
		tmp := BrasilBoard{}
		err3 := rows.Scan(&tmp.User.User_id, &tmp.BresilReceived, &tmp.BresilSent)
		if err3 != nil {
			utils.Logger.Error(err3.Error())
			continue
		}

		member, err4 := discordClient.CacheHandler.Members.Get(func(t *discordgo.Member) bool {
			return t.User.ID == tmp.User.User_id
		})
		if err4 != nil {
			// utils.Logger.Warn(err4.Error())
			continue
		}

		tmp.User.Username = member.User.Username
		tmp.User.DisplayAvatarURL = GetUserAvatar(discordClient, member.User.ID)

		result = append(result, tmp)
	}

	return result
}

func GetTimesDB(discordClient *botClient.BotClient, db *sql.DB) []Time {
	statement, err := db.Prepare("SELECT * FROM Time_connected")
	if err != nil {
		utils.Logger.Error(err.Error())
		return []Time{}
	}
	defer statement.Close()

	rows, err2 := statement.Query()
	if err2 != nil {
		utils.Logger.Error(err2.Error())
		return []Time{}
	}

	var result []Time = make([]Time, 0, 100)
	for rows.Next() {
		tmp := Time{}
		err3 := rows.Scan(&tmp.User.User_id, &tmp.TimeSpent)
		if err3 != nil {
			utils.Logger.Error(err3.Error())
			continue
		}

		member, err4 := discordClient.CacheHandler.Members.Get(func(t *discordgo.Member) bool {
			return t.User.ID == tmp.User.User_id
		})
		if err4 != nil {
			// utils.Logger.Warn(err4.Error())
			continue
		}

		tmp.User.Username = member.User.Username
		tmp.User.DisplayAvatarURL = GetUserAvatar(discordClient, member.User.ID)

		result = append(result, tmp)
	}

	return result
}

func GetAchievementsDB(discordClient *botClient.BotClient, db *sql.DB) []APIAchievement {
	statement, err := db.Prepare("SELECT user_id, achievement_name, Achievement_get.type, requirements FROM Achievement_get, Achievements WHERE Achievement_get.achievement_name = Achievements.name")
	if err != nil {
		utils.Logger.Error(err.Error())
		return []APIAchievement{}
	}
	defer statement.Close()

	rows, err2 := statement.Query()
	if err2 != nil {
		utils.Logger.Error(err2.Error())
	}

	tmpAll := make(map[string]APIAchievement, len(discordClient.CacheHandler.Members.Data))

	var result []APIAchievement = make([]APIAchievement, 0, 100)
	var counter int8 = 0
	for rows.Next() {
		var userId string
		tmpAchievement := Achievement{}
		err3 := rows.Scan(&userId, &tmpAchievement.Name, &tmpAchievement.Type, &tmpAchievement.Requirements)
		if err3 != nil {
			utils.Logger.Error(err3.Error())
			continue
		}

		member, err4 := discordClient.CacheHandler.Members.Get(func(t *discordgo.Member) bool {
			return t.User.ID == userId
		})
		if err4 != nil {
			// utils.Logger.Warn(err4.Error())
			continue
		}

		if val, ok := tmpAll[userId]; ok {
			old_achievements := val.Achievements
			old_achievements = append(old_achievements, tmpAchievement)
			val.Achievements = old_achievements
			tmpAll[userId] = val
		} else {
			tmp := APIAchievement{
				Counter:  counter,
				Id:       member.User.ID,
				Username: member.User.Username,
			}
			tmp.Achievements = append(tmp.Achievements, tmpAchievement)
			tmpAll[userId] = tmp
			counter++
		}
	}

	for key, value := range tmpAll {
		if len(value.Achievements) > 0 {
			result = append(result, value)
		} else {
			delete(tmpAll, key)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Counter < result[j].Counter
	})

	return result
}

func GetFavsDB(db *sql.DB, userId string) ([]Fav, error) {
	statement, err := db.Prepare("SELECT user_id, video_id, url, name, thumbnail, type, date_added FROM Favorites, Videos WHERE user_id = ? AND video_id = id")
	if err != nil {
		utils.Logger.Error(err.Error())
		return nil, errors.New("internal error")
	}
	defer statement.Close()

	rows, err2 := statement.Query(userId)
	if err2 != nil {
		return nil, errors.New("internal error")
	}

	var result []Fav = make([]Fav, 100)
	for rows.Next() {
		tmp := Fav{}
		err3 := rows.Scan(&tmp.User_id, &tmp.Id, &tmp.Url, &tmp.Name, &tmp.Thumbnail, &tmp.Type, &tmp.DateAdded)
		if err3 != nil {
			utils.Logger.Error(err3.Error())
			continue
		}

		result = append(result, tmp)
	}

	return result, nil
}

func AddFavDB(discordClient *botClient.BotClient, db *sql.DB, userId string, url string) (Fav, error) {
	songRegex, _ := regexp.Compile(`https?:\/\/(www.youtube.com|youtube.com)\/watch\?v=(?P<videoID>[^#\&\?]*)(&list=(?:[^#\&\?]*))?`)
	playlistRegex, _ := regexp.Compile(`https?:\/\/(?:www.youtube.com|youtube.com)\/(?:playlist\?list=(?P<listID>[^#\&\?]*)|watch\?v=(?:[^#\&\?]*)&list=(?P<listID2>[^#\&\?]*))`)

	isYoutubeSong := songRegex.MatchString(url)
	isYoutubePlaylist := playlistRegex.MatchString(url)

	if !isYoutubeSong && !isYoutubePlaylist {
		return Fav{}, errors.New("invalid url")
	}

	var video Fav
	var err2 error
	if isYoutubeSong {
		var videoId string
		if isYoutubePlaylist {
			tmp := songRegex.FindStringSubmatch(url)
			for i, name := range songRegex.SubexpNames() {
				if i != 0 && name != "" {
					if tmp[i] == "" {
						continue
					}
					videoId = tmp[i]
				}
			}
		}
		videoURL := "https://www.youtube.com/watch?v=" + videoId

		utils.Logger.Debug(videoURL)

		video, err2 = addFavoriteDB(discordClient, db, videoURL)
	} else {
		video, err2 = addFavoritePlaylistDB(discordClient, db, url)
	}
	if err2 != nil {
		return Fav{}, err2
	}

	statement, err := db.Prepare("INSERT INTO Favorites (user_id, video_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE user_id = user_id")
	if err != nil {
		utils.Logger.Error(err.Error())
		return Fav{}, errors.New("internal error")
	}
	defer statement.Close()

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
	statement2, err := db.Prepare("INSERT INTO Videos (id, url, name, thumbnail, type) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = id")
	if err != nil {
		utils.Logger.Error(err.Error())
		return Fav{}, errors.New("internal error")
	}
	defer statement2.Close()

	playlist, err2 := discordClient.MusicHandler.YoutubeClient.GetPlaylist(url)
	if err2 != nil {
		utils.Logger.Error(err2.Error())
		return Fav{}, errors.New("internal error")
	}

	_, err3 := statement2.Exec(playlist.ID, url, playlist.Title, utils.GetMaxResThumbnail(playlist.Videos[0].Thumbnails).URL, "playlist")
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
	statement2, err := db.Prepare("INSERT INTO Videos (id, url, name, thumbnail, type) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = id")
	if err != nil {
		utils.Logger.Error(err.Error())
		return Fav{}, errors.New("internal error")
	}
	defer statement2.Close()

	video, err2 := discordClient.MusicHandler.YoutubeClient.GetVideo(url)
	if err2 != nil {
		utils.Logger.Error(err2.Error())
		return Fav{}, errors.New("internal error")
	}

	_, err3 := statement2.Exec(video.ID, url, video.Title, utils.GetMaxResThumbnail(video.Thumbnails).URL, "video")
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

func RemoveFavoriteDB(db *sql.DB, userId string, videoId string) error {
	statement, err := db.Prepare("DELETE FROM Favorites WHERE user_id = ? AND video_id = ?")
	if err != nil {
		utils.Logger.Error(err.Error())
		return errors.New("internal error")
	}
	defer statement.Close()

	_, err2 := statement.Exec(userId, videoId)
	if err2 != nil {
		utils.Logger.Error(err2.Error())
		return errors.New("internal error")
	}

	return nil
}
