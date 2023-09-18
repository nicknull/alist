package data

import "github.com/alist-org/alist/v3/cmd/flags"

func InitData() {
	initUser()
	initSettings()
	if flags.Dev {
		initDevData()
		initDevDo()
	}
}

func InitDataIOS() (err error) {
	err = initUserIOS()
	if err != nil {
		return err
	}
	err = initSettingsIOS()
	if err != nil {
		return err
	}
	return
}
