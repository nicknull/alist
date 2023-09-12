package AList

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/pkg/utils/random"
)

func initConfig(dir string) {

	configPath := filepath.Join(dir, "config.json")

	logrus.Infof("reading config file: %s", configPath)

	if !utils.Exists(configPath) {
		logrus.Infof("config file not exists, creating default config file")
		_, err := utils.CreateNestedFile(configPath)
		if err != nil {
			logrus.Fatalf("failed to create config file: %+v", err)
		}
		conf.Conf = defaultConfig(dir)
		if !utils.WriteJsonToFile(configPath, conf.Conf) {
			logrus.Fatalf("failed to create default config file")
		}
	} else {
		configBytes, err := os.ReadFile(configPath)
		if err != nil {
			logrus.Fatalf("reading config file error: %+v", err)
		}
		conf.Conf = defaultConfig(dir)
		err = utils.Json.Unmarshal(configBytes, conf.Conf)
		if err != nil {
			logrus.Fatalf("load config error: %+v", err)
		}
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

	initURL()
}

func defaultConfig(dir string) *conf.Config {
	tempDir := filepath.Join(dir, "temp")
	indexDir := filepath.Join(dir, "bleve")
	logPath := filepath.Join(dir, "log/log.log")
	dbPath := filepath.Join(dir, "data.db")
	return &conf.Config{
		Scheme: conf.Scheme{
			Address:    "127.0.0.1",
			UnixFile:   "",
			HttpPort:   5244,
			HttpsPort:  -1,
			ForceHttps: false,
			CertFile:   "",
			KeyFile:    "",
		},
		JwtSecret:      random.String(16),
		TokenExpiresIn: 48,
		TempDir:        tempDir,
		Database: conf.Database{
			Type:        "sqlite3",
			Port:        0,
			TablePrefix: "x_",
			DBFile:      dbPath,
		},
		BleveDir: indexDir,
		Log: conf.LogConfig{
			Enable:     true,
			Name:       logPath,
			MaxSize:    50,
			MaxBackups: 30,
			MaxAge:     28,
		},
		MaxConnections:        0,
		TlsInsecureSkipVerify: true,
	}
}

func initURL() {
	if !strings.Contains(conf.Conf.SiteURL, "://") {
		conf.Conf.SiteURL = utils.FixAndCleanPath(conf.Conf.SiteURL)
	}
	u, err := url.Parse(conf.Conf.SiteURL)
	if err != nil {
		utils.Log.Fatalf("can't parse site_url: %+v", err)
	}
	conf.URL = u
}
