package validation

import (
	"krizzy/internal/models"
	"regexp"
	"strings"
)

var nameRegex = regexp.MustCompile(`[^\p{L}\p{N} \-_.,']`)
var personColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func SanitizeName(name string) string {
	return strings.TrimSpace(nameRegex.ReplaceAllString(name, ""))
}

func NormalizePersonColor(color string) string {
	color = strings.TrimSpace(color)
	if personColorRegex.MatchString(color) {
		return strings.ToUpper(color)
	}
	return models.DefaultPersonColor
}
