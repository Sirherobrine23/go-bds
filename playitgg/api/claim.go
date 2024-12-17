package api

import (
	"net/netip"

	"github.com/google/uuid"
)

const (
	ClaimSetupResponseWaitingForUserVisit ClaimSetupResponse = iota
	ClaimSetupResponseWaitingForUser
	ClaimSetupResponseUserAccepted
	ClaimSetupResponseUserRejected
)

type AgentClaimDetails struct {
	Name      string     `json:"name"`
	RemoteIP  netip.Addr `json:"remote_ip"`
	AgentType string     `json:"agent_type"`
	Version   string     `json:"version"`
}

type ClaimSetupResponse int

func (client Client) ClaimDetails(code string) (*AgentClaimDetails, error) {
	return postRequest[*AgentClaimDetails]("/claim/details", client.ClientToken, Map{"code": code}, nil)
}

func (client Client) ClaimExchange(code string) (string, error) {
	d, err := postRequest[struct {
		Key string `json:"secret_key"`
	}]("/claim/exchange", client.ClientToken, Map{"code": code}, nil)
	return d.Key, err
}

func (client Client) ClaimAccept(code, agent, version string) (uuid.UUID, error) {
	type Accept struct {ID uuid.UUID `json:"agent_id"`}
	d, err := postRequest[Accept]("/claim/accept", client.ClientToken, Map{"code": code, "agent_type": agent, "version": version}, nil)
	return d.ID, err
}
func (client Client) ClaimReject(code string) error {
	_, err := postRequest[any]("/claim/reject", client.ClientToken, Map{"code": code}, nil)
	return err
}

func (client Client) ClaimSetup(code, agent, version string) (ClaimSetupResponse, error) {
	return postRequest[ClaimSetupResponse]("/claim/setup", client.ClientToken, Map{"code": code, "agent_type": agent, "version": version}, nil)
}

func (ClaimResponse ClaimSetupResponse) String() string {
	switch ClaimResponse {
	case ClaimSetupResponseWaitingForUserVisit:
		return "Waiting For User Visit"
	case ClaimSetupResponseWaitingForUser:
		return "Waiting For User"
	case ClaimSetupResponseUserAccepted:
		return "User Accepted"
	case ClaimSetupResponseUserRejected:
		return "User Rejected"
	default:
		return ""
	}
}

func (ClaimResponse ClaimSetupResponse) MarshalText() ([]byte, error) {
	return []byte(ClaimResponse.String()), nil
}
func (ClaimResponse *ClaimSetupResponse) UnmarshalText(text []byte) error {
	switch string(text) {
	case "WaitingForUserVisit":
		*ClaimResponse = ClaimSetupResponseWaitingForUserVisit
	case "WaitingForUser":
		*ClaimResponse = ClaimSetupResponseWaitingForUser
	case "UserAccepted":
		*ClaimResponse = ClaimSetupResponseUserAccepted
	case "UserRejected":
		*ClaimResponse = ClaimSetupResponseUserRejected
	}
	return nil
}
