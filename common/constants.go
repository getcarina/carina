package common

import "github.com/getcarina/carina/version"

// BuildUserAgent generates the user agent for the Carina client
func BuildUserAgent() string {
	return " carina/" + version.Version
}
