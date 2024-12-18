package api

import (
	"net"
	"time"

	"github.com/google/uuid"
)

type AssignedDefaultCreate struct {
	Ip   net.IP  `json:"local_ip"`
	Port *uint16 `json:"local_port,omitempty"`
}

type AssignedAgentCreate struct {
	ID   uuid.UUID `json:"agent_id"`
	Ip   net.IP    `json:"local_ip"`
	Port *uint16   `json:"local_port,omitempty"`
}

type AssignedManagedCreate struct {
	ID *uuid.UUID `json:"agent_id,omitempty"`
}

type AgentMerged struct {
	*AssignedDefaultCreate
	*AssignedAgentCreate
	*AssignedManagedCreate
}

type TunnelOriginCreate struct {
	Type  string      `json:"type"` // Agent type: default, agent or managed
	Agent AgentMerged `json:"data"` // Assingned agent
}

type UseAllocDedicatedIp struct {
	IpHost string  `json:"ip_hostname"`
	Port   *uint16 `json:"port,omitempty"`
}

type UseAllocPortAlloc struct {
	ID uuid.UUID `json:"alloc_id"`
}

type UseRegion struct {
	Region string `json:"region"`
}

type TunnelCreateUseAllocationDetails struct {
	*UseRegion
	*UseAllocPortAlloc
	*UseAllocDedicatedIp
}

/*
*
"status": "allocated",

	"data": {
		"assigned_domain": "going-scales.gl.at.ply.gg",
		"assigned_srv": null,
		"assignment": {
			"type": "shared-ip"
		},
		"id": "f667b538-0294-4817-9332-5cba5e94d79e",
		"ip_hostname": "19.ip.gl.ply.gg",
		"ip_type": "both",
		"port_end": 49913,
		"port_start": 49912,
		"region": "global",
		"static_ip4": "147.185.221.19",
		"tunnel_ip": "2602:fbaf:0:1::13"
	}
*/
type TunnelCreateUseAllocation struct {
	Status string                           `json:"status"`  // For tunnel list
	Type   string                           `json:"type"`    // "dedicated-ip", "port-allocation" or "region"
	Data   TunnelCreateUseAllocationDetails `json:"details"` // UseAllocDedicatedIp, UseAllocPortAlloc, UseRegion
}

type Tunnel struct {
	ID         *uuid.UUID                 `json:"tunnel_id,omitempty"`   // Tunnel UUID
	Name       string                     `json:"name,omitempty"`        // Tunnel name
	TunnelType string                     `json:"tunnel_type,omitempty"` // Tunnel type from TunnelType const's
	PortType   PortProto                  `json:"port_type"`             // tcp, udp or both
	PortCount  uint16                     `json:"port_count"`            // Port count to assign to connect
	Origin     TunnelOriginCreate         `json:"origin"`
	Enabled    bool                       `json:"enabled"`
	Alloc      *TunnelCreateUseAllocation `json:"alloc,omitempty"`
	Firewall   *uuid.UUID                 `json:"firewall_id,omitempty"` // Firewall ID
}

func (w Api) CreateTunnel(tun Tunnel) (uuid.UUID, error) {
	type tunType struct {
		ID uuid.UUID `json:"id"`
	}
	tunnelId, _, err := requestAPI[tunType](w, "/tunnels/create", tun, nil)
	if err != nil {
		return [16]byte{}, err
	}
	tun.ID = &tunnelId.ID

	info, err := w.AgentInfo()
	if err != nil {
		return [16]byte{}, err
	}

	for {
		tuns, err := w.ListTunnels(tun.ID, &info.ID)
		if err != nil {
			return [16]byte{}, err
		}
		if tuns.Tunnels[0].Alloc.Status == "pending" {
			time.Sleep(time.Second * 2)
			continue
		}
		break
	}

	return tunnelId.ID, nil
}

func (w Api) DeleteTunnel(TunnelID *uuid.UUID) error {
	if TunnelID == nil {
		return nil
	}
	_, _, err := requestAPI[any](w, "/tunnels/delete", struct {
		TunnelID uuid.UUID `json:"tunnel_id"`
	}{*TunnelID}, nil)
	return err
}

type AccountTunnel struct {
	ID         uuid.UUID                 `json:"id"`
	TunnelType string                    `json:"tunnel_type"`
	CreatedAt  time.Time                 `json:"created_at"`
	Name       string                    `json:"name"`
	PortType   PortProto                 `json:"port_type"`
	PortCount  int32                     `json:"port_count"`
	Alloc      TunnelCreateUseAllocation `json:"alloc"`
	Origin     TunnelOriginCreate        `json:"origin"`
	Domain     *struct {
		ID         uuid.UUID `json:"id"`
		Name       string    `json:"name"`
		IsExternal bool      `json:"is_external"`
		Parent     string    `json:"parent"`
		Source     string    `json:"source"`
	} `json:"domain"`
	FirewallID string `json:"firewall_id"`
	Ratelimit  struct {
		BytesSecs   uint64 `json:"bytes_per_second"`
		PacketsSecs uint64 `json:"packets_per_second"`
	} `json:"ratelimit"`
	Active         bool   `json:"active"`
	DisabledReason string `json:"disabled_reason"`
	Region         string `json:"region"`
	ExpireNotice   *struct {
		Disable time.Time `json:"disable_at"`
		Remove  time.Time `json:"remove_at"`
	} `json:"expire_notice"`
}

type AlloctedPorts struct {
	Allowed uint16 `json:"allowed"`
	Claimed uint16 `json:"claimed"`
	Desired uint16 `json:"desired"`
}

type AccountTunnels struct {
	Tcp     AlloctedPorts   `json:"tcp_alloc"`
	Udp     AlloctedPorts   `json:"udp_alloc"`
	Tunnels []AccountTunnel `json:"tunnels"`
}

func (w Api) ListTunnels(TunnelID, AgentID *uuid.UUID) (*AccountTunnels, error) {
	type TunList struct {
		TunnelID *uuid.UUID `json:"tunnel_id,omitempty"`
		AgentID  *uuid.UUID `json:"agent_id,omitempty"`
	}

	Tuns, _, err := requestAPI[AccountTunnels](w, "/tunnels/list", TunList{TunnelID, AgentID}, nil)
	if err != nil {
		return nil, err
	}
	return &Tuns, nil
}
