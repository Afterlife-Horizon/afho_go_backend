package botClient

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func MessageCreate(client *BotClient) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		stmt, err := client.DB.Prepare("INSERT INTO Levels (user_id, xp) VALUES (?, ?) ON DUPLICATE KEY UPDATE xp = xp + 1")
		if err != nil {
			log.Println(err.Error())
		}

		_, err = stmt.Exec(m.Author.ID, 1)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

const (
	join = iota
	leave
	self_deafen
	self_undeafen
	self_mute
	self_unmute
	server_deafen
	server_undeafen
	server_mute
	server_unmute
	move
	bresil
	none
)

func VoiceStateUpdate(client *BotClient) func(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	return func(s *discordgo.Session, state *discordgo.VoiceStateUpdate) {
		switch getStateChangeType(client, state) {
		case join:
			client.CacheHandler.addConnectedMember(state.Member)
		case leave:
			client.CacheHandler.removeConnectedMember(state.Member)
		default:

		}
	}
}

func getStateChangeType(client *BotClient, newState *discordgo.VoiceStateUpdate) int {
	if newState.Member.User.ID == client.Session.State.User.ID {
		return -1
	}
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
		return join
	}
	if newState.ChannelID == "" && oldState.ChannelID != "" {
		log.Printf("%v left %v\n", newState.Member.User.Username, oldChannel.Name)
		return leave
	}

	// --------- Server deafen / undeafened ---------
	if newState.Deaf && !oldState.Deaf {
		log.Printf("%v was deafened\n", newState.Member.User.Username)
		return server_deafen
	}
	if !newState.Deaf && oldState.Deaf {
		log.Printf("%v was undeafened\n", newState.Member.User.Username)
		return server_undeafen
	}

	// --------- Self deafen / undeafened ---------
	if newState.SelfDeaf && !oldState.SelfDeaf {
		log.Printf("%v was deafened\n", newState.Member.User.Username)
		return self_deafen
	}
	if !newState.SelfDeaf && oldState.SelfDeaf {
		log.Printf("%v was undeafened\n", newState.Member.User.Username)
		return self_undeafen
	}

	// --------- Server Mute / Unmute ---------
	if newState.Mute && !oldState.Mute {
		log.Printf("%v was muted\n", newState.Member.User.Username)
		return server_mute
	}
	if !newState.Mute && oldState.Mute {
		log.Printf("%v was unmuted\n", newState.Member.User.Username)
		return server_unmute
	}

	// --------- Self Mute / Unmute ---------
	if newState.SelfMute && !oldState.SelfMute {
		log.Printf("%v was muted\n", newState.Member.User.Username)
		return self_unmute
	}
	if !newState.SelfMute && oldState.SelfMute {
		log.Printf("%v was unmuted\n", newState.Member.User.Username)
		return self_mute
	}

	// --------- Move ---------
	if oldState.ChannelID != "" && newState.ChannelID != "" && oldState.ChannelID != newState.ChannelID {
		// if newState.ChannelID == client.Config.BrasilChannelID {
		// 	// check if user was sent or moved on their own
		// 	// var test, err = client.Session.GuildAuditLog(client.Config.GuildID, newState.UserID, "", 26, 0)
		// 	// if err != nil {
		// 	// 	log.Println(err.Error())
		// 	// 	return none
		// 	// }

		// 	return bresil
		// }

		log.Printf("%v moved from %v to %v", newState.Member.User.Username, oldChannel.Name, newChannel.Name)
		return move
	}

	return none
}

func InteractionCreate(client *BotClient) func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if h, ok := client.CommandsBuilder.Handlers[interaction.ApplicationCommandData().Name]; ok {
			h(session, interaction)
		}
	}
}
