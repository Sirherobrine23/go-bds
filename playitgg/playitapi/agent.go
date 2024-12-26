package playitapi

import (
	"net/netip"

	"github.com/google/uuid"
)

type AgentPendingTunnel struct {
	ID        uuid.UUID  `json:"id"`
	Disabled  bool       `json:"is_disabled"`
	Name      string     `json:"name"`
	Proto     PortType   `json:"proto"`
	PortCount uint16     `json:"port_count"`
	Tunnel    TunnelType `json:"tunnel_type"`
}

/*
	pub struct AgentTunnel {
		pub disabled: Option<AgentTunnelDisabled>,
	}
*/
type AgentTunnel struct {
	ID             uuid.UUID           `json:"id"`
	Name           string              `json:"name"`
	Ip             uint64              `json:"ip_num"`
	Region         uint64              `json:"region_num"`
	Proto          PortType            `json:"proto"`
	LocalIP        netip.Addr          `json:"local_ip"`
	LocalPort      uint16              `json:"local_port"`
	TunnelType     TunnelType          `json:"tunnel_type"`
	AssignedDomain string              `json:"assigned_domain"`
	CustomDomain   string              `json:"custom_domain"`
	ProxyProtocol  ProxyProtocol       `json:"proxy_protocol"`
	Disabled       AgentTunnelDisabled `json:"disabled"`
	Port           struct {
		From uint16 `json:"from"`
		To   uint16 `json:"to"`
	} `json:"port"`
}

type Rundata struct {
	ID              uuid.UUID            `json:"agent_id"`
	Type            AgentType            `json:"agent_type"`
	AccountStatus   AccountStatus        `json:"account_status"`
	Tunnels         []AgentTunnel        `json:"tunnels"`
	Peddings        []AgentPendingTunnel `json:"peddings"`
	AccountFeatures struct {
		RegionalTunnels bool `json:"regional_tunnels"`
	} `json:"account_features"`
}

func (api Api) RundataAgents() (*Rundata, error) {
	info, _, err := RequestAPI[*Rundata](api.Secret, "/agents/rundata", nil, nil)
	return info, err
}

type AgentRouting struct {
	ID          uuid.UUID    `json:"agent_id"`
	TargetsIPv4 []netip.Addr `json:"targets4"`
	TargetsIPv6 []netip.Addr `json:"targets6"`
}

func (api Api) Routing(AgentID uuid.UUID) (*AgentRouting, error) {
	info, _, err := RequestAPI[*AgentRouting](api.Secret, "/agents/routing/get", struct {
		AgentID uuid.NullUUID `json:"agent_id"`
	}{uuid.NullUUID{UUID: AgentID}}, nil)
	return info, err
}
