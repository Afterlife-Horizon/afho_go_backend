package botClient

import (
	"afho__backend/utils"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func JoinHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	utils.Logger.Debug("Recieved join command")
	var returnValue, error = HandleJoin(s, i.GuildID, i.Member.User.ID)
	if error != nil {
		utils.Logger.Error(error.Error())
	}

	utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
		Content: returnValue,
	})
	utils.Logger.Debug("Replied to join command with: ", returnValue)
}

func leaveHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	utils.Logger.Debug("Recieved leave command")
	var returnValue = HandleLeave(s, i.GuildID)

	utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
		Content: returnValue,
	})
	utils.Logger.Debug("Replied to leave command with: ", returnValue)
}

func playHandler(client *BotClient) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		utils.Logger.Debug("Recieved play command")
		if i.ApplicationCommandData().Options[0] == nil {
			utils.Logger.Debug("No input provided")
			utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
				Content: "Please provide a search term or a URL",
			})
			return
		}
		var input = i.ApplicationCommandData().Options[0].StringValue()

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Searching for " + input,
			},
		})

		utils.Logger.Debug("Searching for ", input)

		var returnValue, _ = client.MusicHandler.Add(client, input, i.Member.User.Username, i.Member.User.ID, false)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &returnValue,
		})
		utils.Logger.Debug("Replied to play command with: ", returnValue)
	}
}

func addOnTopHandler(client *BotClient) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		utils.Logger.Debug("Recieved addontop command")
		if client.MusicHandler.Queue.Tracks.Data == nil || len(client.MusicHandler.Queue.Tracks.Data) == 0 {
			utils.Logger.Debug("No queue to add to")
			utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
				Content: "Please use /play instead",
			})
			return
		}

		if i.ApplicationCommandData().Options[0] == nil {
			utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
				Content: "Please provide a search term or a URL",
			})
			return
		}

		var input = i.ApplicationCommandData().Options[0].StringValue()
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Searching for " + input,
			},
		})

		var returnValue, _ = client.MusicHandler.AddOnTop(client, input, i.Member.User.Username, i.Member.User.ID)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &returnValue,
		})
	}
}

func pauseHandler(client *BotClient) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		client.MusicHandler.SetPause(true)
		utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
			Content: "paused",
		})
	}
}

func unpausehandler(client *BotClient) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		client.MusicHandler.SetPause(true)
		utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
			Content: "unpaused",
		})
	}
}

func skipHandler(client *BotClient) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		client.MusicHandler.Skip(client)

		utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
			Content: "skipped",
		})
	}
}

func seekHandler(client *BotClient) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.ApplicationCommandData().Options[0] == nil {
			utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
				Content: "Please provide a timecode",
			})
			return
		}
		var input = i.ApplicationCommandData().Options[0].IntValue()

		var newPosition = time.Duration(input) * time.Second

		client.MusicHandler.Seek(client, newPosition)

		utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("seeked to %vs", newPosition),
		})
	}
}

func shuffleHandler(client *BotClient) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		client.MusicHandler.Shuffle(client)

		utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
			Content: "shuffled",
		})
	}
}

// --------------  misc  -------------- //
func bresilHandler(client *BotClient) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.ApplicationCommandData().Options[0] == nil {
			utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
				Content: "Please provide a user",
			})
			return
		}
		var input = i.ApplicationCommandData().Options[0].UserValue(s)

		var output, _ = client.BrazilUser(i.Member.User, input)

		utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
			Content: output,
		})
	}
}

func addbirthdayHandler(client *BotClient) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "not implemented yet",
			},
		})
	}
}

func addChatSoundHandler(client *BotClient) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "not implemented yet",
			},
		})
	}
}
