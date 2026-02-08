package validation

import "regexp"

var nameRegex = regexp.MustCompile(`[^\p{L}\p{N} \-_.,']`)

func SanitizeName(name string) string {
	return nameRegex.ReplaceAllString(name, "")
}
