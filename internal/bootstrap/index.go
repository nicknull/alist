package bootstrap

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/search"
	log "github.com/sirupsen/logrus"
)

func InitIndex() {
	progress, err := search.Progress()
	if err != nil {
		log.Errorf("init index error: %+v", err)
		return
	}
	if !progress.IsDone {
		progress.IsDone = true
		search.WriteProgress(progress)
	}
}

func InitIndexIOS() error {
	progress, err := search.Progress()
	if err != nil {
		log.Errorf("init index error: %+v", err)
		return fmt.Errorf("索引初始化失败")
	}
	if !progress.IsDone {
		progress.IsDone = true
		search.WriteProgress(progress)
	}
	return nil
}
