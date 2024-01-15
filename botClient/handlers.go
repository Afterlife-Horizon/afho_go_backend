package botClient

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func MessageCreate(client *BotClient) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Member.User.ID == s.State.User.ID {
			return
		}
	}
}

func VoiceStateUpdate(client *BotClient) func(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	return func(s *discordgo.Session, state *discordgo.VoiceStateUpdate) {
		switch getStateChangeType(client, state) {
		case "join":
			client.CacheHandler.addConnectedMember(state.Member)
		case "leave":
			client.CacheHandler.removeConnectedMember(state.Member)
		default:

		}
	}
}

func getStateChangeType(client *BotClient, newState *discordgo.VoiceStateUpdate) string {
	var oldState = newState.BeforeUpdate

	client.CacheHandler.ChannelsMutex.RLock()
	var newChannel, _ = client.CacheHandler.Channels.Get(func(c *discordgo.Channel) bool {
		return c.Type == discordgo.ChannelTypeGuildVoice && c.GuildID == newState.GuildID
	})

	var oldChannel, _ = client.CacheHandler.Channels.Get(func(c *discordgo.Channel) bool {
		return oldState != nil && c.Type == discordgo.ChannelTypeGuildVoice && c.GuildID == oldState.GuildID
	})
	client.CacheHandler.ChannelsMutex.RUnlock()

	// --------- Voice leave / join ---------
	if oldState == nil || (newState.ChannelID != "" && oldState.ChannelID == "") {
		log.Printf("%v joined %v\n", newState.Member.User.Username, newChannel.Name)
		return "join"
	}
	if newState.ChannelID == "" && oldState.ChannelID != "" {
		log.Printf("%v left %v\n", oldState.Member.User.Username, oldChannel.Name)
		return "leave"
	}

	// --------- Server deafen / undeafened ---------
	if newState.Deaf && !oldState.Deaf {
		log.Printf("%v was deafened\n", newState.Member.User.Username)
		return "server deafen"
	}
	if !newState.Deaf && oldState.Deaf {
		log.Printf("%v was undeafened\n", newState.Member.User.Username)
		return "server undeafen"
	}

	// --------- Self deafen / undeafened ---------
	if newState.SelfDeaf && !oldState.SelfDeaf {
		log.Printf("%v was deafened\n", newState.Member.User.Username)
		return "self deafen"
	}
	if !newState.SelfDeaf && oldState.SelfDeaf {
		log.Printf("%v was undeafened\n", newState.Member.User.Username)
		return "self undeafen"
	}

	// --------- Server Mute / Unmute ---------
	if newState.Mute && !oldState.Mute {
		log.Printf("%v was muted\n", newState.Member.User.Username)
		return "server mute"
	}
	if !newState.Mute && oldState.Mute {
		log.Printf("%v was unmuted\n", newState.Member.User.Username)
		return "server unmute"
	}

	// --------- Self Mute / Unmute ---------
	if newState.SelfMute && !oldState.SelfMute {
		log.Printf("%v was muted\n", newState.Member.User.Username)
		return "self mute"
	}
	if !newState.SelfMute && oldState.SelfMute {
		log.Printf("%v was unmuted\n", newState.Member.User.Username)
		return "self unmute"
	}

	if oldState.ChannelID != "" && newState.ChannelID != "" && oldState.ChannelID != newState.ChannelID {
		log.Printf("%v moved from %v to %v", newState.Member.User.Username, oldChannel.Name, newChannel.Name)
		return "move"
	}

	return ""
}
