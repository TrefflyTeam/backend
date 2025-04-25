package common

import (
	"fmt"
	"strings"
)

func ImageURL(env, domain, path string) string {
	if path == "" {
		return ""
	}

	protocol := "http"
	if env == "production" {
		protocol = "https"
	}
	normalizedPath := strings.ReplaceAll(path, "\\", "/")

	url := fmt.Sprintf("%s://%s/images/%s", protocol, domain, normalizedPath)

	return url
}
