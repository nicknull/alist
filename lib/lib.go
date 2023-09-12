package alist

import (
	"github.com/alist-org/alist/v3/cmd"

	_ "golang.org/x/mobile/bind"
)

//export Run
func Run() {
	cmd.Execute()
}
