package nodotimport

import (
	"fmt"
	. "strings" // want `dot import of "strings"`

	alias "bytes"
)

func Use() string {
	return fmt.Sprint(ToUpper("x"), alias.MinRead)
}
