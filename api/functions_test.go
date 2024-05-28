package api_test

import (
	"afho_backend/api"
	test_constants "afho_backend/test"
	"afho_backend/utils"
	"os"
	"reflect"
	"testing"
)

func TestMain(m *testing.M) {
	utils.InitLogger(false)
	code := m.Run()
	os.Exit(code)
}

func TestGetUserAvatar(t *testing.T) {
	members := test_constants.GetMockMembers()

	// test for non gif version
	if api.GetUserAvatar(members, "testID1") != "https://cdn.discordapp.com/avatars/testID1/testAvatarString.png" {
		t.Errorf("Avatar value was incorrect (png version)")
	}

	// test for gif version
	if api.GetUserAvatar(members, "testID2") != "https://cdn.discordapp.com/avatars/testID2/a_testAvatarString.gif" {
		t.Errorf("Avatar value was incorrect (gif version)")
	}

	// test incorrect id
	if api.GetUserAvatar(members, "incorrectID") != "" {
		t.Errorf("Avatar value was incorrect")
	}
}

func TestGetAdmins(t *testing.T) {
	members := test_constants.GetMockMembers()
	if !reflect.DeepEqual(api.GetAdmins(members, "testRoleAdmin"), []string{"TestUserName2"}) {
		t.Error("Admins list was incorrect")
	}
}
