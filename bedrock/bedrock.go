// Minecraft bedrock oficial server from Mojang
package bedrock

import (
	"errors"

	"sirherobrine23.com.br/go-bds/request/v2"
)

// Player permision type
const (
	Visitor PermissionLevel = iota
	Member
	Operator
)

var (
	ErrNoVersion error = errors.New("version not found")
	ErrPlatform  error = errors.New("current platform no supported or cannot emulate arch") // Cannot run server in platform or cannot emulate arch

	MojangHeaders = request.Header{
		// "Accept-Encoding":           "gzip, deflate",
		"Accept-Language":           "en-US;q=0.9,en;q=0.8",
		"Priority":                  "u=0, i",
		"Sec-Ch-Ua":                 "\"Google Chrome\";v=\"131\", \"Chromium\";v=\"131\", \"Not_A Brand\";v=\"24\"",
		"Sec-Ch-Ua-Mobile":          "?0",
		"Sec-Ch-Ua-Platform":        "\"Linux\"",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	}

	permisionName = []string{
		Visitor:  "visitor",
		Member:   "member",
		Operator: "operator",
	}
)
