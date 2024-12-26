package playitapi

var (
	PlayitAPI     string = "https://api.playit.gg" // Playit API
	PlayitVersion string = "0.17.0"                // Playit API version
)

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

// API Account status
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

const (
	_            PortType = iota
	PortTypeBoth          // Support both Protocols
	PortTypeTcp           // Only support TCP
	PortTypeUdp           // Only support UDP
)

type PortType int

func (port PortType) MarshalText() ([]byte, error) {
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
func (port *PortType) UnmarshalText(text []byte) error {
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

const (
	_                          TunnelType = iota << 3
	TunnelTypeMinecraftJava               // Minecraft java server
	TunnelTypeMinecraftBedrock            // Minecraft Bedrock server
	TunnelTypeValheim                     // valheim
	TunnelTypeTerraria                    // Terraria multiplayer
	TunnelTypeStarbound                   // starbound
	TunnelTypeRust                        // Rust (No programmer language)
	TunnelType7Days                       // 7days
	TunnelTypeUnturned                    // unturned
)

// Set tunnel type to playit.gg
type TunnelType int

func (tun TunnelType) MarshalText() ([]byte, error) {
	switch tun {
	case TunnelTypeMinecraftJava:
		return []byte("minecraft-java"), nil
	case TunnelTypeMinecraftBedrock:
		return []byte("minecraft-bedrock"), nil
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

const (
	RegionGlobal       Region = iota << 1 // Free account and premium
	RegionSmartGlobal                     // Require premium account
	RegionNorthAmerica                    // Require premium account
	RegionEurope                          // Require premium account
	RegionAsia                            // Require premium account
	RegionIndia                           // Require premium account
	RegionSouthAmerica                    // Require premium account
)

type Region int

func (region Region) String() string {
	switch region {
	case RegionSmartGlobal:
		return "smart-global"
	case RegionNorthAmerica:
		return "north-america"
	case RegionEurope:
		return "europe"
	case RegionAsia:
		return "asia"
	case RegionIndia:
		return "india"
	case RegionSouthAmerica:
		return "south-america"
	default:
		return "global"
	}
}
func (region Region) MarshalText() ([]byte, error) { return []byte(region.String()), nil }
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
	PlatformUnknown         Platform = iota // unknown
	PlatformLinux                           // linux
	PlatformFreebsd                         // freebsd
	PlatformWindows                         // windows
	PlatformMacos                           // macos
	PlatformAndroid                         // android
	PlatformIos                             // ios
	PlatformDocker                          // docker
	PlatformMinecraftPlugin                 // minecraft-plugin
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
	case "macos", "darwin":
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

const (
	_ AgentType = iota
	AgentTypeDefault
	AgentTypeAssignable
	AgentTypeSelfManaged
)

type AgentType int

func (agent AgentType) MarshalText() ([]byte, error) {
	switch agent {
	case AgentTypeDefault:
		return []byte("default"), nil
	case AgentTypeAssignable:
		return []byte("assignable"), nil
	case AgentTypeSelfManaged:
		return []byte("self-managed"), nil
	}
	return []byte{}, nil
}

func (agent *AgentType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "default":
		*agent = AgentTypeDefault
	case "assignable":
		*agent = AgentTypeAssignable
	case "self-managed":
		*agent = AgentTypeSelfManaged
	}
	return nil
}

type TunnelDomainSource int

const (
	_                             TunnelDomainSource = iota
	TunnelDomainSourceFromIP                         // from-ip
	TunnelDomainSourceFromTunnel                     // from-tunnel
	TunnelDomainSourceFromAgentIP                    // from-agent-ip
)

func (source *TunnelDomainSource) UnmarshalText(text []byte) error {
	switch string(text) {
	case "from-ip":
		*source = TunnelDomainSourceFromIP
	case "from-tunnel":
		*source = TunnelDomainSourceFromTunnel
	case "from-agent-ip":
		*source = TunnelDomainSourceFromAgentIP
	}
	return nil
}

func (source TunnelDomainSource) MarshalText() ([]byte, error) {
	switch source {
	case TunnelDomainSourceFromIP:
		return []byte("from-ip"), nil
	case TunnelDomainSourceFromTunnel:
		return []byte("from-tunnel"), nil
	case TunnelDomainSourceFromAgentIP:
		return []byte("from-agent-ip"), nil
	}
	return []byte{}, nil
}

type DisablesReason int

const (
	_                             DisablesReason = iota
	DisablesReasonRequiresPremium                // requires-premium
	DisablesReasonOverPortLimit                  // over-port-limit
	DisablesReasonIPUsedInGRE                    // ip-used-in-gre
)

func (reason *DisablesReason) UnmarshalText(text []byte) error {
	switch string(text) {
	case "requires-premium":
		*reason = DisablesReasonRequiresPremium
	case "over-port-limit":
		*reason = DisablesReasonOverPortLimit
	case "ip-used-in-gre":
		*reason = DisablesReasonIPUsedInGRE
	}
	return nil
}

func (reason DisablesReason) MarshalText() ([]byte, error) {
	switch reason {
	case DisablesReasonRequiresPremium:
		return []byte("requires-premium"), nil
	case DisablesReasonOverPortLimit:
		return []byte("over-port-limit"), nil
	case DisablesReasonIPUsedInGRE:
		return []byte("ip-used-in-gre"), nil
	}
	return []byte{}, nil
}
