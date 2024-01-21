package botClient

import (
	"afho__backend/utils"
	"afho__backend/utils/commands"
	"fmt"
	"log"
	"time"

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
	err := builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "join",
		Description: "Make the bot join the voice channel you are currently in",
	})
	if err != nil {
		log.Printf("Cannot add '%v' command: %v\n", "join", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "leave",
		Description: "Make the bot leave the voice channel",
	})
	if err != nil {
		log.Printf("Cannot add '%v' command: %v\n", "leave", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "pause",
		Description: "Pause the music",
	})
	if err != nil {
		log.Printf("Cannot add '%v' command: %v\n", "pause", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "unpause",
		Description: "Unpause the music",
	})
	if err != nil {
		log.Printf("Cannot add '%v' command: %v\n", "unpause", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "play",
		Description: "Play music in the voice channel you are in",
		Options: []*discordgo.ApplicationCommandOption{{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "input",
			Description: "Name or URL of the video to play",
			Required:    true,
		}},
	})
	if err != nil {
		log.Printf("Cannot add '%v' command: %v\n", "play", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "addontop",
		Description: "Add a song to the top of the queue",
		Options: []*discordgo.ApplicationCommandOption{{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "input",
			Description: "Name or URL of the video to play",
			Required:    true,
		}},
	})
	if err != nil {
		log.Printf("Cannot add '%v' command: %v\n", "addontop", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "skip",
		Description: "skip the music",
	})
	if err != nil {
		log.Printf("Cannot add '%v' command: %v\n", "skip", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "seek",
		Description: "seek to a timecode in the music",
		Options: []*discordgo.ApplicationCommandOption{{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "timecode",
			Description: "new time code in seconds",
			Required:    true,
		}},
	})
	if err != nil {
		log.Printf("Cannot add '%v' command: %v\n", "seek", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "bresil",
		Description: "Send someone to Brazil",
		Options: []*discordgo.ApplicationCommandOption{{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "the user to send to Brazil",
			Required:    true,
		}},
	})
	if err != nil {
		log.Printf("Cannot add '%v' command: %v\n", "seek", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "shuffle",
		Description: "shuffle the queue",
	})
	if err != nil {
		log.Printf("Cannot add '%v' command: %v\n", "skip", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "addbirthday",
		Description: "add a birthday to the list",
		Options: []*discordgo.ApplicationCommandOption{{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "date",
			Description: "the date of the birthday in the format DD/MM/YYYY",
			Required:    true,
		}},
	})
	if err != nil {
		log.Printf("Cannot add '%v' command: %v\n", "addbirthday", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "addchatsound",
		Description: "add a sound to the list",
		Options: []*discordgo.ApplicationCommandOption{{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "name",
			Description: "the name of the sound",
			Required:    true,
		}, {
			Type:        discordgo.ApplicationCommandOptionAttachment,
			Name:        "sound",
			Description: "the sound file",
			Required:    true,
		}},
	})
	if err != nil {
		log.Printf("Cannot add '%v' command: %v\n", "addchatsound", err)
	}

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

			var returnValue = client.MusicHandler.Add(client, input, i.Member.User.Username, i.Member.User.ID)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &returnValue,
			})
		},
		"addontop": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if client.MusicHandler.Queue.Tracks.Data == nil || len(client.MusicHandler.Queue.Tracks.Data) == 0 {
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

			var returnValue = client.MusicHandler.AddOnTop(client, input, i.Member.User.Username, i.Member.User.ID)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &returnValue,
			})
		},
		"pause": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			client.MusicHandler.SetPause(true)
			utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
				Content: "paused",
			})
		},
		"unpause": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			client.MusicHandler.SetPause(true)
			utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
				Content: "unpaused",
			})
		},
		"skip": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			client.MusicHandler.Skip(client)

			utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
				Content: "skipped",
			})
		},
		"seek": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
		},
		"bresil": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.ApplicationCommandData().Options[0] == nil {
				utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
					Content: "Please provide a user",
				})
				return
			}
			var input = i.ApplicationCommandData().Options[0].UserValue(s)

			var output = client.BrazilUser(i.Member.User, input)

			utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
				Content: output,
			})
		},
		"addbirthday": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		},
		"addchatsound": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		},
		"shuffle": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			client.MusicHandler.Shuffle(client)

			utils.InteractionReply(s, i, &discordgo.InteractionResponseData{
				Content: "shuffled",
			})
		},
	}
}

func (builder *CommandsBuilder) RegisterCommands(client *BotClient) {
	for index, command := range builder.Commands {
		log.Printf("Adding command '%v'\n", command.Name)
		cmd, err := client.Session.ApplicationCommandCreate(client.Session.State.User.ID, client.Config.GuildID, command)
		if err != nil {
			log.Printf("Cannot create '%v' command: %v\n", command.Name, err)
		}
		builder.Commands[index] = cmd
	}
}

func (builder *CommandsBuilder) DeleteCommands(client *BotClient) {
	commands, err := client.Session.ApplicationCommands(client.Session.State.User.ID, client.Config.GuildID)
	if err != nil {
		log.Fatalf("Cannot get commands: %v\n", err)
	}
	for _, command := range commands {
		log.Printf("Deleting command '%v'\n", command.Name)

		err := client.Session.ApplicationCommandDelete(client.Session.State.User.ID, client.Config.GuildID, command.ID)
		if err != nil {
			log.Printf("Cannot delete '%v' command: %v\n", command.Name, err)
		}
	}
}
