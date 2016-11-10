// +build tcmu

package app

import (
	"github.com/openebs/longhorn/frontend/tcmu"
)

func init() {
	frontends["tcmu"] = tcmu.New()
}
