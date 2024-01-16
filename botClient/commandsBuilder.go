package botClient

import (
	"afho__backend/utils"
	"afho__backend/utils/commands"
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

func (builder *CommandsBuilder) Init(client *BotClient) {
	builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "join",
		Description: "Make the bot join the voice channel you are currently in",
	})
	builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "leave",
		Description: "Make the bot leave the voice channel",
	})
	builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "play",
		Description: "Play music in the voice channel you are in",
		Options: []*discordgo.ApplicationCommandOption{{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "input",
			Description: "Name or URL of the video to play",
			Required:    true,
		}},
	})
	builder.initHandlers(client)
}

func (builder *CommandsBuilder) initHandlers(client *BotClient) {
	builder.Handlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"join": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var returnValue = commands.HandleJoin(s, i.GuildID, i.Member.User.ID)

			utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
				Content: returnValue,
			})
		},
		"leave": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var returnValue = commands.HandleLeave(s, i.GuildID, i.Member.User.ID)

			utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
				Content: returnValue,
			})
		},
		"play": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.ApplicationCommandData().Options[0] == nil {
				utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
					Content: "",
				})
				return
			}
			var input = i.ApplicationCommandData().Options[0].StringValue()

			utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
				Content: client.MusicHandler.Add(client, input, i.Message.Author.Username),
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
