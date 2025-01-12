package utils

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
)

var (
	SongRegex, _     = regexp.Compile(`https?:\/\/(www.youtube.com|youtube.com)\/watch\?v=(?P<videoID>[^#\&\?]*)(&list=(?:[^#\&\?]*))?`)
	PlaylistRegex, _ = regexp.Compile(`https?:\/\/(?:www.youtube.com|youtube.com)\/(?:playlist\?list=(?P<listID>[^#\&\?]*)|watch\?v=(?:[^#\&\?]*)&list=(?P<listID2>[^#\&\?]*))`)
)

// FormatTime formats a time.Duration into a string with the format HH:MM:SS
func FormatTime(t time.Duration) string {
	hours := int(t.Hours())
	minutes := int(t.Minutes()) % 60
	seconds := int(t.Seconds()) % 60

	if hours == 0 && minutes == 0 {
		return fmt.Sprintf("00:%02d", seconds)
	}
	if hours == 0 {
		return fmt.Sprintf("%02d:%02d", minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func GetMaxResThumbnail(thumbnails youtube.Thumbnails) youtube.Thumbnail {
	maxResThumbnail := youtube.Thumbnail{
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

func InteractionReply(s *discordgo.Session, i *discordgo.InteractionCreate, message *discordgo.InteractionResponseData) {
	// Logger.Info("Replying to interaction", i.Interaction.ID, "with message", message.Content)
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: message,
	}); err != nil {
		Logger.Error("Error while replying to interaction", i.ID, "with message", message.Content)
		Logger.Error(err.Error())
	}
}

func GetUserDisplayName(member *discordgo.Member) string {
	if member.DisplayName() != "" {
		return member.DisplayName()
	}
	return member.User.Username
}

func RandomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[RandomInt(0, len(letterRunes)-1)]
	}
	return string(b)
}

func RandomInt(i1, i2 int) int {
	return i1 + rand.Intn(i2-i1+1)
}
