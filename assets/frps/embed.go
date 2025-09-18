package frpc

import (
	"embed"

	"github.com/laosan-xx/frp/assets"
)

//go:embed static/*
var content embed.FS

func init() {
	assets.Register(content)
}
