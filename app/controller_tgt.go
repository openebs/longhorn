package app

import (
	"github.com/openebs/longhorn/frontend/tgt"
)

func init() {
	frontends["tgt"] = tgt.New()
}
