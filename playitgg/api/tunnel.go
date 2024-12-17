package api

import (
	"net/netip"

	"github.com/google/uuid"
)

// Create tunnel
func (client Client) CreateTunnel(tun TunnelCreate) (*ObjectID, error) {
	return postRequest[*ObjectID]("/tunnels/create", client.ClientToken, tun, nil)
}

func (client Client) DeleteTunnel(tunID uuid.UUID) error {
	_, err := postRequest[any]("/tunnels/delete", client.ClientToken, Map{"tunnel_id": tunID}, nil)
	return err
}

type TunnelCreate struct {
	Name       string                    `json:"name,omitempty"`
	TunType    TunnelType                `json:"tunnel_type,omitempty"`
	PortType   PortType                  `json:"port_type"`
	PortCount  uint16                    `json:"port_count"`
	Origin     TunnelOriginCreate        `json:"origin"`
	Enabled    bool                      `json:"enabled"`
	Alloc      TunnelCreateUseAllocation `json:"alloc"`
	FirewallID uuid.UUID                 `json:"firewall_id"`
}

type TunnelOriginCreate struct {
	Type string                 `json:"type"`
	Data TunnelOriginCreateData `json:"data"`
}

type TunnelOriginCreateData struct {
	LocalIP   netip.Addr `json:"local_ip,omitempty"`
	LocalPort uint16     `json:"local_port,omitempty"`
	AgentID   uuid.UUID  `json:"agend_id,omitempty"`
}

type TunnelCreateUseAllocation struct {
	Type    string `json:"type"`
	Details struct {
		*TunnelCreateUseAllocationDedicatedIp
		*TunnelCreateUseAllocationPortAllocation
		*TunnelCreateUseAllocationRegion
	} `json:"details"`
}

type TunnelCreateUseAllocationDedicatedIp struct {
	IpHostname string `json:"ip_hostname"`
	Port       uint16 `json:"port,omitempty"`
}

type TunnelCreateUseAllocationPortAllocation struct {
	ID uuid.UUID `json:"alloc_id"`
}

type TunnelCreateUseAllocationRegion struct {
	Region AllocationRegion `json:"region"`
}
