package botClient

import "errors"

type VoiceHandler struct {
	client *BotClient
}

func NewVoiceHandler(client *BotClient) VoiceHandler {
	return VoiceHandler{
		client: client,
	}
}

func (v *VoiceHandler) JoinVoiceChannel(channelID string) error {
	var voiceState, err = v.client.Session.State.VoiceState(v.client.Config.GuildID, v.client.Session.State.User.ID)
	if err != nil || voiceState != nil {
		return errors.New("cannot join channel")
	}

	v.client.Session.ChannelVoiceJoin(v.client.Config.GuildID, channelID, false, true)
	return nil
}
