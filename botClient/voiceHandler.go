package botClient

import (
	"afho_backend/utils"
	"errors"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type VoiceHandler struct {
	client                *BotClient
	userCurrentVoiceTimes map[string]time.Time
	voiceConnection       *discordgo.VoiceConnection
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
		if userID == v.client.Session.State.User.ID {
			continue
		}
		wg.Add(1)
		go v.updateUserDBTime(userID, &wg, &timeLock)
	}
	wg.Wait()
}

func (v *VoiceHandler) updateUserDBTime(userID string, wg *sync.WaitGroup, timeLock *sync.RWMutex) {
	defer wg.Done()
	stmt, err := v.client.DB.Prepare("INSERT INTO Time_connected (user_id, time_spent) VALUES (?, ?) ON DUPLICATE KEY UPDATE time_spent = time_spent + ?")
	if err != nil {
		utils.Logger.Error(err.Error())
		return
	}
	defer stmt.Close()

	timeLock.RLock()
	currentActiveTime := time.Now().Unix() - v.userCurrentVoiceTimes[userID].Unix()
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
	voiceState, err := v.client.Session.State.VoiceState(v.client.Config.GuildID, v.client.Session.State.User.ID)
	if err != nil || voiceState != nil {
		return errors.New("cannot join channel")
	}

	v.voiceConnection, err = v.client.Session.ChannelVoiceJoin(v.client.Config.GuildID, channelID, false, true)
	if err != nil {
		utils.Logger.Error(err.Error())
		return err
	}
	return nil
}

func HandleJoin(client *BotClient, GuildID string, userID string) (string, error) {
	utils.Logger.Debug("Joining Voice Channel", GuildID, userID)
	voiceState, err := client.Session.State.VoiceState(GuildID, userID)
	if err != nil {
		utils.Logger.Error(err.Error())
		return "Could not join voice channel!", err
	}

	if client.VoiceHandler.voiceConnection != nil {
		return "", errors.New("already have a voice connection")
	}

	voiceconnection, err := client.Session.ChannelVoiceJoin(voiceState.GuildID, voiceState.ChannelID, false, true)
	if err != nil {
		utils.Logger.Error(err.Error())
		return "Could not join voice channel!", err
	}

	client.VoiceHandler.voiceConnection = voiceconnection

	return "Joined Voice Channel", nil
}

func HandleLeave(client *BotClient) string {
	if client.VoiceHandler.voiceConnection == nil {
		return "Not in a Channel"
	}
	err := client.VoiceHandler.voiceConnection.Disconnect()
	if err != nil {
		utils.Logger.Error(err.Error())
	}

	client.VoiceHandler.voiceConnection = nil
	return "Left the channel!"
}
