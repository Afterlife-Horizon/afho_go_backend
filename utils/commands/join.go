package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func HandleJoin(s *discordgo.Session, GuildID string, userID string) string {
	var voiceState, err = s.State.VoiceState(GuildID, userID)
	if err != nil {
		log.Println(err.Error())
		return "Could not join voice channel!"
	}

	s.ChannelVoiceJoin(voiceState.GuildID, voiceState.ChannelID, false, true)
	return "Joined Voice Channel"
}
