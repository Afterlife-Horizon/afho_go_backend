package test_constants

import (
	"afho_backend/api"
	"afho_backend/utils"
	"database/sql"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

func setup() {
	_ = godotenv.Load("../.env.test")
	utils.InitLogger(false)

	env := MockEnv()

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
}

func MockEnv() utils.Env {
	falsep := false
	env := utils.LoadEnv(utils.Flags{
		AddCommands: &falsep,
		DelCommands: &falsep,
	})

	mockDBName, ok := os.LookupEnv("MockDBName")
	if !ok || mockDBName == "" {
		log.Println("MockDBName not found in environment variables, setting to afho_mock")
		mockDBName = "afho_mock"
	}
	env.DbName = mockDBName

	return env
}

func GetFakeMockMembers() *utils.Collection[*discordgo.Member] {
	members := utils.NewCollection([]*discordgo.Member{})
	members.Insert(
		&discordgo.Member{
			Roles: []string{"testRole", "testRole2"},
			User: &discordgo.User{
				ID:       "testID1",
				Username: "TestUserName1",
				Avatar:   "testAvatarString",
			},
		}, &discordgo.Member{
			Roles: []string{"testRole", "testRole2", "testRoleAdmin"},
			User: &discordgo.User{
				ID:       "testID2",
				Username: "TestUserName2",
				Avatar:   "a_testAvatarString",
			},
		}, &discordgo.Member{
			Roles: []string{},
			User: &discordgo.User{
				ID:       "testID3",
				Username: "TestUserName3",
				Avatar:   "a_testAvatarString",
			},
		},
	)
	return &members
}

func GetMockMembers(t *testing.T) *utils.Collection[*discordgo.Member] {
	if db == nil {
		setup()
	}

	rows, err := db.Query("SELECT id, username, nickname, avatar, roles FROM Users")
	if err != nil {
		t.Fatal(err.Error())
	}

	members := utils.NewCollection([]*discordgo.Member{})
	for rows.Next() {
		var memberID, username, avatar, roles string
		var nullableNickname sql.NullString
		err = rows.Scan(&memberID, &username, &nullableNickname, &avatar, &roles)
		if err != nil {
			t.Fatal(err.Error())
		}

		rolesList := strings.Split(roles, ",")

		var nickname string
		if nullableNickname.Valid {
			nickname = nullableNickname.String
		}

		members.Insert(&discordgo.Member{
			Roles: rolesList,
			Nick:  nickname,
			User: &discordgo.User{
				ID:       memberID,
				Username: username,
				Avatar:   avatar,
			},
		})
	}
	return &members
}

func GetMockDBUsers(t *testing.T) []api.DbUser {
	if db == nil {
		setup()
	}

	dbUser, err := db.Query("SELECT * FROM Users")
	if err != nil {
		t.Fatal("Error while querying db")
	}

	var dbUsers []api.DbUser
	for dbUser.Next() {
		var user api.DbUser = api.DbUser{}
		var roles string
		var nickname, lodestone_id sql.NullString
		err := dbUser.Scan(&user.ID, &user.Username, &nickname, &user.Avatar, &roles, &lodestone_id)
		if err != nil {
			t.Fatal("Error while scanning db: ", err.Error())
		}

		user.Roles = strings.Split(roles, ",")

		if nickname.Valid {
			user.NickName = nickname.String
		}

		if lodestone_id.Valid {
			user.LodeStoneId = lodestone_id.String
		}

		dbUsers = append(dbUsers, user)
	}

	return dbUsers
}
