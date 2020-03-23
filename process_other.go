// +build !396,!amd64,!amd64p32

package subvert

import (
	"fmt"
)

func osRedirectCalls(src, dst uintptr) (err error) {
	return fmt.Errorf("Not implemented on this arch")
}
