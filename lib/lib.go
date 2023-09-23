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

func (i *Instance) LoadCore(dir string) (err error) {
	dir = filepath.Join(dir, "data")
	err = bootstrap.InitConfigIOS(dir)
	if err != nil {
		return
	}
	bootstrap.LogIOS()
	err = bootstrap.InitDBIOS()
	if err != nil {
		return
	}
	err = data.InitDataIOS()
	if err != nil {
		return
	}
	err = bootstrap.InitIndexIOS()
	if err != nil {
		return
	}
	err = bootstrap.LoadStoragesIOS()
	if err != nil {
		return
	}
	return
}

func (i *Instance) StartAPIServer() (err error) {
	token, err := common.GenerateToken("admin")
	if err != nil {
		return
	}
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Request.Header.Add("Authorization", token)
	})
	engine.Use(gin.LoggerWithWriter(logrus.StandardLogger().Out), gin.RecoveryWithWriter(logrus.StandardLogger().Out))
	server.Init(engine)
	i.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", conf.Conf.Scheme.Address, conf.Conf.Scheme.HttpPort),
		Handler: engine,
	}
	go func() {
		err := i.server.ListenAndServe()
		if err != nil {
			utils.Log.Errorf("服务退出错误: %v", err)
		} else {
			utils.Log.Infof("服务退出成功")
		}
	}()
	ping := fmt.Sprintf("http://%s:%d/%s", conf.Conf.Scheme.Address, conf.Conf.Scheme.HttpPort, "ping")
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		rsp, pingErr := http.Get(ping)
		if pingErr != nil {
			utils.Log.Println("start server failed, try later... : %v", err)
			continue
		}
		_ = rsp.Body.Close()
		if rsp.StatusCode != http.StatusOK {
			utils.Log.Println("start server failed, try later... : %d", rsp.StatusCode)
			continue
		}
		utils.Log.Println("start server success")
		return
	}
	err = fmt.Errorf("启动服务失败")
	return
}

func (i *Instance) RestartAPIServerIfNeeded() error {
	ping := fmt.Sprintf("http://%s:%d/%s", conf.Conf.Scheme.Address, conf.Conf.Scheme.HttpPort, "ping")
	rsp, err := http.Get(ping)
	if err == nil && rsp.StatusCode == http.StatusOK {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = i.server.Shutdown(ctx)
	}()
	wg.Wait()
	return i.StartAPIServer()
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
