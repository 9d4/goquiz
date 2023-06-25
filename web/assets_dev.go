//go:build dev
// +build dev

package web

import (
	iofs "io/fs"
	"os"
)

var fs iofs.FS = os.DirFS("web")

func Assets() iofs.FS {
	return fs
}

var dashboard []byte

func Dashboard() []byte {
	b, err := os.ReadFile("web/dashboard.html")
	dashboard = b
	if err != nil {
		dashboard = []byte("dashboard.html read error")
	}

	return dashboard
}
