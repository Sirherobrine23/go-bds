package mojang

type Permission struct {
	Permission string `json:"permission"`
	XUID       string `json:"xuid"`
}

type PlayerAllowList struct {
	Name         string `json:"name"`               // Player name
	IgnoreLimits bool   `json:"ignoresPlayerLimit"` // True if this user should not count towards the maximum player limit. Currently there's another soft limit of 30 (or 1 higher than the specified number of max players) connected players, even if players use this option. The intention for this is to have some players be able to join even if the server is full.
	XUID         string `json:"xuid,omitempty"`     // Optional. The XUID of the user. If it's not set then it will be populated when someone with a matching name connects.
}

type AllowList []PlayerAllowList
type Permissions []Permission

type MojangConfig struct {
	// Used as the server name
	//
	// Allowed values: Any string without semicolon symbol.
	ServerName string `json:"serverName" properties:"server-name"`
	// Sets the game mode for new players.
	//
	// Allowed values: "survival", "creative", or "adventure"
	Gamemode string `json:"gamemode" properties:"gamemode"`
	// force-gamemode=false (or force-gamemode is not defined in the server.properties)
	// prevents the server from sending to the client gamemode values other
	// than the gamemode value saved by the server during world creation
	// even if those values are set in server.properties after world creation.
	//
	// force-gamemode=true forces the server to send to the client gamemode values
	// other than the gamemode value saved by the server during world creation
	// if those values are set in server.properties after world creation.
	ForceGamemode bool `json:"forceGamemode" properties:"force-gamemode"`
	// Sets the difficulty of the world.
	//
	// Allowed values: "peaceful", "easy", "normal", or "hard"
	Difficulty string `json:"difficulty" properties:"difficulty"`
	// If true then cheats like commands can be used.
	AllowCheats bool `json:"allowCheats" properties:"allow-cheats"`
	// The maximum number of players that can play on the server.
	MaxPlayers int64 `json:"maxPlayers" properties:"max-players"`
	// If true then all connected players must be authenticated to Xbox Live.
	// Clients connecting to remote (non-LAN) servers will always require Xbox Live authentication regardless of this setting.
	// If the server accepts connections from the Internet, then it's highly recommended to enable online-mode.
	OnlineMode bool `json:"onlineMode" properties:"online-mode"`
	// If true then all connected players must be listed in the separate allowlist.json file.
	AllowList bool `json:"allowList" properties:"allow-list"`
	// Which IPv4 port the server should listen to.
	Port int16 `json:"port" properties:"server-port"`
	// Which IPv6 port the server should listen to.
	Portv6 int16 `json:"portv6" properties:"server-portv6"`
	// Listen and respond to clients that are looking for servers on the LAN. This will cause the server
	// to bind to the default ports (19132, 19133) even when `server-port` and `server-portv6`
	// have non-default values. Consider turning this off if LAN discovery is not desirable, or when
	// running multiple servers on the same host may lead to port conflicts.
	EnableLanVisibility bool `json:"enableLan" properties:"enable-lan-visibility"`
	// The maximum allowed view distance in number of chunks.
	ViewDistance int64 `json:"viewDistance" properties:"view-distance"`
	// The world will be ticked this many chunks away from any player.
	//
	// Allowed values: Integers in the range [4, 12]
	TickDistance int `json:"tickDistance" properties:"tick-distance"`
	// After a player has idled for this many minutes they will be kicked. If set to 0 then players can idle indefinitely.
	PlayerTimeout uint64 `json:"playerTimeout" properties:"player-idle-timeout"`
	// Maximum number of threads the server will try to use. If set to 0 or removed then it will use as many as possible.
	Threads uint64 `json:"threads" properties:"max-threads"`
	// Allowed values: Any string without semicolon symbol or symbols illegal for file name: /\n\r\t\f`?*\\<>|\":
	LevelName string `json:"levelName" properties:"level-name"`
	// Use to randomize the world
	LevelSeed int64 `json:"levelSeed" properties:"level-seed"`
	// Permission level for new players joining for the first time.
	// Allowed values: "visitor", "member", "operator"
	DefaultPlayerPermission string `json:"defaultPlayerPermission" properties:"default-player-permission-level"`
	// Force clients to use texture packs in the current world
	RequireTexture bool `json:"requireTexture" properties:"texturepack-required"`
	// Enables logging content errors to a file
	EnableLogFile bool `json:"enebleLogFile" properties:"content-log-file-enabled"`
	// Determines the smallest size of raw network payload to compress
	CompressionThreshold int `json:"compressionThreshold" properties:"compression-threshold"`
	// Determines the compression algorithm to use for networking
	// Allowed values: "zlib", "snappy"
	CompressionAlgorithm string `json:"compressionAlgorithm" properties:"compression-algorithm"`
	// Allowed values: "client-auth", "server-auth", "server-auth-with-rewind"
	// Changes the server authority on movement:
	// "client-auth": Server has no authority and accepts all positions from the client.
	// "server-auth": Server takes user input and simulates the Player movement but accepts the Client version if there is disagreement.
	// "server-auth-with-rewind": The server will replay local user input and will push it to the Client so it can correct the position if it does not match.
	// 		The clients will rewind time back to the correction time, apply the correction, then replay all the player's inputs since then. This results in smoother and more frequent corrections.
	AuthoritativeMovement string `json:"authoritativeMovement" properties:"server-authoritative-movement"`
	// Only used with "server-auth-with-rewind".
	// This is the tolerance of discrepancies between the Client and Server Player position. This helps prevent sending corrections too frequently
	// for non-cheating players in cases where the server and client have different perceptions about when a motion started. For example damage knockback or being pushed by pistons.
	// The higher the number, the more tolerant the server will be before asking for a correction. Values beyond 1.0 have increased chances of allowing cheating.
	PlayerPositionAcceptanceThreshold string `json:"playerPositionAcceptanceThreshold" properties:"player-position-acceptance-threshold"`
	// The amount that the player's attack direction and look direction can differ.
	// Allowed values: Any value in the range of [0, 1] where 1 means that the
	// direction of the players view and the direction the player is attacking
	// must match exactly and a value of 0 means that the two directions can
	// differ by up to and including 90 degrees.
	PlayerMovementActionDirectionThreshold string `properties:"player-movement-action-direction-threshold"`
	ServerAuthoritativeBlockBreaking       string `properties:"server-authoritative-block-breaking"`
	// If true, the server will compute block mining operations in sync with the client so it can verify that the client should be able to break blocks when it thinks it can.
	ServerAuthoritativeBlockBreakingPickRangeScalar string `properties:"server-authoritative-block-breaking-pick-range-scalar"`
	// Allowed values: "None", "Dropped", "Disabled"
	// This represents the level of restriction applied to the chat for each player that joins the server.
	//
	// "None" is the default and represents regular free chat.
	//
	// "Dropped" means the chat messages are dropped and never sent to any client. Players receive a message to let them know the feature is disabled.
	//
	// "Disabled" means that unless the player is an operator, the chat UI does not even appear. No information is displayed to the player.
	ChatRestriction string `json:"chatRestriction" properties:"chat-restriction"`
	// If true, the server will inform clients that they should ignore other players when interacting with the world. This is not server authoritative.
	DisablePlayerInteraction bool `json:"disablePlayerInteraction" properties:"disable-player-interaction"`
	// If true, the server will inform clients that they have the ability to generate visual level chunks outside of player interaction distances.
	ClientSideChunkGenerationEnabled bool `json:"clientSideChunkGenerationEnabled" properties:"client-side-chunk-generation-enabled"`
	// If true, the server will send hashed block network ID's instead of id's that start from 0 and go up.  These id's are stable and won't change regardless of other block changes.
	BlockNetworkIdsAreHashes string `json:"blockNetworkIdsAreHashes" properties:"block-network-ids-are-hashes"`
	// Internal Use Only
	DisablePersona bool `json:"disablePersona" properties:"disable-persona"`
	// If true, disable players customized skins that were customized outside of the Minecraft store assets or in game assets.  This is used to disable possibly offensive custom skins players make.
	DisableCustomSkins bool `json:"disableCustomSkins" properties:"disable-custom-skins"`
	// Allowed values: "Disabled" or any value in range [0.0, 1.0]
	// If "Disabled" the server will dynamically calculate how much of the player's view it will generate, assigning the rest to the client to build.
	// Otherwise from the overridden ratio tell the server how much of the player's view to generate, disregarding client hardware capability.
	// Only valid if client-side-chunk-generation-enabled is enabled
	ServerBuildRadiusRatio string `json:"serverBuildRadiusRatio" properties:"server-build-radius-ratio"`
	// Microsoft/Mojang telemetry
	Telemetry bool `json:"telemetry" properties:"emit-server-telemetry"`
}
