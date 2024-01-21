package utils

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func InteractionReply(s *discordgo.Session, i *discordgo.InteractionCreate, message *discordgo.InteractionResponseData) {
	// fmt.Println("Replying to interaction", i.Interaction.ID, "with message", message.Content)
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: message,
	}); err != nil {
		fmt.Println("Error while replying to interaction", i.Interaction.ID, "with message", message.Content)
		fmt.Println(err.Error())
	}
}
