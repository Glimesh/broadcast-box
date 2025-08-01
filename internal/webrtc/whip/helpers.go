package whip

import (
	"fmt"
	"slices"
)

func NextAvailableName(prefix string, names []string) string {
	for i := 0; ; i++ {
		name := prefix
		if i > 0 {
			name = fmt.Sprintf(prefix+"_%d", i)
		}
		found := slices.Contains(names, name)
		if !found {
			return name
		}
	}
}
