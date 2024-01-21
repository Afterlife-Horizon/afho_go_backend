package api

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
	Prog         float64  `json:"prog"`
	Queue        []Queue  `json:"queue"`
}

type connectedMembers struct {
	Username string `json:"username"`
}

type ConnectedMembersResponse struct {
	Data []connectedMembers `json:"data"`
}

type Time struct {
	User      User `json:"user"`
	TimeSpent int  `json:"time_spent"`
}
