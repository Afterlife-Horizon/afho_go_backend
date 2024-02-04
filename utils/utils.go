package utils

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
)

// FormatTime formats a time.Duration into a string with the format HH:MM:SS
func FormatTime(t time.Duration) string {
	var hours = int(t.Hours())
	var minutes = int(t.Minutes()) % 60
	var seconds = int(t.Seconds()) % 60

	if hours == 0 && minutes == 0 {
		return fmt.Sprintf("00:%02d", seconds)
	}
	if hours == 0 {
		return fmt.Sprintf("%02d:%02d", minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func GetMaxResThumbnail(thumbnails youtube.Thumbnails) youtube.Thumbnail {
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

func InteractionReply(s *discordgo.Session, i *discordgo.InteractionCreate, message *discordgo.InteractionResponseData) {
	// Logger.Info("Replying to interaction", i.Interaction.ID, "with message", message.Content)
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: message,
	}); err != nil {
		Logger.Error("Error while replying to interaction", i.Interaction.ID, "with message", message.Content)
		Logger.Error(err.Error())
	}
}
