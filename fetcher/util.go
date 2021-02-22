package fetcher

import "strings"

func stringIsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}
