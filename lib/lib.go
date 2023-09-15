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

	"github.com/alist-org/alist/v3/internal/aria2"
	"github.com/alist-org/alist/v3/internal/bootstrap"
	"github.com/alist-org/alist/v3/internal/bootstrap/data"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/qbittorrent"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server"
	"github.com/alist-org/alist/v3/server/common"

	_ "github.com/alist-org/alist/v3/drivers"

	_ "golang.org/x/mobile/bind"
)

type Instance struct {
	server *http.Server
}

func (i *Instance) Server(dir string) string {

	dir = filepath.Join(dir, "data")

	bootstrap.InitConfig(dir)
	bootstrap.Log()
	bootstrap.InitDB()
	data.InitData()
	bootstrap.InitIndex()
	i.initAria2()
	i.initQbittorrent()
	i.loadStorages()

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.LoggerWithWriter(logrus.StandardLogger().Out), gin.RecoveryWithWriter(logrus.StandardLogger().Out))
	server.Init(engine)
	i.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", conf.Conf.Scheme.Address, conf.Conf.Scheme.HttpPort),
		Handler: engine,
	}
	go func() {
		err := i.server.ListenAndServe()
		if err != nil {
			utils.Log.Fatalf("failed to server: %+v", err)
		}
	}()
	for {
		time.Sleep(time.Second)
		rsp, err := http.Get("http://localhost:5244/ping")
		if err != nil {
			utils.Log.Println("start server failed, try later... : %v", err)
			continue
		}
		_ = rsp.Body.Close()
		if rsp.StatusCode != http.StatusOK {
			utils.Log.Println("start server failed, try later... : %d", rsp.StatusCode)
			continue
		}
		utils.Log.Println("start server success")
		token, err := common.GenerateToken("admin")
		if err != nil {
			utils.Log.Fatal("generate token failed, exit app...")
		}
		return token
	}
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

func (i *Instance) initAria2() {
	utils.Log.Infof("start init Aria2.")
	_, err := aria2.InitClient(2)
	if err != nil {
		utils.Log.Infof("Aria2 not ready.")
	} else {
		utils.Log.Infof("success init Aria2.")
	}
}

func (i *Instance) initQbittorrent() {
	utils.Log.Infof("start init qbittorrent.")
	err := qbittorrent.InitClient()
	if err != nil {
		utils.Log.Infof("qbittorrent not ready.")
	} else {
		utils.Log.Infof("success init qbittorrent.")
	}
}

func (i *Instance) loadStorages() {
	storages, err := db.GetEnabledStorages()
	if err != nil {
		utils.Log.Fatalf("failed get enabled storages: %+v", err)
	}
	for i := range storages {
		utils.Log.Infof("start load storage: [%s], driver: [%s]",
			storages[i].MountPath, storages[i].Driver)
		err := op.LoadStorage(context.Background(), storages[i])
		if err != nil {
			utils.Log.Errorf("failed get enabled storages: %+v", err)
		} else {
			utils.Log.Infof("success load storage: [%s], driver: [%s]",
				storages[i].MountPath, storages[i].Driver)
		}
	}
	conf.StoragesLoaded = true
}
