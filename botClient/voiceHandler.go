package botClient

import (
	"errors"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

type VoiceHandler struct {
	client                *BotClient
	userCurrentVoiceTimes map[string]time.Duration
}

func NewVoiceHandler(client *BotClient) VoiceHandler {
	return VoiceHandler{
		client:                client,
		userCurrentVoiceTimes: make(map[string]time.Duration),
	}
}

func (v *VoiceHandler) UpdateDBTime() {
	for userID := range v.userCurrentVoiceTimes {
		v.updateUserDBTime(userID)
	}
}

func (v *VoiceHandler) updateUserDBTime(userID string) {
	var stmt, err = v.client.DB.Prepare("INSERT INTO Time_connected (user_id, time_spent) VALUES (?, ?) ON DUPLICATE KEY UPDATE time = time + ?")
	if err != nil {
		log.Println(err.Error())
		return
	}

	_, err = stmt.Exec(v.userCurrentVoiceTimes[userID].Seconds(), userID)
	if err != nil {
		log.Println(err.Error())
		return
	}

	v.userCurrentVoiceTimes[userID] = 0
}

func (v *VoiceHandler) JoinVoiceChannel(channelID string) error {
	var voiceState, err = v.client.Session.State.VoiceState(v.client.Config.GuildID, v.client.Session.State.User.ID)
	if err != nil || voiceState != nil {
		return errors.New("cannot join channel")
	}

	v.client.Session.ChannelVoiceJoin(v.client.Config.GuildID, channelID, false, true)
	return nil
}

func HandleJoin(s *discordgo.Session, GuildID string, userID string) (string, error) {
	var voiceState, err = s.State.VoiceState(GuildID, userID)
	if err != nil {
		log.Println(err.Error())
		return "Could not join voice channel!", err
	}

	s.ChannelVoiceJoin(voiceState.GuildID, voiceState.ChannelID, false, true)
	return "Joined Voice Channel", nil
}

func HandleLeave(s *discordgo.Session, GuildID string) string {
	var voiceState = s.VoiceConnections[GuildID]

	if voiceState == nil {
		return "Not in a voice channel!"
	}

	voiceState.Disconnect()
	return "Left the channel!"
}
