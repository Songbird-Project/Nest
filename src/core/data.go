package core

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

func GetUsers() ([]string, []string, error) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var users []string
	var homes []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		userInfo := strings.Split(line, ":")

		if len(userInfo) < 6 {
			continue
		}

		username := userInfo[0]
		uidStr := userInfo[2]
		home := userInfo[5]

		uid, err := strconv.Atoi(uidStr)
		if err != nil || uid == 0 {
			continue
		}

		if uid < 1000 {
			continue
		}

		users = append(users, username)
		homes = append(homes, home)
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return users, homes, nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}
