package api

import "net/netip"

func (w Api) GetPings() ([]Ping, error) {
	type pings struct {
		Pings []Ping `json:"experiments"`
	}
	pingsData, _, err := requestAPI[pings](w, "/ping/get", nil, nil)
	if err != nil {
		return nil, err
	}
	return pingsData.Pings, nil
}

func (w Api) PingSubmit(Pings ...PingSubmit) error {
	type send struct {
		Results []PingSubmit `json:"results"`
	}
	_, _, err := requestAPI[any](w, "/ping/submit", send{Pings}, nil)
	return err
}

type PingTarget struct {
	IP   netip.Addr `json:"ip"`
	Port uint16     `json:"port"`
}

type PingSample struct {
	TunnelServerID uint64 `json:"tunnel_server_id"`
	DCID           uint64 `json:"dc_id"`
	ServerTS       uint64 `json:"sever_ts"`
	Latency        uint64 `json:"latency"`
	Count          uint16 `json:"count"`
	Num            uint16 `json:"num"`
}

type PingSubmit struct {
	ID      uint64       `json:"id"`
	Target  PingTarget   `json:"target"`
	Samples []PingSample `json:"samples"`
}

type Ping struct {
	ID           uint64       `json:"id"`
	TestInterval uint64       `json:"test_interval"`
	PingInterval uint64       `json:"ping_interval"`
	Samples      uint64       `json:"samples"`
	Targets      []PingTarget `json:"targets"`
}
