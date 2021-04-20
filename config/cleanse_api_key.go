package config

import "strings"

// CleanseApiKey removes characters like \r or \n that cause issues with authentication
// These characters accidentally get introduced by users copy/pasting into a file usually
func CleanseApiKey(apiKey string) string {
	apiKey = strings.Replace(apiKey, "\r", "", -1)
	apiKey = strings.Replace(apiKey, "\n", "", -1)
	return apiKey
}
