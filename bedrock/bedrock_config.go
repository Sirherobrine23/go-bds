package bedrock

import "errors"

// Permission value
type PermissionLevel uint

func (p PermissionLevel) String() string {
	if int(p) < len(permisionName) {
		return permisionName[p]
	}
	return ""
}

func (p *PermissionLevel) UnmarshalText(text []byte) error {
	*p = Visitor
	switch string(text) {
	case permisionName[Member]:
		*p = Member
	case permisionName[Operator]:
		*p = Operator
	}
	return nil
}

func (p PermissionLevel) MarshalText() ([]byte, error) {
	if d := p.String(); d != "" {
		return []byte(d), nil
	}
	return nil, errors.New("invalid Permision")
}

type Permissions []Permission

// Player permission
type Permission struct {
	Permission PermissionLevel `json:"permission"`
	XUID       string          `json:"xuid"`
}

type AllowList []PlayerAllowList

// Player allow list
type PlayerAllowList struct {
	Name         string `json:"name"`               // Player name
	IgnoreLimits bool   `json:"ignoresPlayerLimit"` // True if this user should not count towards the maximum player limit. Currently there's another soft limit of 30 (or 1 higher than the specified number of max players) connected players, even if players use this option. The intention for this is to have some players be able to join even if the server is full.
	XUID         string `json:"xuid,omitempty"`     // Optional. The XUID of the user. If it's not set then it will be populated when someone with a matching name connects.
}
