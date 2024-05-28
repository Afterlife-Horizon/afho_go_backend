package test_constants

import (
	"afho_backend/utils"

	"github.com/bwmarrin/discordgo"
)

func GetMockMembers() *utils.Collection[*discordgo.Member] {
	members := utils.NewCollection([]*discordgo.Member{})
	members.Insert(
		&discordgo.Member{
			Roles: []string{"testRole", "testRole2"},
			User: &discordgo.User{
				ID:       "testID1",
				Username: "TestUserName1",
				Avatar:   "testAvatarString",
			},
		}, &discordgo.Member{
			Roles: []string{"testRole", "testRole2", "testRoleAdmin"},
			User: &discordgo.User{
				ID:       "testID2",
				Username: "TestUserName2",
				Avatar:   "a_testAvatarString",
			},
		}, &discordgo.Member{
			Roles: []string{},
			User: &discordgo.User{
				ID:       "testID3",
				Username: "TestUserName3",
				Avatar:   "a_testAvatarString",
			},
		},
	)
	return &members
}
