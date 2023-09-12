package AList

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/alist-org/alist/v3/internal/bootstrap"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/server"

	_ "golang.org/x/mobile/bind"
)

type Instance struct {
	server *http.Server
}

func (i *Instance) Server() {

	bootstrap.InitAria2()
	bootstrap.InitQbittorrent()
	bootstrap.LoadStorages()

	engine := gin.New()
	engine.Use(gin.LoggerWithWriter(logrus.StandardLogger().Out), gin.RecoveryWithWriter(logrus.StandardLogger().Out))
	server.Init(engine)

	i.server = &http.Server{Addr: fmt.Sprintf("%s:%d", "127.0.0.1", 5244), Handler: engine}

	go func() {
		err := i.server.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()
}

func (i *Instance) GetAdminPassword() string {
	user, err := op.GetAdmin()
	if err != nil {
		return ""
	} else {
		return user.Password
	}
}
