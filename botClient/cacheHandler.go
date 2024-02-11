package botClient

import (
	"afho_backend/utils"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type CacheHandler struct {
	discordBot            *BotClient
	Guild                 *discordgo.Guild
	Members               *utils.Collection[*discordgo.Member]
	Channels              *utils.Collection[*discordgo.Channel]
	VoiceConnectedMembers *utils.Collection[*discordgo.Member]

	GuildMutex                 sync.RWMutex
	MembersMutex               sync.RWMutex
	ChannelsMutex              sync.RWMutex
	VoiceConnectedMembersMutex sync.RWMutex
}

func (c *CacheHandler) Init(client *BotClient) {
	utils.Logger.Debug("Initialising Cache Handler")
	c.discordBot = client
	c.GuildMutex = sync.RWMutex{}
	c.MembersMutex = sync.RWMutex{}
	c.ChannelsMutex = sync.RWMutex{}
	c.VoiceConnectedMembersMutex = sync.RWMutex{}

	c.UpdateCache()
}

func (c *CacheHandler) UpdateCache() {
	utils.Logger.Debug("Updating Cache")
	client := c.discordBot
	c.updateGuildCache(client)

	wg := sync.WaitGroup{}
	wg.Add(6)
	go c.updateConnectedMembers(client, &wg)
	go c.updateAchievements(client, &wg)
	go c.updateFavorites(client, &wg)
	go c.updateVoiceTimes(client, &wg)
	go c.updateUserXps(client, &wg)
	go c.updateSoundPaths(client, &wg)

	wg.Wait()
}

func (c *CacheHandler) updateGuildCache(client *BotClient) {
	utils.Logger.Debug("Updating Guild Cache")
	wg := sync.WaitGroup{}
	wg.Add(3)
	go c.cacheGuild(client, &wg)
	go c.cacheMembers(client, &wg)
	go c.cacheChannels(client, &wg)
	wg.Wait()
	utils.Logger.Debug("Guild Cache Updated")
}

func (c *CacheHandler) cacheGuild(client *BotClient, wg *sync.WaitGroup) {
	guild, err := client.Session.Guild(client.Config.GuildID)
	if err != nil {
		return
	}
	c.GuildMutex.Lock()
	c.Guild = guild
	c.GuildMutex.Unlock()

	utils.Logger.Debug("Guild Cached")
	wg.Done()
}

func (c *CacheHandler) cacheMembers(client *BotClient, wg *sync.WaitGroup) {
	members, err := client.Session.GuildMembers(client.Config.GuildID, "", 1000)
	if err != nil {
		return
	}

	membersCollection := utils.NewCollection[*discordgo.Member](members)

	c.MembersMutex.Lock()
	c.Members = &membersCollection
	c.MembersMutex.Unlock()

	utils.Logger.Debug("Members Cached")
	wg.Done()
}

func (c *CacheHandler) cacheChannels(client *BotClient, wg *sync.WaitGroup) {
	channels, err := client.Session.GuildChannels(client.Config.GuildID)
	if err != nil {
		return
	}

	channelsCollection := utils.NewCollection[*discordgo.Channel](channels)

	c.ChannelsMutex.Lock()
	c.Channels = &channelsCollection
	c.ChannelsMutex.Unlock()

	utils.Logger.Debug("Channels Cached")
	wg.Done()
}

func (c *CacheHandler) addConnectedMember(member *discordgo.Member) {
	c.VoiceConnectedMembersMutex.Lock()
	index, err := c.VoiceConnectedMembers.GetIndex(func(m *discordgo.Member) bool { return m.User.ID == member.User.ID })
	if err != nil {
		c.VoiceConnectedMembers.Insert(member)
	} else {
		c.VoiceConnectedMembers.Update(index, member)
	}
	c.VoiceConnectedMembersMutex.Unlock()
}

func (c *CacheHandler) removeConnectedMember(member *discordgo.Member) {
	c.VoiceConnectedMembersMutex.Lock()
	c.VoiceConnectedMembers.RemoveItem(func(m *discordgo.Member) bool {
		return m.User.ID == member.User.ID
	})
	c.VoiceConnectedMembersMutex.Unlock()
}

func (c *CacheHandler) updateConnectedMembers(client *BotClient, wg *sync.WaitGroup) {
	utils.Logger.Debug("Updating Connected Members")
	voiceConnectedMembers := []*discordgo.Member{}

	c.MembersMutex.RLock()
	for _, member := range c.Members.Data {
		if member.User.ID == client.Session.State.User.ID {
			continue
		}
		_, err := client.Session.State.VoiceState(client.Config.GuildID, member.User.ID)
		if err != nil {
			if err != discordgo.ErrStateNotFound {
				utils.Logger.Error(err.Error())
			}
			continue
		}
		voiceConnectedMembers = append(voiceConnectedMembers, member)
	}
	c.MembersMutex.RUnlock()

	voiceConnectedMembersCollection := utils.NewCollection[*discordgo.Member](voiceConnectedMembers)
	c.VoiceConnectedMembersMutex.Lock()
	c.VoiceConnectedMembers = &voiceConnectedMembersCollection
	c.VoiceConnectedMembersMutex.Unlock()

	utils.Logger.Debug("Connected Members Updated")

	if wg != nil {
		wg.Done()
	}
}

func (c *CacheHandler) updateAchievements(client *BotClient, wg *sync.WaitGroup) {
	wg.Done()
}

func (c *CacheHandler) updateFavorites(client *BotClient, wg *sync.WaitGroup) {
	wg.Done()
}

func (c *CacheHandler) updateVoiceTimes(client *BotClient, wg *sync.WaitGroup) {
	wg.Done()
}

func (c *CacheHandler) updateUserXps(client *BotClient, wg *sync.WaitGroup) {
	wg.Done()
}

func (c *CacheHandler) updateSoundPaths(client *BotClient, wg *sync.WaitGroup) {
	wg.Done()
}

func (c *CacheHandler) UpdateDB() {
	utils.Logger.Debug("Updating DB")
	c.updateDBUsers()
}

func (c *CacheHandler) updateDBUsers() {
	utils.Logger.Debug("Updating DB Users")
	c.MembersMutex.RLock()
	stmt, err := c.discordBot.DB.Prepare("INSERT INTO Users (id, username, nickname, avatar, roles) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE username = ?, nickname = ?, avatar = ?, roles = ?")
	defer stmt.Close()
	if err != nil {
		utils.Logger.Error(err.Error())
		return
	}

	for _, member := range c.Members.Data {
		var roles string = ""
		for _, role := range member.Roles {
			roles += role + ","
		}
		roles = roles[:len(roles)-1]

		utils.Logger.Debug("Updating DB User", member.User.Username)
		_, err := stmt.Exec(member.User.ID, member.User.Username, member.Nick, member.User.Avatar, roles, member.User.Username, member.Nick, member.User.Avatar, roles)
		if err != nil {
			utils.Logger.Error(err.Error())
		}
	}
	c.MembersMutex.RUnlock()

	utils.Logger.Debug("DB Users Updated")
}
