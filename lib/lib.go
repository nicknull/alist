package AList

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/alist-org/alist/v3/internal/bootstrap"
	"github.com/alist-org/alist/v3/internal/bootstrap/data"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/pkg/utils/random"
	"github.com/alist-org/alist/v3/server"

	_ "golang.org/x/mobile/bind"
)

type Instance struct {
	server *http.Server
}

func (i *Instance) Server(dir string) {

	bootstrap.InitConfig(dir)
	bootstrap.Log()
	bootstrap.InitDB()
	data.InitData()
	bootstrap.InitIndex()
	bootstrap.InitAria2()
	bootstrap.InitQbittorrent()
	bootstrap.LoadStorages()

	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.LoggerWithWriter(logrus.StandardLogger().Out), gin.RecoveryWithWriter(logrus.StandardLogger().Out))
	server.Init(engine)

	i.server = &http.Server{Addr: fmt.Sprintf("%s:%d", conf.Conf.Scheme.Address, conf.Conf.Scheme.HttpPort), Handler: engine}

	go func() {
		err := i.server.ListenAndServe()
		if err != nil {
			logrus.Fatalf("failed to server: %+v", err)
		}
	}()
}

func (i *Instance) RandomString(n int) string {
	return random.String(n)
}

func (i *Instance) SetEnv(key, value string) error {
	return os.Setenv(key, value)
}

func (i *Instance) GetEnv(key string) string {
	return os.Getenv(key)
}
