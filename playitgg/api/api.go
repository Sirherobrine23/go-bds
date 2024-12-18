package api

import (
	"errors"
)

var (
	ErrInvalidAuth            error = errors.New("cannot process api request because return not auth")
	ErrTooMenyRequest         error = errors.New("wait seconds to make new request")
	ErrNotFound               error = errors.New("path request not found")
	ErrAuthRequired           error = errors.New("auth required")
	ErrInvalidHeader          error = errors.New("invalid header")
	ErrInvalidSignature       error = errors.New("invalid signature")
	ErrInvalidTimestamp       error = errors.New("invalid timestamp")
	ErrInvalidApiKey          error = errors.New("invalid api key")
	ErrInvalidAgentKey        error = errors.New("invalid agent key")
	ErrSessionExpired         error = errors.New("session expired")
	ErrInvalidAuthType        error = errors.New("invalid auth type")
	ErrScopeNotAllowed        error = errors.New("scope not allowed")
	ErrNoLongerValid          error = errors.New("no longer valid")
	ErrGuestAccountNotAllowed error = errors.New("guest account not allowed")
	ErrEmailMustBeVerified    error = errors.New("email must be verified")
	ErrAccountDoesNotExist    error = errors.New("account does not exist")
	ErrAdminOnly              error = errors.New("admin only")
	ErrInvalidToken           error = errors.New("invalid token")
	ErrTotpRequred            error = errors.New("totp requred")
)

const (
	GoPlayitVersion string = "0.17.1"
	PlayitAPI       string = "https://api.playit.gg" // Playit API
)

const (
	TunnelTypeMinecraftBedrock TunnelType = iota + 1 // Minecraft Bedrock server
	TunnelTypeMinecraftJava                          // Minecraft java server
	TunnelTypeValheim                                // valheim
	TunnelTypeTerraria                               // Terraria multiplayer
	TunnelTypeStarbound                              // starbound
	TunnelTypeRust                                   // Rust (No programmer language)
	TunnelType7Days                                  // 7days
	TunnelTypeUnturned                               // unturned
)

const (
	_            PortProto = iota // Tunnel support tcp and udp protocol
	PortTypeBoth                  // Tunnel support tcp and udp protocol
	PortTypeTcp                   // Tunnel support only tcp protocol
	PortTypeUdp                   // Tunnel support only udp protocol
)

const (
	RegionGlobal       Region = iota + 1 // Free account and premium
	RegionSmartGlobal                    // Require premium account
	RegionNorthAmerica                   // Require premium account
	RegionEurope                         // Require premium account
	RegionAsia                           // Require premium account
	RegionIndia                          // Require premium account
	RegionSouthAmerica                   // Require premium account
)

type errStatus struct {
	Status string `json:"status"`
	Data   struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"data"`
}

func (err errStatus) Error() error {
	switch err.Data.Type {
	case "internal", "validation":
		return errors.New(err.Data.Message)
	case "path-not-found":
		return ErrNotFound
	case "auth":
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
		}
	}
	return errors.New(err.Data.Message)
}

type Api struct {
	Code   string `json:"-"`                // Claim code
	Secret string `json:"secret,omitempty"` // Agent Secret
}

const (
	_                      AccountStatus = iota
	StatusReady                          // "ready"
	StatusGuest                          // "guest"
	StatusBanned                         // "banned"
	StatusEmailNotVerified               // "email-not-verified"
	StatusHasMessage                     // "has-message"
	StatusDelete                         // "account-delete-scheduled"
	StatusOverLimit                      // "agent-over-limit"
	StatusDisabled                       // "agent-disabled"
)

type AccountStatus int

func (account *AccountStatus) UnmarshalText(text []byte) error {
	switch string(text) {
	case "ready":
		*account = StatusReady
	case "guest":
		*account = StatusGuest
	case "banned":
		*account = StatusBanned
	case "email-not-verified":
		*account = StatusEmailNotVerified
	case "has-message":
		*account = StatusHasMessage
	case "account-delete-scheduled":
		*account = StatusDelete
	case "agent-over-limit":
		*account = StatusOverLimit
	case "agent-disabled":
		*account = StatusDisabled
	}
	return nil
}

func (account AccountStatus) MarshalText() ([]byte, error) {
	switch account {
	case StatusReady:
		return []byte("ready"), nil
	case StatusGuest:
		return []byte("guest"), nil
	case StatusBanned:
		return []byte("banned"), nil
	case StatusEmailNotVerified:
		return []byte("email-not-verified"), nil
	case StatusHasMessage:
		return []byte("has-message"), nil
	case StatusDelete:
		return []byte("account-delete-scheduled"), nil
	case StatusOverLimit:
		return []byte("agent-over-limit"), nil
	case StatusDisabled:
		return []byte("agent-disabled"), nil
	}
	return []byte{}, nil
}

type PortProto int

func (port PortProto) MarshalText() ([]byte, error) {
	switch port {
	case PortTypeTcp:
		return []byte("tcp"), nil
	case PortTypeUdp:
		return []byte("udp"), nil
	case PortTypeBoth:
		return []byte("both"), nil
	}
	return []byte{}, nil
}
func (port *PortProto) UnmarshalText(text []byte) error {
	switch string(text) {
	case "tcp":
		*port = PortTypeTcp
	case "udp":
		*port = PortTypeUdp
	case "both":
		*port = PortTypeBoth
	}
	return nil
}

type TunnelType int

func (tun TunnelType) MarshalText() ([]byte, error) {
	switch tun {
	case TunnelTypeMinecraftBedrock:
		return []byte("minecraft-bedrock"), nil
	case TunnelTypeMinecraftJava:
		return []byte("minecraft-java"), nil
	case TunnelTypeValheim:
		return []byte("valheim"), nil
	case TunnelTypeTerraria:
		return []byte("terraria"), nil
	case TunnelTypeStarbound:
		return []byte("starbound"), nil
	case TunnelTypeRust:
		return []byte("rust"), nil
	case TunnelType7Days:
		return []byte("7days"), nil
	case TunnelTypeUnturned:
		return []byte("unturned"), nil
	default:
		return []byte(""), nil
	}
}
func (tun *TunnelType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "minecraft-bedrock":
		*tun = TunnelTypeMinecraftBedrock
	case "minecraft-java":
		*tun = TunnelTypeMinecraftJava
	case "valheim":
		*tun = TunnelTypeValheim
	case "terraria":
		*tun = TunnelTypeTerraria
	case "starbound":
		*tun = TunnelTypeStarbound
	case "rust":
		*tun = TunnelTypeRust
	case "7days":
		*tun = TunnelType7Days
	case "unturned":
		*tun = TunnelTypeUnturned
	default:
		*tun = 0
	}
	return nil
}

type Region int

func (region Region) MarshalText() ([]byte, error) {
	switch region {
	case RegionSmartGlobal:
		return []byte("smart-global"), nil
	case RegionNorthAmerica:
		return []byte("north-america"), nil
	case RegionEurope:
		return []byte("europe"), nil
	case RegionAsia:
		return []byte("asia"), nil
	case RegionIndia:
		return []byte("india"), nil
	case RegionSouthAmerica:
		return []byte("south-america"), nil
	default:
		return []byte("global"), nil
	}
}
func (region *Region) UnmarshalText(text []byte) error {
	switch string(text) {
	case "smart-global":
		*region = RegionSmartGlobal
	case "north-america":
		*region = RegionNorthAmerica
	case "europe":
		*region = RegionEurope
	case "asia":
		*region = RegionAsia
	case "india":
		*region = RegionIndia
	case "south-america":
		*region = RegionSouthAmerica
	default:
		*region = RegionGlobal
	}
	return nil
}

const (
	ProxyProtocolV1 ProxyProtocol = iota
	ProxyProtocolV2
)

type ProxyProtocol int

func (proxy *ProxyProtocol) UnmarshalText(text []byte) error {
	switch string(text) {
	case "proxy-protocol-v1":
		*proxy = ProxyProtocolV1
	case "proxy-protocol-v2":
		*proxy = ProxyProtocolV2
	}
	return nil
}

func (proxy ProxyProtocol) MarshalText() ([]byte, error) {
	switch proxy {
	case ProxyProtocolV1:
		return []byte("proxy-protocol-v1"), nil
	case ProxyProtocolV2:
		return []byte("proxy-protocol-v2"), nil
	}
	return []byte{}, nil
}

const (
	AgentTunnelDisabledByUser AgentTunnelDisabled = iota
	AgentTunnelDisabledBySystem
)

type AgentTunnelDisabled int

func (tun *AgentTunnelDisabled) UnmarshalText(text []byte) error {
	switch string(text) {
	case "0", "ByUser":
		*tun = AgentTunnelDisabledByUser
	case "1", "BySystem":
		*tun = AgentTunnelDisabledBySystem
	}
	return nil
}

func (tun AgentTunnelDisabled) MarshalText() ([]byte, error) {
	switch tun {
	case AgentTunnelDisabledByUser:
		return []byte("ByUser"), nil
	case AgentTunnelDisabledBySystem:
		return []byte("BySystem"), nil
	}
	return []byte{}, nil
}

const (
	_                       Platform = iota
	PlatformUnknown                  // unknown
	PlatformLinux                    // linux
	PlatformFreebsd                  // freebsd
	PlatformWindows                  // windows
	PlatformMacos                    // macos
	PlatformAndroid                  // android
	PlatformIos                      // ios
	PlatformDocker                   // docker
	PlatformMinecraftPlugin          // minecraft-plugin
)

type Platform int

func (plat *Platform) UnmarshalText(text []byte) error {
	switch string(text) {
	case "linux":
		*plat = PlatformLinux
	case "freebsd":
		*plat = PlatformFreebsd
	case "windows":
		*plat = PlatformWindows
	case "macos":
		*plat = PlatformMacos
	case "android":
		*plat = PlatformAndroid
	case "ios":
		*plat = PlatformIos
	case "docker":
		*plat = PlatformDocker
	case "minecraft-plugin":
		*plat = PlatformMinecraftPlugin
	default:
		*plat = PlatformUnknown
	}
	return nil
}

func (plat Platform) MarshalText() ([]byte, error) {
	switch plat {
	case PlatformLinux:
		return []byte("linux"), nil
	case PlatformFreebsd:
		return []byte("freebsd"), nil
	case PlatformWindows:
		return []byte("windows"), nil
	case PlatformMacos:
		return []byte("macos"), nil
	case PlatformAndroid:
		return []byte("android"), nil
	case PlatformIos:
		return []byte("ios"), nil
	case PlatformDocker:
		return []byte("docker"), nil
	case PlatformMinecraftPlugin:
		return []byte("minecraft-plugin"), nil
	default:
		return []byte("unknown"), nil
	}
}
