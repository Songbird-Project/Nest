package core

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/Songbird-Project/scsv"
)

func GetUserConfigs() ([]User, error) {
	users := []User{}

	userConfigSCSV, err := scsv.ParseFile(filepath.Join(os.Getenv("NEST_AUTOGEN"), "users.scsv"))
	if err != nil {
		return nil, err
	}

	for userName := range userConfigSCSV {
		userInfo := userConfigSCSV[userName]

		manageHome, err := strconv.ParseBool(userInfo[3][1])
		if err != nil {
			return nil, err
		}

		user := User{
			Username:   userName,
			Fullname:   userInfo[0][1],
			HomeDir:    userInfo[2][1],
			ManageHome: manageHome,
			Shell:      userInfo[1][1],
			Groups:     userInfo[3],
		}

		users = append(users, user)
	}

	return users, nil
}
