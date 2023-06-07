package web

import "embed"

//go:embed dist/*
var fs embed.FS

func Assets() embed.FS {
	return fs
}
