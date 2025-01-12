package api_test

import (
	"afho_backend/api"
	test_constants "afho_backend/test"
	"afho_backend/utils"
	"database/sql"
	"os"
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

func TestMain(m *testing.M) {
	_ = godotenv.Load("../.env.test")
	utils.InitLogger(false)

	env := test_constants.MockEnv()

	cfg := mysql.Config{
		User:                 env.DbUser,
		Passwd:               env.DbPass,
		Net:                  "tcp",
		Addr:                 env.DbAddress,
		DBName:               env.DbName,
		AllowNativePasswords: true,
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		utils.Logger.Fatal(err.Error())
	}

	code := m.Run()
	os.Exit(code)
}

func TestGetUserAvatar(t *testing.T) {
	members := test_constants.GetFakeMockMembers()

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
	members := test_constants.GetFakeMockMembers()
	if !reflect.DeepEqual(api.GetAdmins(members, "testRoleAdmin"), []string{"TestUserName2"}) {
		t.Error("Admins list was incorrect")
	}
}

func TestGetLevelsDb(t *testing.T) {
	members := test_constants.GetMockMembers(t)

	levels := api.GetLevelsDb(db, members)
	if levels == nil {
		t.Error("Levels db was nil")
	}

	dbUsers := test_constants.GetMockDBUsers(t)

	DbLevels, err := db.Query("SELECT * FROM Levels")
	if err != nil {
		t.Error("Error while querying levels table")
	}

	for DbLevels.Next() {
		var levelDB api.Level = api.Level{}
		levelDB.User = api.User{}

		var user_id string
		var xp int
		err := DbLevels.Scan(&user_id, &xp)
		if err != nil {
			t.Error("Error while scanning levels table")
		}

		levelDB.User.User_id = user_id
		levelDB.Xp = xp
		levelDB.Lvl = int(api.XptoLvl(xp))

		member, err := members.Get(func(i *discordgo.Member) bool {
			return i.User.ID == user_id
		})

		for _, dbUser := range dbUsers {
			if dbUser.ID == levelDB.User.User_id {
				levelDB.User.Username = utils.GetUserDisplayName(member)
				levelDB.User.DisplayAvatarURL = api.GetUserAvatar(members, dbUser.ID)
			}
		}

		var level api.Level
		for _, tmplevel := range levels {
			if tmplevel.User.User_id == levelDB.User.User_id {
				level = tmplevel
			}
		}

		if !reflect.DeepEqual(level, levelDB) {
			t.Error("Levels db was incorrect")
		}
	}

}

func TestGetBrasilBoardDB(t *testing.T) {
	members := test_constants.GetMockMembers(t)

	BrasilBoard := api.GetBrasilBoardDB(db, members)
	if BrasilBoard == nil {
		t.Error("Levels db was nil")
	}

	dbUsers := test_constants.GetMockDBUsers(t)

	DbBrasilBoard, err := db.Query("SELECT * FROM Bresil_count")
	if err != nil {
		t.Error("Error while querying levels table")
	}

	var DBBrasilBoard []api.BrasilBoard
	for DbBrasilBoard.Next() {
		var levelDB api.BrasilBoard = api.BrasilBoard{}
		levelDB.User = api.User{}

		err := DbBrasilBoard.Scan(&levelDB.User.User_id, &levelDB.BresilReceived, &levelDB.BresilSent)
		if err != nil {
			t.Error("Error while scanning levels table")
		}

		member, err := members.Get(func(i *discordgo.Member) bool {
			return i.User.ID == levelDB.User.User_id
		})

		for _, dbUser := range dbUsers {
			if dbUser.ID == levelDB.User.User_id {
				levelDB.User.Username = utils.GetUserDisplayName(member)
				levelDB.User.DisplayAvatarURL = api.GetUserAvatar(members, dbUser.ID)
			}
		}

		DBBrasilBoard = append(DBBrasilBoard, levelDB)
	}

	if !reflect.DeepEqual(BrasilBoard, DBBrasilBoard) {
		t.Error("Levels db was incorrect")
	}
}
