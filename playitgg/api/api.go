// API to playit.gg
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"

	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

var (
	PlayitAPI       string = "https://api.playit.gg" // Playit API
	GoPlayitVersion string = "0.17.1"                // playtit.gg agent version

	ErrInvalidAuth            error = errors.New("cannot process api request because return not auth")
	ErrTooMenyRequest         error = errors.New("wait seconds to make new request")
	ErrNotFound               error = errors.New("path request not found")
	ErrAuthRequired           error = errors.New("auth required")
	ErrInvalidHeader          error = errors.New("invalid header")
	ErrInvalidSignature       error = errors.New("invalid signature")
	ErrInvalidTimestamp       error = errors.New("invalid timestamp")
	ErrInvalidApiKey          error = errors.New("invalid api key")
	ErrInvalidAgentKey        error = errors.New("invalid agent key")
	ErrSessionExpired         error = errors.New("session expired")
	ErrInvalidAuthType        error = errors.New("invalid auth type")
	ErrScopeNotAllowed        error = errors.New("scope not allowed")
	ErrNoLongerValid          error = errors.New("no longer valid")
	ErrGuestAccountNotAllowed error = errors.New("guest account not allowed")
	ErrEmailMustBeVerified    error = errors.New("email must be verified")
	ErrAccountDoesNotExist    error = errors.New("account does not exist")
	ErrAdminOnly              error = errors.New("admin only")
	ErrInvalidToken           error = errors.New("invalid token")
	ErrTotpRequred            error = errors.New("totp requred")
)

const (
	TunnelTypeMCBedrock TunnelType = iota + 1 // Minecraft Bedrock server
	TunnelTypeMCJava                          // Minecraft java server
	TunnelTypeValheim                         // valheim
	TunnelTypeTerraria                        // Terraria multiplayer
	TunnelTypeStarbound                       // starbound
	TunnelTypeRust                            // Rust (No programmer language)
	TunnelType7Days                           // 7days
	TunnelTypeUnturned                        // unturned
)

const (
	PortTypeBoth PortType = iota // Tunnel support tcp and udp protocol
	PortTypeTcp                  // Tunnel support only tcp protocol
	PortTypeUdp                  // Tunnel support only udp protocol
)

// Default regions to allocate or allocated
const (
	RegionGlobal       AllocationRegion = iota + 1 // Free account and premium
	RegionSmartGlobal                              // Require premium account
	RegionNorthAmerica                             // Require premium account
	RegionEurope                                   // Require premium account
	RegionAsia                                     // Require premium account
	RegionIndia                                    // Require premium account
	RegionSouthAmerica                             // Require premium account
)

type errStatus struct {
	Status string `json:"status"`
	Data   struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"data"`
}

func (err errStatus) Error() error {
	switch err.Data.Type {
	case "internal", "validation":
		return errors.New(err.Data.Message)
	case "path-not-found":
		return ErrNotFound
	case "auth":
		switch err.Data.Message {
		case "AuthRequired":
			return ErrAuthRequired
		case "InvalidHeader":
			return ErrInvalidHeader
		case "InvalidSignature":
			return ErrInvalidSignature
		case "InvalidTimestamp":
			return ErrInvalidTimestamp
		case "InvalidApiKey":
			return ErrInvalidApiKey
		case "InvalidAgentKey":
			return ErrInvalidAgentKey
		case "SessionExpired":
			return ErrSessionExpired
		case "InvalidAuthType":
			return ErrInvalidAuthType
		case "ScopeNotAllowed":
			return ErrScopeNotAllowed
		case "NoLongerValid":
			return ErrNoLongerValid
		case "GuestAccountNotAllowed":
			return ErrGuestAccountNotAllowed
		case "EmailMustBeVerified":
			return ErrEmailMustBeVerified
		case "AccountDoesNotExist":
			return ErrAccountDoesNotExist
		case "AdminOnly":
			return ErrAdminOnly
		case "InvalidToken":
			return ErrInvalidToken
		case "TotpRequred":
			return ErrTotpRequred
		}
	}
	return nil
}

type Map map[string]any

func postRequest[T any](path, authToken string, body any, options *request.Options) (T, error) {
	if options == nil {
		options = &request.Options{}
	}

	res, err := request.Request(fmt.Sprintf("%s/%s", PlayitAPI, path), &request.Options{
		Method: "POST",
		Body:   body,
		Header: options.Header.Merge(map[string]string{"authorization": authToken}),
		CodeProcess: map[int]request.CodeCallback{
			429: func(res *http.Response) (*http.Response, error) { return nil, ErrTooMenyRequest },
			401: func(res *http.Response) (*http.Response, error) { return nil, ErrInvalidAuth },
			-1:  func(res *http.Response) (*http.Response, error) { return res, nil },
		},
	})
	if err != nil {
		return *(*T)(nil), err
	}
	defer res.Body.Close()

	fullBody, err := io.ReadAll(res.Body)
	if err != nil {
		return *(*T)(nil), err
	}

	if slices.Contains([]int{200, 201}, res.StatusCode) {
		var n T
		if err = json.Unmarshal(fullBody, &n); err != nil {
			return *(*T)(nil), err
		}
		return n, nil
	}

	var errorStr errStatus
	if err = json.Unmarshal(fullBody, &errorStr); err != nil {
		return *(*T)(nil), err
	}
	return *(*T)(nil), errorStr.Error()
}
