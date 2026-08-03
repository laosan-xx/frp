//go:build !noweb

package frpc

import (
	"embed"

	"github.com/laosan-xx/frp/assets"
)

//go:embed dist
var EmbedFS embed.FS

func init() {
	assets.Register(EmbedFS)
}
