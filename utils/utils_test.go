package utils_test

import (
	"afho_backend/utils"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
)

func TestFormatTime(t *testing.T) {
	seconde := 4 * time.Second
	minute := 2 * time.Minute
	hours := 2 * time.Hour

	if utils.FormatTime(seconde) != "00:04" {
		t.Errorf("Expected 00:04 but got %s", utils.FormatTime(seconde))
	}

	if utils.FormatTime(minute) != "02:00" {
		t.Errorf("Expected 02:00 but got %s", utils.FormatTime(minute))
	}

	if utils.FormatTime(11*minute) != "22:00" {
		t.Errorf("Expected 22:00 but got %s", utils.FormatTime(11*minute))
	}

	if utils.FormatTime(11*minute+seconde) != "22:04" {
		t.Errorf("Expected 22:04 but got %s", utils.FormatTime(11*minute+seconde))
	}

	if utils.FormatTime(11*minute+11*seconde) != "22:44" {
		t.Errorf("Expected 22:44 but got %s", utils.FormatTime(11*minute+11*seconde))
	}

	if utils.FormatTime(hours) != "02:00:00" {
		t.Errorf("Expected 02:00:00 but got %s", utils.FormatTime(hours))
	}

	if utils.FormatTime(11*hours) != "22:00:00" {
		t.Errorf("Expected 22:00:00 but got %s", utils.FormatTime(11*hours))
	}

	if utils.FormatTime(11*hours+seconde) != "22:00:04" {
		t.Errorf("Expected 22:00:04 but got %s", utils.FormatTime(11*hours+seconde))
	}

	if utils.FormatTime(11*hours+11*seconde) != "22:00:44" {
		t.Errorf("Expected 22:00:44 but got %s", utils.FormatTime(11*hours+11*seconde))
	}

	if utils.FormatTime(11*hours+11*minute) != "22:22:00" {
		t.Errorf("Expected 22:22:00 but got %s", utils.FormatTime(11*hours+11*minute))
	}

	if utils.FormatTime(11*hours+11*minute+seconde) != "22:22:04" {
		t.Errorf("Expected 22:22:04 but got %s", utils.FormatTime(11*hours+11*minute+seconde))
	}
}

func TestGetMaxResThumbnail(t *testing.T) {
	thumbnails := []youtube.Thumbnail{
		{
			Width:  100,
			Height: 100,
		},
		{
			Width:  400,
			Height: 200,
		},
		{
			Width:  700,
			Height: 300,
		},
	}

	maxResThumbnail := utils.GetMaxResThumbnail(thumbnails)
	if maxResThumbnail.Width != 700 && maxResThumbnail.Height != 300 {
		t.Errorf("Expected 700x300 but got %dx%d", maxResThumbnail.Width, maxResThumbnail.Height)
	}
}

func TestGetUserDisplayName(t *testing.T) {
	member := &discordgo.Member{
		Nick: "Nick",
		User: &discordgo.User{
			Username: "Username",
		},
	}

	if utils.GetUserDisplayName(member) != "Nick" {
		t.Errorf("Expected 'Nick' but got '%s'", utils.GetUserDisplayName(member))
	}

	member = &discordgo.Member{
		User: &discordgo.User{
			Username: "Username",
		},
	}

	if utils.GetUserDisplayName(member) != "Username" {
		t.Errorf("Expected 'Username' but got '%s'", utils.GetUserDisplayName(member))
	}
}
