package playitapi

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWaitingForUserVisit error = errors.New("waiting for user visit")
	ErrWaitingForUser      error = errors.New("waiting for user")
	ErrUserAccepted        error = errors.New("user accepted")
	ErrUserRejected        error = errors.New("user rejected")

	ClaimURL string = "https://playit.gg/claim"
)

type Claim struct {
	Code string
	api  *Api
}

func (api *Api) Claim() (*Claim, error) {
	codeClaim := make([]byte, 5)
	if _, err := rand.Read(codeClaim); err != nil {
		return nil, err
	}
	return &Claim{
		api:  api,
		Code: hex.EncodeToString(codeClaim),
	}, nil
}

type AgentClaimDetails struct {
	Name      string    `json:"name"`
	RemoteIP  net.IP    `json:"remote_ip"`
	AgentType AgentType `json:"agent_type"`
	Version   string    `json:"version"`
}

func (claim Claim) Details() (*AgentClaimDetails, error) {
	info, _, err := RequestAPI[*AgentClaimDetails](claim.api.Secret, "/claim/details", map[string]string{"code": claim.Code}, nil)
	return info, err
}
func (claim Claim) Exchange() (string, error) {
	info, _, err := RequestAPI[struct {
		Secret string `json:"secret_key"`
	}](claim.api.Secret, "/claim/exchange", map[string]string{"code": claim.Code}, nil)
	return info.Secret, err
}
func (claim Claim) Accept(name string, agentType AgentType) (uuid.UUID, error) {
	info, _, err := RequestAPI[struct {
		Id uuid.UUID `json:"agent_id"`
	}](claim.api.Secret, "/claim/accept", map[string]any{"code": claim.Code, "name": name, "agent_type": agentType}, nil)
	return info.Id, err
}
func (claim Claim) Reject() error {
	_, _, err := RequestAPI[any](claim.api.Secret, "/claim/reject", map[string]string{"code": claim.Code}, nil)
	return err
}

// Wait code responses for the claim
// Switch on the status to handle the response with:
//   - ErrWaitingForUserVisit
//   - ErrWaitingForUser
//   - ErrUserAccepted
//   - ErrUserRejected
//
// Any other status will return an error
func (claim *Claim) Setup(agentType AgentType) error {
	status, _, err := RequestAPI[string](claim.api.Secret, "/claim/setup", map[string]any{
		"code":       claim.Code,
		"agent_type": agentType,
		"version":    fmt.Sprintf("go-bds/%s", PlayitVersion),
	}, nil)
	switch status {
	case "WaitingForUserVisit":
		return ErrWaitingForUserVisit
	case "WaitingForUser":
		return ErrWaitingForUser
	case "UserAccepted":
		return ErrUserAccepted
	case "UserRejected":
		return ErrUserRejected
	}
	return err
}

func (claim Claim) String() string {
	return fmt.Sprintf("%s/%s", ClaimURL, url.PathEscape(claim.Code))
}

func (claim Claim) Do(agentType AgentType) error {
	for {
		err := claim.Setup(agentType)
		switch err {
		case ErrWaitingForUserVisit, ErrWaitingForUser:
			<-time.After(time.Millisecond * 200)
			continue
		case ErrUserRejected:
			return err
		case nil:
			return errors.New("unexpected status")
		}
		if err != ErrUserAccepted {
			return err
		}
		break // User accepted
	}

	secretCode, err := claim.Exchange()
	if err != nil {
		return err
	}
	claim.api.Secret = secretCode

	return nil
}
