package AList

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/alist-org/alist/v3/internal/bootstrap"
	"github.com/alist-org/alist/v3/internal/bootstrap/data"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server"
	"github.com/alist-org/alist/v3/server/common"

	_ "github.com/alist-org/alist/v3/drivers"

	_ "golang.org/x/mobile/bind"
)

type Instance struct {
	server *http.Server
}

func (i *Instance) Server(dir string) {

	dir = filepath.Join(dir, "data")

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
			utils.Log.Fatalf("failed to server: %+v", err)
		}
	}()
}

func (i *Instance) GenerateToken() string {
	token, _ := common.GenerateToken("admin")
	return token
}

func (i *Instance) Shutdown() {
	db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := i.server.Shutdown(ctx); err != nil {
			utils.Log.Fatal("HTTP server shutdown err: ", err)
		}
	}()
	wg.Wait()
	utils.Log.Println("Server exit")
}
