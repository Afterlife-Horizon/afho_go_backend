package utils

import "github.com/bwmarrin/discordgo"

func InteractionReply(s *discordgo.Session, i *discordgo.InteractionCreate, message *discordgo.InteractionResponseData) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: message,
	})
}
