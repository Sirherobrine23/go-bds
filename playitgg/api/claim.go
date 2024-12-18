package api

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

const (
	AgentDefault Agent = iota + 1
	AgentAssignable
	AgentSelfManaged
)

type Agent int // Agents

func (agent Agent) MarshalText() ([]byte, error) {
	switch agent {
	case AgentAssignable:
		return []byte("assignable"), nil
	case AgentSelfManaged:
		return []byte("self-managed"), nil
	default:
		return []byte("default"), nil
	}
}
func (agent *Agent) UnmarshalText(text []byte) error {
	switch string(text) {
	case "assignable":
		*agent = AgentAssignable
	case "self-managed":
		*agent = AgentSelfManaged
	default:
		*agent = AgentDefault
	}
	return nil
}

func (w *Api) AiisgnClaimCode() (err error) {
	if len(w.Code) > 0 {
		return nil
	}

	// Make code buffer
	codeBuff := make([]byte, 5)
	if _, err = rand.Read(codeBuff); err != nil {
		return err
	}

	// Convert to hex string
	w.Code = hex.EncodeToString(codeBuff)
	return nil
}

// Get claim url
func (w *Api) ClaimUrl() string {
	return fmt.Sprintf("https://playit.gg/claim/%s", url.PathEscape(w.Code))
}

func (w *Api) ClaimAgentSecret(AgentType string) error {
	if w.Code == "" {
		return fmt.Errorf("assign claim code")
	} else if w.Secret != "" {
		return fmt.Errorf("agent secret key ared located")
	}

	type Claim struct {
		Code    string `json:"code"`       // Claim code
		Agent   string `json:"agent_type"` // "default" | "assignable" | "self-managed"
		Version string `json:"version"`    // Project version
	}
	type Code struct {
		Code      string `json:"code,omitempty"`
		SecretKey string `json:"secret_key,omitempty"`
	}

	assignSecretRequestBody, err := json.Marshal(Claim{
		Code:    w.Code,
		Agent:   AgentType,
		Version: fmt.Sprintf("go-playit %s", GoPlayitVersion),
	})
	if err != nil {
		return err
	}

	for {
		var waitUser string
		if waitUser, _, err = requestAPI[string](*w, "/claim/setup", bytes.NewReader(assignSecretRequestBody[:]), nil); err != nil {
			return err
		}

		if waitUser == "UserRejected" {
			return fmt.Errorf("claim rejected")
		} else if waitUser == "UserAccepted" {
			break
		}
		// wait for request
		time.Sleep(time.Second)
	}

	// Code to json
	getCode, _, err := requestAPI[Code](*w, "/claim/exchange", Code{Code: w.Code}, nil)
	if err != nil {
		return err
	}
	w.Secret = getCode.SecretKey // Copy secret key to Api struct
	return nil
}
