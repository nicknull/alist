package bootstrap

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/pkg/utils"
)

func InitConfig(dir string) {

	configPath := filepath.Join(dir, "config.json")

	logrus.Infof("reading config file: %s", configPath)

	if !utils.Exists(configPath) {
		logrus.Infof("config file not exists, creating default config file")
		_, err := utils.CreateNestedFile(configPath)
		if err != nil {
			logrus.Fatalf("failed to create config file: %+v", err)
		}
		conf.Conf = conf.DefaultConfig(dir)
		if !utils.WriteJsonToFile(configPath, conf.Conf) {
			logrus.Fatalf("failed to create default config file")
		}
	} else {
		configBytes, err := os.ReadFile(configPath)
		if err != nil {
			logrus.Fatalf("reading config file error: %+v", err)
		}
		conf.Conf = conf.DefaultConfig(dir)
		err = utils.Json.Unmarshal(configBytes, conf.Conf)
		if err != nil {
			logrus.Fatalf("load config error: %+v", err)
		}
		// update config.json struct
		confBody, err := utils.Json.MarshalIndent(conf.Conf, "", "  ")
		if err != nil {
			logrus.Fatalf("marshal config error: %+v", err)
		}
		err = os.WriteFile(configPath, confBody, 0o777)
		if err != nil {
			logrus.Fatalf("update config struct error: %+v", err)
		}
	}
	if !filepath.IsAbs(conf.Conf.TempDir) {
		absPath, err := filepath.Abs(conf.Conf.TempDir)
		if err != nil {
			logrus.Fatalf("get abs path error: %+v", err)
		}
		conf.Conf.TempDir = absPath
	}
	err := os.RemoveAll(filepath.Join(conf.Conf.TempDir))
	if err != nil {
		logrus.Errorln("failed delete temp file:", err)
	}
	err = os.MkdirAll(conf.Conf.TempDir, 0o777)
	if err != nil {
		logrus.Fatalf("create temp dir error: %+v", err)
	}
	logrus.Debugf("config: %+v", conf.Conf)

	base.InitClient()
}
