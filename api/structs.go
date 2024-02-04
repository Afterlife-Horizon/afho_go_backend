package api

import (
	"fmt"
	"strconv"
)

type BrasilBoard struct {
	User           User `json:"user"`
	BresilReceived int  `json:"bresil_received"`
	BresilSent     int  `json:"bresil_sent"`
}

type User struct {
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

	fmt.Println(filters)
	return filters
}

type Track struct {
	Id                string `json:"id"`
	Title             string `json:"title"`
	Duration          int    `json:"duration"`
	DurationFormatted string `json:"durationFormatted"`
	Requester         string `json:"requester"`
	Author            string `json:"author"`
	Thumbnail         string `json:"thumbnail"`
}

type Queue struct {
	Effects Effects `json:"effects"`
	Paused  bool    `json:"paused"`
	Tracks  []Track `json:"tracks"`
}

type Level struct {
	User User `json:"user"`
	Xp   int  `json:"xp"`
	Lvl  int  `json:"lvl"`
}

type FetchResults struct {
	Admins       []string `json:"admins"`
	Formatedprog string   `json:"formatedprog"`
	Prog         int      `json:"prog"`
	Queue        []Queue  `json:"queue"`
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
	Counter      int8
	Id           string        `json:"id"`
	Username     string        `json:"username"`
	Achievements []Achievement `json:"achievements"`
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
