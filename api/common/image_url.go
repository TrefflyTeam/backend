package common

import (
	"fmt"
	"strings"
)

func ImageURL(env, domain, path string) string {
	if path == "" {
		return ""
	}
	prefix := ""
	protocol := "http"
	if env == "production" {
		protocol = "https"
		prefix = "/api"
	}
	normalizedPath := strings.ReplaceAll(path, "\\", "/")

	url := fmt.Sprintf("%s://%s%s/images/%s", protocol, domain, prefix, normalizedPath)

	return url
}
