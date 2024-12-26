package playitapi

import (
	"encoding/json"
	"net"
	"time"

	"github.com/google/uuid"
)

type ReqTunnelsCreate struct {
	Name       string                     `json:"name,omitempty"`
	TunnelType TunnelType                 `json:"tunnel_type,omitempty"`
	PortType   PortType                   `json:"port_type"`
	PortCount  uint16                     `json:"port_count"`
	Origin     TunnelOriginCreate         `json:"origin"`
	Enabled    bool                       `json:"enabled"`
	Alloc      *TunnelCreateUseAllocation `json:"alloc,omitempty"`
	Firewall   *uuid.UUID                 `json:"firewall_id,omitempty"`
}

type TunnelOriginCreateDefault struct {
	LocalIP   net.IP `json:"local_ip"`
	LocalPort uint16 `json:"local_port,omitempty"`
}
type TunnelOriginCreateAgent struct {
	AgentID   uuid.UUID `json:"agent_id"`
	LocalIP   net.IP    `json:"local_ip"`
	LocalPort uint16    `json:"local_port,omitempty"`
}
type TunnelOriginCreateManaged struct {
	ManagedID *uuid.UUID `json:"managed_id,omitempty"`
}

type TunnelOriginCreate struct {
	*TunnelOriginCreateDefault
	*TunnelOriginCreateAgent
	*TunnelOriginCreateManaged
}

func (origin *TunnelOriginCreate) MarshalJSON() ([]byte, error) {
	if origin.TunnelOriginCreateDefault != nil {
		return json.Marshal(map[string]interface{}{"type": "default", "data": origin.TunnelOriginCreateDefault})
	} else if origin.TunnelOriginCreateAgent != nil {
		return json.Marshal(map[string]interface{}{"type": "agent", "data": origin.TunnelOriginCreateAgent})
	} else if origin.TunnelOriginCreateManaged != nil {
		return json.Marshal(map[string]interface{}{"type": "managed", "data": origin.TunnelOriginCreateManaged})
	}
	return []byte{}, nil
}

func (origin *TunnelOriginCreate) UnmarshalJSON(data []byte) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	switch string(obj["type"]) {
	case `"default"`:
		origin.TunnelOriginCreateDefault = &TunnelOriginCreateDefault{}
		return json.Unmarshal(obj["data"], origin.TunnelOriginCreateDefault)
	case `"agent"`:
		origin.TunnelOriginCreateAgent = &TunnelOriginCreateAgent{}
		return json.Unmarshal(obj["data"], origin.TunnelOriginCreateAgent)
	case `"managed"`:
		origin.TunnelOriginCreateManaged = &TunnelOriginCreateManaged{}
		return json.Unmarshal(obj["data"], origin.TunnelOriginCreateManaged)
	default:
		return nil
	}
}

type TunnelCreateUseAllocationDedicatedIp struct {
	Hostname_IP string `json:"ip_hostname"`
	Port        uint16 `json:"port,omitempty"`
}
type TunnelCreateUseAllocationPortAllocation struct {
	ID uuid.UUID `json:"alloc_id"`
}
type TunnelCreateUseAllocationRegion struct {
	Region Region `json:"region"`
}

type TunnelCreateUseAllocation struct {
	*TunnelCreateUseAllocationDedicatedIp
	*TunnelCreateUseAllocationPortAllocation
	*TunnelCreateUseAllocationRegion
}

func (alloc *TunnelCreateUseAllocation) MarshalJSON() ([]byte, error) {
	if alloc.TunnelCreateUseAllocationDedicatedIp != nil {
		return json.Marshal(map[string]interface{}{"type": "dedicated-ip", "details": alloc.TunnelCreateUseAllocationDedicatedIp})
	} else if alloc.TunnelCreateUseAllocationPortAllocation != nil {
		return json.Marshal(map[string]interface{}{"type": "port-allocation", "details": alloc.TunnelCreateUseAllocationPortAllocation})
	} else if alloc.TunnelCreateUseAllocationRegion != nil {
		return json.Marshal(map[string]interface{}{"type": "region", "details": alloc.TunnelCreateUseAllocationRegion})
	}
	return []byte{}, nil
}

func (alloc *TunnelCreateUseAllocation) UnmarshalJSON(data []byte) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	switch string(obj["type"]) {
	case `"dedicated-ip"`:
		alloc.TunnelCreateUseAllocationDedicatedIp = &TunnelCreateUseAllocationDedicatedIp{}
		return json.Unmarshal(obj["details"], alloc.TunnelCreateUseAllocationDedicatedIp)
	case `"port-allocation"`:
		alloc.TunnelCreateUseAllocationPortAllocation = &TunnelCreateUseAllocationPortAllocation{}
		return json.Unmarshal(obj["details"], alloc.TunnelCreateUseAllocationPortAllocation)
	case `"region"`:
		alloc.TunnelCreateUseAllocationRegion = &TunnelCreateUseAllocationRegion{}
		return json.Unmarshal(obj["details"], alloc.TunnelCreateUseAllocationRegion)
	default:
		return nil
	}
}

func (api Api) CreateTunnel(TunConfig *ReqTunnelsCreate) (uuid.UUID, error) {
	id, _, err := RequestAPI[objectID](api.Secret, "/tunnels/create", TunConfig, nil)
	return id.ID, err
}

func (api Api) DeleteTunnel(id uuid.UUID) error {
	_, _, err := RequestAPI[objectID](api.Secret, "/tunnels/delete", map[string]string{"tunnel_id": id.String()}, nil)
	return err
}

type ProtoAlloc struct {
	Allowd  uint16 `json:"allowed"`
	Claimed uint16 `json:"claimed"`
	Desired uint16 `json:"desired"`
}

type TunnelAllocData struct {
	ID         uuid.UUID `json:"id"`
	IPv4       net.IP    `json:"static_ip4"`
	IPv6       net.IP    `json:"static_ip6"`
	TunnelAddr net.IP    `json:"tunnel_ip"`
	Hostname   string    `json:"ip_hostname"`
	Domain     string    `json:"assigned_domain"`
	SRV        string    `json:"assigned_srv"`
	PortStart  uint16    `json:"port_start"`
	PortEnd    uint16    `json:"port_end"`
	IPType     PortType  `json:"ip_type"`
	Region     Region    `json:"region"`
	Assignment struct {
		Type string `json:"type"`
	} `json:"assignment"`
}

type TunnelAlloc struct {
	Pedding  bool
	Disabled bool
	*TunnelAllocData
}

func (alloc TunnelAlloc) MarshalJSON() ([]byte, error) {
	if alloc.Pedding {
		return json.Marshal(map[string]interface{}{"status": "pending"})
	} else if alloc.Disabled {
		return json.Marshal(map[string]interface{}{"status": "disabled"})
	} else {
		return json.Marshal(alloc.TunnelAllocData)
	}
}

func (alloc *TunnelAlloc) UnmarshalJSON(data []byte) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	switch string(obj["status"]) {
	case `"pending"`:
		alloc.Pedding = true
		return nil
	case `"disabled"`:
		alloc.Disabled = true
	case `"allocated"`:
		alloc.Pedding = false
		alloc.Disabled = false
		alloc.TunnelAllocData = &TunnelAllocData{}
		return json.Unmarshal(obj["data"], alloc.TunnelAllocData)
	}
	return nil
}

type Tunnel struct {
	ID             uuid.UUID          `json:"id"`
	Name           string             `json:"name"`
	Enabled        bool               `json:"active"`
	TunnelType     TunnelType         `json:"tunnel_type"`
	CreatedAt      time.Time          `json:"created_at"`
	PortType       PortType           `json:"port_type"`
	PortCount      uint16             `json:"port_count"`
	Region         Region             `json:"region"`
	ProxyProtocol  ProxyProtocol      `json:"proxy_protocol"`
	FirewallID     *uuid.UUID         `json:"firewall_id"`
	Origin         TunnelOriginCreate `json:"origin"`
	Alloc          TunnelAlloc        `json:"alloc"`
	DisablesReason DisablesReason     `json:"disabled_reason"`
	Domain         *struct {
		ID       uuid.UUID          `json:"id"`
		Name     string             `json:"name"`
		Parent   string             `json:"parent"`
		External bool               `json:"is_external"`
		Source   TunnelDomainSource `json:"source"`
	} `json:"domain"`
	ExpireNotice *struct {
		Remove  time.Time `json:"remove_at"`
		Disable time.Time `json:"disable_at"`
	} `json:"expire_notice"`
	RateLimit struct {
		BytesPerSecond   uint64 `json:"bytes_per_second"`
		PacketsPerSecond uint64 `json:"packets_per_second"`
	} `json:"ratelimit"`
}

type Tunnels struct {
	TCP     ProtoAlloc `json:"tcp_alloc"` // TCP Allocations
	UDP     ProtoAlloc `json:"udp_alloc"` // UDP Allocations
	Tunnels []Tunnel   `json:"tunnels"`   // Tunnels
}

func (api Api) Tunnels(Tunnel, Agent uuid.UUID) (*Tunnels, error) {
	tuns, _, err := RequestAPI[*Tunnels](api.Secret, "/tunnels/list", struct {
		Tunnel uuid.NullUUID `json:"tunnel_id,omitempty"`
		Agent  uuid.NullUUID `json:"agent_id,omitempty"`
	}{uuid.NullUUID{UUID: Tunnel}, uuid.NullUUID{UUID: Agent}}, nil)
	return tuns, err
}
