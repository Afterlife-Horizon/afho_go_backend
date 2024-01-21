package botClient

import (
	"afho__backend/utils"
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type CacheHandler struct {
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
	c.GuildMutex = sync.RWMutex{}
	c.MembersMutex = sync.RWMutex{}
	c.ChannelsMutex = sync.RWMutex{}
	c.VoiceConnectedMembersMutex = sync.RWMutex{}

	c.UpdateCache(client)
}

func (c *CacheHandler) UpdateCache(client *BotClient) {
	c.updateGuildCache(client)

	var wg = sync.WaitGroup{}
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
	var wg = sync.WaitGroup{}
	wg.Add(3)
	go c.cacheGuild(client, &wg)
	go c.cacheMembers(client, &wg)
	go c.cacheChannels(client, &wg)
	wg.Wait()
}

func (c *CacheHandler) cacheGuild(client *BotClient, wg *sync.WaitGroup) {
	guild, err := client.Session.Guild(client.Config.GuildID)
	if err != nil {
		return
	}
	c.GuildMutex.Lock()
	c.Guild = guild
	c.GuildMutex.Unlock()

	wg.Done()
}

func (c *CacheHandler) cacheMembers(client *BotClient, wg *sync.WaitGroup) {
	members, err := client.Session.GuildMembers(client.Config.GuildID, "", 1000)
	if err != nil {
		return
	}

	var membersCollection = utils.NewCollection[*discordgo.Member](members)

	c.MembersMutex.Lock()
	c.Members = &membersCollection
	c.MembersMutex.Unlock()

	wg.Done()
}

func (c *CacheHandler) cacheChannels(client *BotClient, wg *sync.WaitGroup) {
	channels, err := client.Session.GuildChannels(client.Config.GuildID)
	if err != nil {
		return
	}

	var channelsCollection = utils.NewCollection[*discordgo.Channel](channels)

	c.ChannelsMutex.Lock()
	c.Channels = &channelsCollection
	c.ChannelsMutex.Unlock()

	wg.Done()
}

func (c *CacheHandler) addConnectedMember(member *discordgo.Member) {
	c.VoiceConnectedMembersMutex.Lock()
	var index, err = c.VoiceConnectedMembers.GetIndex(func(m *discordgo.Member) bool { return m.User.ID == member.User.ID })
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
	var voiceConnectedMembers = []*discordgo.Member{}

	c.MembersMutex.RLock()
	for _, member := range c.Members.Data {
		var _, err = client.Session.State.VoiceState(client.Config.GuildID, member.User.ID)
		if err != nil {
			if err != discordgo.ErrStateNotFound {
				log.Println(err.Error())
			}
			continue
		}
		voiceConnectedMembers = append(voiceConnectedMembers, member)
	}
	c.MembersMutex.RUnlock()

	var voiceConnectedMembersCollection = utils.NewCollection[*discordgo.Member](voiceConnectedMembers)
	c.VoiceConnectedMembersMutex.Lock()
	c.VoiceConnectedMembers = &voiceConnectedMembersCollection
	c.VoiceConnectedMembersMutex.Unlock()

	wg.Done()
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
