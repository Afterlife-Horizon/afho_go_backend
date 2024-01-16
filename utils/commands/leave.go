package commands

import (
	"github.com/bwmarrin/discordgo"
)

func HandleLeave(s *discordgo.Session, GuildID string, userID string) string {
	var voiceState = s.VoiceConnections[GuildID]

	if voiceState == nil {
		return "Not in a voice channel!"
	}

	voiceState.Disconnect()
	return "Left the channel!"
}
