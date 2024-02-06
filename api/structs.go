package api

import (
	"afho_backend/utils"
	"strconv"
)

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
