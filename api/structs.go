package api

import (
	"afho_backend/utils"
	"strconv"
)

type DbUser struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	NickName    string   `json:"nickname"`
	Avatar      string   `json:"avatar"`
	Roles       []string `json:"roles"`
	LodeStoneId string   `json:"lodestone_id"`
}

type BrasilBoard struct {
	User           User `json:"user"`
	BresilReceived int  `json:"bresil_received"`
	BresilSent     int  `json:"bresil_sent"`
}

type User struct {
	Username         string `json:"displayName"`
	User_id          string `json:"user_id"`
	DisplayAvatarURL string `json:"displayAvatarURL"`
}

// accent_color int
// avatar string
// avatar_decoration_data	DiscordApiAvatarDecorationData
// banner string
// banner_color string
// clan string
// discriminator string
// email string
// flags int
// global_name string
// id string
// locale string
// premium_type int
// primary_guild string
// public_flags int
// username string
// verified bool

type DiscordApiUser struct {
	Id                     string   `json:"id"`
	Username               string   `json:"username"`
	Discriminator          string   `json:"discriminator"`
	Global_name            *string  `json:"global_name"`
	Avatar                 *string  `json:"avatar"`
	System                 *bool    `json:"system"`
	MfaEnabled             *bool    `json:"mfa_enabled"`
	Banner                 *string  `json:"banner"`
	Accent_color           *float64 `json:"accent_color"`
	Locale                 *string  `json:"locale"`
	Verified               *bool    `json:"verified"`
	Flags                  *float64 `json:"flags"`
	PremiumType            *float64 `json:"premium_type"`
	PublicFlags            *float64 `json:"public_flags"`
	Avatar_decoration_data *DiscordApiAvatarDecorationData
}

func DiscordUserFromMap(m map[string]interface{}) *DiscordApiUser {

	// Setup nullable fields
	var global_name *string
	if m["global_name"] != nil {
		value := m["global_name"].(string)
		global_name = &value
	}

	var avatar *string
	if m["avatar"] != nil {
		value := m["avatar"].(string)
		avatar = &value
	}

	var system *bool
	if m["system"] != nil {
		value := m["system"].(bool)
		system = &value
	}

	var mfaEnabled *bool
	if m["mfa_enabled"] != nil {
		value := m["mfa_enabled"].(bool)
		mfaEnabled = &value
	}

	var banner *string
	if m["banner"] != nil {
		value := m["banner"].(string)
		banner = &value
	}

	var accent_color *float64
	if m["accent_color"] != nil {
		value := m["accent_color"].(float64)
		accent_color = &value
	}

	var locale *string
	if m["locale"] != nil {
		value := m["locale"].(string)
		locale = &value
	}

	var verified *bool
	if m["verified"] != nil {
		value := m["verified"].(bool)
		verified = &value
	}

	var flags *float64
	if m["flags"] != nil {
		value := m["flags"].(float64)
		flags = &value
	}

	var premium_type *float64
	if m["premium_type"] != nil {
		value := m["premium_type"].(float64)
		premium_type = &value
	}

	var public_flags *float64
	if m["public_flags"] != nil {
		value := m["public_flags"].(float64)
		public_flags = &value
	}

	var avatar_decoration_data *DiscordApiAvatarDecorationData
	if m["avatar_decoration_data"] != nil {
		avatar_decoration_data = &DiscordApiAvatarDecorationData{
			Asset:  m["avatar_decoration_data"].(map[string]interface{})["asset"].(string),
			Sku_id: m["avatar_decoration_data"].(map[string]interface{})["sku_id"].(string),
		}
	}

	return &DiscordApiUser{
		Id:                     m["id"].(string),
		Username:               m["username"].(string),
		Discriminator:          m["discriminator"].(string),
		Global_name:            global_name,
		Avatar:                 avatar,
		System:                 system,
		MfaEnabled:             mfaEnabled,
		Banner:                 banner,
		Accent_color:           accent_color,
		Locale:                 locale,
		Verified:               verified,
		Flags:                  flags,
		PremiumType:            premium_type,
		PublicFlags:            public_flags,
		Avatar_decoration_data: avatar_decoration_data,
	}
}

type DiscordApiAvatarDecorationData struct {
	Asset  string `json:"asset"`
	Sku_id string `json:"sku_id"`
}

type Effects struct {
	Speed       int  `json:"speed"`
	Bassboost   int  `json:"bassboost"`
	Subboost    bool `json:"subboost"`
	Mcompand    bool `json:"mcompand"`
	Haas        bool `json:"haas"`
	Gate        bool `json:"gate"`
	Karaoke     bool `json:"karaoke"`
	Flanger     bool `json:"flanger"`
	Pulsator    bool `json:"pulsator"`
	Surrounding bool `json:"surrounding"`
	ThreeD      bool `json:"3d"`
	Vaporwave   bool `json:"vaporwave"`
	Nightcore   bool `json:"nightcore"`
	Phaser      bool `json:"phaser"`
	Normalizer  bool `json:"normalizer"`
	Tremolo     bool `json:"tremolo"`
	Vibrato     bool `json:"vibrato"`
	Reverse     bool `json:"reverse"`
	Treble      bool `json:"treble"`
}

func (effects *Effects) ToFilters() string {
	var filters string

	filters += "bass=g=" + strconv.Itoa(effects.Bassboost)
	if effects.Speed <= 0 {
		effects.Speed = 1
	}
	filters += ",atempo=" + strconv.Itoa(effects.Speed)

	if effects.Normalizer {
		filters += ",dynaudnorm=f=200"
	}
	if effects.Subboost {
		filters += ",asubboost"
	}
	if effects.Nightcore {
		filters += ",asetrate=48000*1.25"
	}
	if effects.Phaser {
		filters += ",aphaser"
	}
	if effects.Reverse {
		filters += ",areverse"
	}
	if effects.Surrounding {
		filters += ",surround"
	}
	if effects.ThreeD {
		filters += ",apulsator=hz=0.125"
	}
	if effects.Tremolo {
		filters += ",tremolo"
	}
	if effects.Vibrato {
		filters += ",vibrato=f=6.5"
	}
	if effects.Gate {
		filters += ",agate"
	}
	if effects.Karaoke {
		filters += ",stereotools=mlev=0.03"
	}
	if effects.Vaporwave {
		filters += ",aresample=48000,asetrate=48000*0.8"
	}
	if effects.Flanger {
		filters += ",flanger"
	}
	if effects.Treble {
		filters += ",treble=g=5"
	}
	if effects.Haas {
		filters += ",haas"
	}
	if effects.Mcompand {
		filters += ",mcompand"
	}

	utils.Logger.Debug("Filters:", filters)
	return filters
}

type Track struct {
	Id                string `json:"id"`
	Title             string `json:"title"`
	DurationFormatted string `json:"durationFormatted"`
	Requester         string `json:"requester"`
	Author            string `json:"author"`
	Thumbnail         string `json:"thumbnail"`
	Duration          int    `json:"duration"`
}

type Queue struct {
	Tracks  []Track `json:"tracks"`
	Effects Effects `json:"effects"`
	Paused  bool    `json:"paused"`
}

type Level struct {
	User User `json:"user"`
	Xp   int  `json:"xp"`
	Lvl  int  `json:"lvl"`
}

type FetchResults struct {
	Formatedprog string   `json:"formatedprog"`
	Admins       []string `json:"admins"`
	Queue        []Queue  `json:"queue"`
	Prog         int      `json:"prog"`
}

type connectedMembers struct {
	Username string `json:"username"`
	ID       string `json:"id"`
}

type ConnectedMembersResponse struct {
	Data []connectedMembers `json:"data"`
}

type Time struct {
	User      User `json:"user"`
	TimeSpent int  `json:"time_spent"`
}

type APIAchievement struct {
	Id           string        `json:"id"`
	Username     string        `json:"username"`
	Achievements []Achievement `json:"achievements"`
	Counter      int8
}

type Achievement struct {
	Name         string `json:"name"`
	Requirements string `json:"requirements"`
	Type         string `json:"type"`
}

type Fav struct {
	Id        string `json:"id"`
	User_id   string `json:"user_id"`
	Name      string `json:"name"`
	Url       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Type      string `json:"type"`
	DateAdded string `json:"date_added"`
}
