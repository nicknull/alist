package bootstrap

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/alist-org/alist/v3/cmd/flags"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/caarlos0/env/v9"
	log "github.com/sirupsen/logrus"
)

func InitConfig() {
	if flags.ForceBinDir {
		if !filepath.IsAbs(flags.DataDir) {
			ex, err := os.Executable()
			if err != nil {
				utils.Log.Fatal(err)
			}
			exPath := filepath.Dir(ex)
			flags.DataDir = filepath.Join(exPath, flags.DataDir)
		}
	}
	configPath := filepath.Join(flags.DataDir, "config.json")
	log.Infof("reading config file: %s", configPath)
	if !utils.Exists(configPath) {
		log.Infof("config file not exists, creating default config file")
		_, err := utils.CreateNestedFile(configPath)
		if err != nil {
			log.Fatalf("failed to create config file: %+v", err)
		}
		conf.Conf = conf.DefaultConfig()
		if !utils.WriteJsonToFile(configPath, conf.Conf) {
			log.Fatalf("failed to create default config file")
		}
	} else {
		configBytes, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatalf("reading config file error: %+v", err)
		}
		conf.Conf = conf.DefaultConfig()
		err = utils.Json.Unmarshal(configBytes, conf.Conf)
		if err != nil {
			log.Fatalf("load config error: %+v", err)
		}
		// update config.json struct
		confBody, err := utils.Json.MarshalIndent(conf.Conf, "", "  ")
		if err != nil {
			log.Fatalf("marshal config error: %+v", err)
		}
		err = os.WriteFile(configPath, confBody, 0o777)
		if err != nil {
			log.Fatalf("update config struct error: %+v", err)
		}
	}
	if !conf.Conf.Force {
		confFromEnv()
	}
	// convert abs path
	if !filepath.IsAbs(conf.Conf.TempDir) {
		absPath, err := filepath.Abs(conf.Conf.TempDir)
		if err != nil {
			log.Fatalf("get abs path error: %+v", err)
		}
		conf.Conf.TempDir = absPath
	}
	err := os.RemoveAll(filepath.Join(conf.Conf.TempDir))
	if err != nil {
		log.Errorln("failed delete temp file:", err)
	}
	err = os.MkdirAll(conf.Conf.TempDir, 0o777)
	if err != nil {
		log.Fatalf("create temp dir error: %+v", err)
	}
	log.Debugf("config: %+v", conf.Conf)
	base.InitClient()
	initURL()
}

func confFromEnv() {
	prefix := "ALIST_"
	if flags.NoPrefix {
		prefix = ""
	}
	log.Infof("load config from env with prefix: %s", prefix)
	if err := env.ParseWithOptions(conf.Conf, env.Options{
		Prefix: prefix,
	}); err != nil {
		log.Fatalf("load config from env error: %+v", err)
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

func InitConfigIOS(dir string) error {

	configPath := filepath.Join(dir, "config.json")

	log.Infof("reading config file: %s", configPath)

	if !utils.Exists(configPath) {
		log.Infof("config file not exists, creating default config file")
		_, err := utils.CreateNestedFile(configPath)
		if err != nil {
			log.Errorf("failed to create config file: %+v", err)
			return fmt.Errorf("创建默认配置失败")
		}
		conf.Conf = conf.DefaultConfig()
		conf.Conf.ResolvePaths(dir)
		if !utils.WriteJsonToFile(configPath, conf.Conf) {
			log.Errorf("failed to create default config file")
			return fmt.Errorf("默认配置写入失败")
		}
	} else {
		configBytes, err := os.ReadFile(configPath)
		if err != nil {
			log.Errorf("reading config file error: %+v", err)
			return fmt.Errorf("读取配置失败")
		}
		conf.Conf = conf.DefaultConfig()
		err = utils.Json.Unmarshal(configBytes, conf.Conf)
		if err != nil {
			log.Errorf("load config error: %+v", err)
			return fmt.Errorf("读取配置失败")
		}
		conf.Conf.ResolvePaths(dir)
		confBody, err := utils.Json.MarshalIndent(conf.Conf, "", "  ")
		if err != nil {
			log.Errorf("marshal config error: %+v", err)
			return fmt.Errorf("更新配置失败")
		}
		err = os.WriteFile(configPath, confBody, 0o777)
		if err != nil {
			log.Errorf("update config struct error: %+v", err)
			return fmt.Errorf("更新配置失败")
		}
	}
	if !filepath.IsAbs(conf.Conf.TempDir) {
		absPath, err := filepath.Abs(conf.Conf.TempDir)
		if err != nil {
			log.Errorf("get abs path error: %+v", err)
			return fmt.Errorf("获取配置文件路径失败")
		}
		conf.Conf.TempDir = absPath
	}
	err := os.RemoveAll(filepath.Join(conf.Conf.TempDir))
	if err != nil && os.IsNotExist(err) {
		log.Errorf("failed delete temp file: %+v", err)
		return fmt.Errorf("删除缓存目录失败")
	}
	err = os.MkdirAll(conf.Conf.TempDir, 0o777)
	if err != nil {
		log.Errorf("create temp dir error: %+v", err)
		return fmt.Errorf("创建缓存目录失败")
	}
	log.Debugf("config: %+v", conf.Conf)

	base.InitClient()

	initURL()

	return nil
}
