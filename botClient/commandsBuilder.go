package botClient

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type CommandsBuilder struct {
	Commands []*discordgo.ApplicationCommand
	Handlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func (builder *CommandsBuilder) AddCommand(command *discordgo.ApplicationCommand) error {
	builder.Commands = append(builder.Commands, command)
	return nil
}

func (builder *CommandsBuilder) Init() {
	builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "join",
		Description: "Make the bot join the voice Channel you are currently in",
	})
	builder.initHandlers()
}

func (builder *CommandsBuilder) initHandlers() {
	builder.Handlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"join": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var voiceState, err = s.State.VoiceState(i.GuildID, i.Member.User.ID)
			if err != nil {
				log.Println(err.Error())
				interactionReply(s, i, &discordgo.InteractionResponseData{
					Content: "Could not join voice channel",
				})
				return
			}

			s.ChannelVoiceJoin(voiceState.GuildID, voiceState.ChannelID, false, true)
			interactionReply(s, i, &discordgo.InteractionResponseData{
				Content: "joined voice channel",
			})
		},
	}
}

func (builder *CommandsBuilder) RegisterCommands(client *BotClient) {
	for index, command := range builder.Commands {
		log.Printf("Adding command '%v'\n", command.Name)
		cmd, err := client.Discord.ApplicationCommandCreate(client.Discord.State.User.ID, client.Config.GuildID, command)
		if err != nil {
			log.Printf("Cannot create '%v' command: %v\n", command.Name, err)
		}
		builder.Commands[index] = cmd
	}
}

func (builder *CommandsBuilder) DeleteCommands(client *BotClient) {
	for _, command := range builder.Commands {
		log.Printf("Deleting command '%v'\n", command.Name)
		err := client.Discord.ApplicationCommandDelete(client.Discord.State.User.ID, client.Config.GuildID, command.ID)
		if err != nil {
			log.Printf("Cannot delete '%v' command: %v\n", command.Name, err)
		}
	}
}

func interactionReply(s *discordgo.Session, i *discordgo.InteractionCreate, message *discordgo.InteractionResponseData) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: message,
	})
}
