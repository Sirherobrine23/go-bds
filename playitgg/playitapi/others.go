package playitapi

import "net/netip"

type TOTPStatus struct {
	Type      string `json:"status"`
	Epoch_sec uint32 `json:"epoch_sec,omitempty"`
}

type WebAuth struct {
	UpdateVersion uint32        `json:"update_version"`
	AccountID     uint64        `json:"account_id"`
	AdminID       uint64        `json:"admin_id"`
	Timestamp     uint64        `json:"timestamp"`
	AccountStatus AccountStatus `json:"account_status"`
	TOTP          TOTPStatus    `json:"totp_status"`
}

type WebSession struct {
	Key  string  `json:"session_key"` // Session key
	Auth WebAuth `json:"auth"`
}

func (api Api) LoginGuest() (WebSession, error) {
	info, _, err := RequestAPI[WebSession](api.Secret, "/login/guest", nil, nil)
	return info, err
}

type AgentVersion struct {
	Platform Platform `json:"platform"`
	Version  string   `json:"version"`
	Expired  string   `json:"has_expired"`
}

type PlayitAgentVersion struct {
	Official bool         `json:"official"`
	Website  string       `json:"details_website,omitempty"`
	Version  AgentVersion `json:"version"`
}

type ProtoRegister struct {
	Tunnel       netip.AddrPort     `json:"tunnel_addr"`
	Client       netip.AddrPort     `json:"client_addr"`
	AgentVersion PlayitAgentVersion `json:"agent_version"`
}

func (api Api) ProtoRegister(proto ProtoRegister) (string, error) {
	info, _, err := RequestAPI[struct {
		Key string `json:"key"`
	}](api.Secret, "/proto/register", proto, nil)

	return info.Key, err
}
