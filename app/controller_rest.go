package app

import (
	"github.com/openebs/longhorn/frontend/rest"
)

func init() {
	frontends["rest"] = rest.New()
}
