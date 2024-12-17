package api

type PortType int

func (port PortType) String() string {
	switch port {
	case PortTypeTcp:
		return "tcp"
	case PortTypeUdp:
		return "udp"
	default:
		return "both"
	}
}

func (port PortType) MarshalText() ([]byte, error) { return []byte(port.String()), nil }
func (port *PortType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "tcp":
		*port = PortTypeTcp
	case "udp":
		*port = PortTypeUdp
	default:
		*port = PortTypeBoth
	}
	return nil
}

// Tunnel type
type TunnelType int

func (tun TunnelType) String() string {
	switch tun {
	case TunnelTypeMCBedrock:
		return "minecraft-bedrock"
	case TunnelTypeMCJava:
		return "minecraft-java"
	case TunnelTypeValheim:
		return "valheim"
	case TunnelTypeTerraria:
		return "terraria"
	case TunnelTypeStarbound:
		return "starbound"
	case TunnelTypeRust:
		return "rust"
	case TunnelType7Days:
		return "7days"
	case TunnelTypeUnturned:
		return "unturned"
	default:
		return ""
	}
}

func (tun TunnelType) MarshalText() ([]byte, error) { return []byte(tun.String()), nil }
func (tun *TunnelType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "minecraft-bedrock":
		*tun = TunnelTypeMCBedrock
	case "minecraft-java":
		*tun = TunnelTypeMCJava
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
	}
	return nil
}

// Regions allocted
type AllocationRegion int

func (aloc AllocationRegion) String() string {
	switch aloc {
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

func (aloc AllocationRegion) MarshalText() ([]byte, error) { return []byte(aloc.String()), nil }
func (aloc *AllocationRegion) UnmarshalText(text []byte) error {
	switch string(text) {
	case "smart-global":
		*aloc = RegionSmartGlobal
	case "north-america":
		*aloc = RegionNorthAmerica
	case "europe":
		*aloc = RegionEurope
	case "asia":
		*aloc = RegionAsia
	case "india":
		*aloc = RegionIndia
	case "south-america":
		*aloc = RegionSouthAmerica
	default:
		*aloc = RegionGlobal
	}
	return nil
}
