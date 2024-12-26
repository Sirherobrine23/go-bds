package playitapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

// Generic playit.gg body process
type ApiResponseBody[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

type Api struct {
	Secret string `json:"secret,omitempty"` // Agent Secret
}

// Errors returned by the API
var (
	ErrInternal                    error = errors.New("internal error in api request")
	ErrInvalidAuth                 error = errors.New("cannot process api request because return not auth")
	ErrTooMenyRequest              error = errors.New("wait seconds to make new request")
	ErrNotFound                    error = errors.New("path request not found")
	ErrAuthRequired                error = errors.New("auth required")
	ErrInvalidHeader               error = errors.New("invalid header")
	ErrInvalidSignature            error = errors.New("invalid signature")
	ErrInvalidTimestamp            error = errors.New("invalid timestamp")
	ErrInvalidApiKey               error = errors.New("invalid api key")
	ErrInvalidAgentKey             error = errors.New("invalid agent key")
	ErrSessionExpired              error = errors.New("session expired")
	ErrInvalidAuthType             error = errors.New("invalid auth type")
	ErrScopeNotAllowed             error = errors.New("scope not allowed")
	ErrNoLongerValid               error = errors.New("no longer valid")
	ErrGuestAccountNotAllowed      error = errors.New("guest account not allowed")
	ErrEmailMustBeVerified         error = errors.New("email must be verified")
	ErrAccountDoesNotExist         error = errors.New("account does not exist")
	ErrAdminOnly                   error = errors.New("admin only")
	ErrInvalidToken                error = errors.New("invalid token")
	ErrTotpRequred                 error = errors.New("totp requred")
	ErrAgentIdRequired             error = errors.New("agent id required")
	ErrAgentNotFound               error = errors.New("agent not found")
	ErrInvalidAgentId              error = errors.New("invalid agent id")
	ErrDedicatedIpNotFound         error = errors.New("dedicated ip not found")
	ErrDedicatedIpPortNotAvailable error = errors.New("dedicated ip port not available")
	ErrDedicatedIpNotEnoughSpace   error = errors.New("dedicated ip not enough space")
	ErrPortAllocNotFound           error = errors.New("port alloc not found")
	ErrInvalidIpHostname           error = errors.New("invalid ip hostname")
	ErrManagedMissingAgentId       error = errors.New("managed missing agent id")
	ErrInvalidPortCount            error = errors.New("invalid port count")
)

type ErrResponse struct {
	Status string `json:"status"`
	Data   struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"data"`
}

func (err ErrResponse) Error() error {
	switch err.Data.Type {
	case "internal":
		return ErrInternal
	case "validation":
		return errors.New(err.Data.Message)
	case "path-not-found":
		return ErrNotFound
	}

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
	case "AgentIdRequired":
		return ErrAgentIdRequired
	case "AgentNotFound":
		return ErrAgentNotFound
	case "InvalidAgentId":
		return ErrInvalidAgentId
	case "DedicatedIpNotFound":
		return ErrDedicatedIpNotFound
	case "DedicatedIpPortNotAvailable":
		return ErrDedicatedIpPortNotAvailable
	case "DedicatedIpNotEnoughSpace":
		return ErrDedicatedIpNotEnoughSpace
	case "PortAllocNotFound":
		return ErrPortAllocNotFound
	case "InvalidIpHostname":
		return ErrInvalidIpHostname
	case "ManagedMissingAgentId":
		return ErrManagedMissingAgentId
	case "InvalidPortCount":
		return ErrInvalidPortCount
	// Return message error
	default:
		return errors.New(err.Data.Message)
	}
}

// Process requests from playit.gg API
func RequestAPI[T any](Secret, Path string, Body any, headers request.Header) (T, *http.Response, error) {
	// Set Authorization header
	n := request.Header{}
	if Secret != "" {
		n["Authorization"] = fmt.Sprintf("Agent-Key %s", strings.TrimSpace(Secret))
	}

	// Empty body
	if Body == nil {
		Body = struct{}{}
	}

	var requestOptions request.Options
	requestOptions = request.Options{
		Method: "POST",
		Body:   Body,
		Header: headers.Merge(n),
		CodeProcess: map[int]request.CodeCallback{
			200: func(res *http.Response) (*http.Response, error) { return res, nil }, // Ok
			201: func(res *http.Response) (*http.Response, error) { return res, nil },
			202: func(res *http.Response) (*http.Response, error) { return res, nil },
			429: func(res *http.Response) (*http.Response, error) {
				<-time.After(5 * time.Second)
				return request.Request(fmt.Sprintf("%s%s", PlayitAPI, Path), &requestOptions)
			},
			-1: func(res *http.Response) (*http.Response, error) { // Return error to response
				defer res.Body.Close()
				full, err := io.ReadAll(res.Body)
				if err != nil {
					return nil, err
				}

				var errSta ErrResponse
				if len(full) > 0 {
					if err := json.Unmarshal(full, &errSta); err != nil {
						return nil, err
					}
				}
				return nil, errSta.Error()
			},
		},
	}

	res, err := request.Request(fmt.Sprintf("%s%s", PlayitAPI, Path), &requestOptions)
	if err != nil {
		return *new(T), res, err
	}
	defer res.Body.Close()

	full, err := io.ReadAll(res.Body)
	if err != nil {
		return *new(T), res, err
	}

	var bodyProcess ApiResponseBody[T]
	if err := json.Unmarshal(full, &bodyProcess); err != nil {
		var n any
		json.Unmarshal(full, &n); full, _ = json.MarshalIndent(n, "", "  ")
		print(string(full) + "\n")
		return *new(T), nil, err
	}
	return bodyProcess.Data, res, nil
}

type objectID struct {
	ID uuid.UUID `json:"id"`
}
