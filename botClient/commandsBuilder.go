package botClient

import (
	"afho_backend/utils"

	"github.com/bwmarrin/discordgo"
)

type CommandsBuilder struct {
	Commands []*discordgo.ApplicationCommand
	Handlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func (builder *CommandsBuilder) AddCommand(command *discordgo.ApplicationCommand) error {
	utils.Logger.Debugf("Adding Command '%v'\n", command.Name)
	builder.Commands = append(builder.Commands, command)
	return nil
}

func (builder *CommandsBuilder) Init(client *BotClient) {
	utils.Logger.Debug("Initialising Commands Builder")

	utils.Logger.Debug("Adding Commands")
	err := builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "join",
		Description: "Make the bot join the voice channel you are currently in",
	})
	if err != nil {
		utils.Logger.Errorf("Cannot add '%v' command: %v\n", "join", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "leave",
		Description: "Make the bot leave the voice channel",
	})
	if err != nil {
		utils.Logger.Errorf("Cannot add '%v' command: %v\n", "leave", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "pause",
		Description: "Pause the music",
	})
	if err != nil {
		utils.Logger.Errorf("Cannot add '%v' command: %v\n", "pause", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "unpause",
		Description: "Unpause the music",
	})
	if err != nil {
		utils.Logger.Errorf("Cannot add '%v' command: %v\n", "unpause", err)
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
		utils.Logger.Errorf("Cannot add '%v' command: %v\n", "play", err)
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
		utils.Logger.Errorf("Cannot add '%v' command: %v\n", "addontop", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "skip",
		Description: "skip the music",
	})
	if err != nil {
		utils.Logger.Errorf("Cannot add '%v' command: %v\n", "skip", err)
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
		utils.Logger.Errorf("Cannot add '%v' command: %v\n", "seek", err)
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
		utils.Logger.Errorf("Cannot add '%v' command: %v\n", "seek", err)
	}

	// ---------------------------- //
	err = builder.AddCommand(&discordgo.ApplicationCommand{
		Name:        "shuffle",
		Description: "shuffle the queue",
	})
	if err != nil {
		utils.Logger.Errorf("Cannot add '%v' command: %v\n", "skip", err)
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
		utils.Logger.Errorf("Cannot add '%v' command: %v\n", "addbirthday", err)
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
		utils.Logger.Errorf("Cannot add '%v' command: %v\n", "addchatsound", err)
	}

	builder.initHandlers(client)
}

func (builder *CommandsBuilder) initHandlers(client *BotClient) {
	utils.Logger.Debug("Initialising Command Handlers")
	builder.Handlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"join":         JoinHandler,
		"leave":        leaveHandler,
		"play":         playHandler(client),
		"addontop":     addOnTopHandler(client),
		"pause":        pauseHandler(client),
		"unpause":      unpausehandler(client),
		"skip":         skipHandler(client),
		"seek":         seekHandler(client),
		"bresil":       bresilHandler(client),
		"addbirthday":  addbirthdayHandler(client),
		"addchatsound": addChatSoundHandler(client),
		"shuffle":      shuffleHandler(client),
	}
}

func (builder *CommandsBuilder) RegisterCommands(client *BotClient) {
	utils.Logger.Debug("Registering Commands")
	for index, command := range builder.Commands {
		utils.Logger.Infof("Adding command '%v'\n", command.Name)
		cmd, err := client.Session.ApplicationCommandCreate(client.Session.State.User.ID, client.Config.GuildID, command)
		if err != nil {
			utils.Logger.Infof("Cannot create '%v' command: %v\n", command.Name, err)
		}
		builder.Commands[index] = cmd
	}
}

func (builder *CommandsBuilder) DeleteCommands(client *BotClient) {
	utils.Logger.Debug("Deleting Commands")
	commands, err := client.Session.ApplicationCommands(client.Session.State.User.ID, client.Config.GuildID)
	if err != nil {
		utils.Logger.Fatalf("Cannot get commands: %v\n", err)
	}
	for _, command := range commands {
		utils.Logger.Infof("Deleting command '%v'\n", command.Name)

		err := client.Session.ApplicationCommandDelete(client.Session.State.User.ID, client.Config.GuildID, command.ID)
		if err != nil {
			utils.Logger.Errorf("Cannot delete '%v' command: %v\n", command.Name, err)
		}
	}
}
