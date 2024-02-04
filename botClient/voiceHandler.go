package botClient

import (
	"afho__backend/utils"
	"errors"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type VoiceHandler struct {
	client                *BotClient
	userCurrentVoiceTimes map[string]time.Time
}

func (v *VoiceHandler) Init(client *BotClient) {
	v.client = client
	v.userCurrentVoiceTimes = make(map[string]time.Time, 100)
	utils.Logger.Debug("Initialising Voice Handler")
	v.client.CacheHandler.updateConnectedMembers(v.client, nil)
	for _, voiceState := range v.client.CacheHandler.VoiceConnectedMembers.Data {
		v.StartUserTime(voiceState.User.ID)
	}
}

func (v *VoiceHandler) StartUserTime(userID string) {
	utils.Logger.Debug("Starting time for user", userID)
	v.userCurrentVoiceTimes[userID] = time.Now()
}

func (v *VoiceHandler) UpdateDBTime() {
	utils.Logger.Debug("Updating time for all users")
	wg := sync.WaitGroup{}
	timeLock := sync.RWMutex{}

	for userID := range v.userCurrentVoiceTimes {
		wg.Add(1)
		go v.updateUserDBTime(userID, &wg, &timeLock)
	}
	wg.Wait()
}

func (v *VoiceHandler) updateUserDBTime(userID string, wg *sync.WaitGroup, timeLock *sync.RWMutex) {
	defer wg.Done()
	var stmt, err = v.client.DB.Prepare("INSERT INTO Time_connected (user_id, time_spent) VALUES (?, ?) ON DUPLICATE KEY UPDATE time_spent = time_spent + ?")
	defer stmt.Close()
	if err != nil {
		utils.Logger.Error(err.Error())
		return
	}

	timeLock.RLock()
	var currentActiveTime = time.Now().Unix() - v.userCurrentVoiceTimes[userID].Unix()
	timeLock.RUnlock()
	utils.Logger.Debugf("Updating time for user %v, adding %vs", userID, currentActiveTime)
	_, err = stmt.Exec(userID, currentActiveTime, currentActiveTime)
	if err != nil {
		utils.Logger.Error(err.Error())
		return
	}

	timeLock.Lock()
	v.userCurrentVoiceTimes[userID] = time.Now()
	timeLock.Unlock()
}

func (v *VoiceHandler) JoinVoiceChannel(channelID string) error {
	var voiceState, err = v.client.Session.State.VoiceState(v.client.Config.GuildID, v.client.Session.State.User.ID)
	if err != nil || voiceState != nil {
		return errors.New("cannot join channel")
	}

	v.client.Session.ChannelVoiceJoin(v.client.Config.GuildID, channelID, false, true)
	return nil
}

func HandleJoin(s *discordgo.Session, GuildID string, userID string) (string, error) {
	utils.Logger.Debug("Joining Voice Channel", GuildID, userID)
	var voiceState, err = s.State.VoiceState(GuildID, userID)
	if err != nil {
		utils.Logger.Error(err.Error())
		return "Could not join voice channel!", err
	}

	s.ChannelVoiceJoin(voiceState.GuildID, voiceState.ChannelID, false, true)
	return "Joined Voice Channel", nil
}

func HandleLeave(s *discordgo.Session, GuildID string) string {
	var voiceState = s.VoiceConnections[GuildID]

	if voiceState == nil {
		return "Not in a voice channel!"
	}

	voiceState.Disconnect()
	return "Left the channel!"
}
