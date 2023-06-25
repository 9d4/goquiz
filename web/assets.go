//go:build !dev
// +build !dev

package web

import "embed"

//go:embed dist/*
var fs embed.FS

func Assets() embed.FS {
	return fs
}

//go:embed dashboard.html
var dashboard []byte

func Dashboard() []byte {
	return dashboard
}
