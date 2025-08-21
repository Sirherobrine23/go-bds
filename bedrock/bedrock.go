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

	// "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",

	MojangHeaders = request.Header{
		// "Accept-Encoding":           "gzip, deflate, br, zstd",
		"Accept-Language":           "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7",
		"Priority":                  "u=0, i",
		"Sec-Ch-Ua":                 "\"Not;A=Brand\";v=\"99\", \"Google Chrome\";v=\"139\", \"Chromium\";v=\"139\"",
		"Sec-Ch-Ua-Mobile":          "?0",
		"Sec-Ch-Ua-Platform":        "\"Linux\"",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36",
	}

	permisionName = []string{
		Visitor:  "visitor",
		Member:   "member",
		Operator: "operator",
	}
)
