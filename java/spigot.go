package java

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gookit/properties"
	"sirherobrine23.org/minecraft-server/go-bds/internal/request"
)

type SpigotVersion struct {
	Version     string `json:"version"`
	ServerUrl   string `json:"spigotUrl"`
	Craftbukkit string `json:"craftbukkitUrl"`
}

func (Version *SpigotVersion) Download(serverPath string) error {
	err := os.MkdirAll(serverPath, os.FileMode(0o666))
	if !(err == os.ErrExist || err == nil) {
		return err
	}

	// Create request file
	serverResponse, err := request.Request(request.RequestOptions{Method: "GET", HttpError: true, Url: Version.ServerUrl})
	if err != nil {
		return err
	}
	defer serverResponse.Body.Close()

	// Create local file
	serverFile, err := os.Create(filepath.Join(serverPath, DefaultServerJarName))
	if err != nil {
		return err
	}

	// Save request to file
	defer serverFile.Close()
	if _, err = io.Copy(serverFile, serverResponse.Body); err != nil {
		return err
	}

	// Craftbukkit
	if len(Version.Craftbukkit) > 0 {
		parsed, _ := url.Parse(Version.Craftbukkit)
		craftbukkitResponse, err := request.Request(request.RequestOptions{Method: "GET", HttpError: true, Url: Version.Craftbukkit})
		if err != nil {
			return err
		}
		defer craftbukkitResponse.Body.Close()

		craftFile, err := os.Create(filepath.Join(serverPath, filepath.Base(parsed.Path)))
		if err != nil {
			return err
		}

		defer craftFile.Close()
		if _, err = io.Copy(craftFile, craftbukkitResponse.Body); err != nil {
			return err
		}
	}

	return nil
}

func SpigotListVersions() ([]SpigotVersion, error) {
	versions := []SpigotVersion{}
	page := 0
	for {
		var data []struct {
			TagName string `json:"tag_name"`
			Files   []struct {
				Name    string `json:"name"`
				FileUrl string `json:"browser_download_url"`
			} `json:"assets"`
		}

		err := request.GetJson(fmt.Sprintf("https://sirherobrine23.org/api/v1/repos/Minecraft-Server/Spigot/releases?page=%d", page), &data)
		if err != nil {
			return nil, err
		} else if len(data) == 0 {
			break
		}
		page++
		for _, v := range data {
			if len(v.Files) >= 2 {
				file1 := v.Files[0]
				file2 := v.Files[1]
				if len(file2.Name) > 0 {
					if file1.Name == "server.jar" {
						versions = append(versions, SpigotVersion{
							Version:     v.TagName,
							ServerUrl:   file1.FileUrl,
							Craftbukkit: file2.FileUrl,
						})
					} else {
						versions = append(versions, SpigotVersion{
							Version:     v.TagName,
							ServerUrl:   file2.FileUrl,
							Craftbukkit: file1.FileUrl,
						})
					}
				}
			} else {
				versions = append(versions, SpigotVersion{
					Version:   v.TagName,
					ServerUrl: v.Files[0].FileUrl,
				})
			}
		}
	}

	// return versions, nil
	return versions, nil
}

type SpigotProprieties struct {
	AllowFlight                    bool   `properties:"allow-flight" json:"allow-flight"`
	AllowNether                    bool   `properties:"allow-nether" json:"allow-nether"`
	BroadcastConsoleToOps          bool   `properties:"broadcast-console-to-ops" json:"broadcast-console-to-ops"`
	BroadcastRconToOps             bool   `properties:"broadcast-rcon-to-ops" json:"broadcast-rcon-to-ops"`
	Difficulty                     string `properties:"difficulty" json:"difficulty"`
	EnableCommandBlock             bool   `properties:"enable-command-block" json:"enable-command-block"`
	EnableJmxMonitoring            bool   `properties:"enable-jmx-monitoring" json:"enable-jmx-monitoring"`
	EnableQuery                    bool   `properties:"enable-query" json:"enable-query"`
	EnableRcon                     bool   `properties:"enable-rcon" json:"enable-rcon"`
	EnableStatus                   bool   `properties:"enable-status" json:"enable-status"`
	EnforceSecureProfile           bool   `properties:"enforce-secure-profile" json:"enforce-secure-profile"`
	EnforceWhitelist               bool   `properties:"enforce-whitelist" json:"enforce-whitelist"`
	EntityBroadcastRangePercentage int64  `properties:"entity-broadcast-range-percentage" json:"entity-broadcast-range-percentage"`
	ForceGamemode                  bool   `properties:"force-gamemode" json:"force-gamemode"`
	FunctionPermissionLevel        int    `properties:"function-permission-level" json:"function-permission-level"`
	Gamemode                       string `properties:"gamemode" json:"gamemode"`
	GenerateStructures             bool   `properties:"generate-structures" json:"generate-structures"`
	GeneratorSettings              string `properties:"generator-settings" json:"generator-settings"`
	Hardcore                       bool   `properties:"hardcore" json:"hardcore"`
	HideOnlinePlayers              bool   `properties:"hide-online-players" json:"hide-online-players"`
	InitialDisabledPacks           bool   `properties:"initial-disabled-packs" json:"initial-disabled-packs"`
	InitialEnabledPacks            bool   `properties:"initial-enabled-packs" json:"initial-enabled-packs"`
	LevelName                      string `properties:"level-name" json:"level-name"`
	LevelSeed                      uint64 `properties:"level-seed" json:"level-seed"`
	LevelType                      string `properties:"level-type" json:"level-type"`
	LogIps                         bool   `properties:"log-ips" json:"log-ips"`
	MaxChainedNeighborUpdates      int64  `properties:"max-chained-neighbor-updates" json:"max-chained-neighbor-updates"`
	MaxPlayers                     int64  `properties:"max-players" json:"max-players"`
	MaxTickTime                    int64  `properties:"max-tick-time" json:"max-tick-time"`
	MaxWorldSize                   int64  `properties:"max-world-size" json:"max-world-size"`
	Motd                           string `properties:"motd" json:"motd"`
	NetworkCompressionThreshold    int32  `properties:"network-compression-threshold" json:"network-compression-threshold"`
	OnlineMode                     bool   `properties:"online-mode" json:"online-mode"`
	OpPermissionLevel              int    `properties:"op-permission-level" json:"op-permission-level"`
	PlayerIdleTimeout              int64  `properties:"player-idle-timeout" json:"player-idle-timeout"`
	PreventProxyConnections        bool   `properties:"prevent-proxy-connections" json:"prevent-proxy-connections"`
	Pvp                            bool   `properties:"pvp" json:"pvp"`
	QueryPort                      int    `properties:"query.port" json:"query.port"`
	RateLimit                      int32  `properties:"rate-limit" json:"rate-limit"`
	RconPassword                   string `properties:"rcon.password" json:"rcon.password"`
	RconPort                       string `properties:"rcon.port" json:"rcon.port"`
	RequireResourcePack            bool   `properties:"require-resource-pack" json:"require-resource-pack"`
	ResourcePack                   string `properties:"resource-pack" json:"resource-pack"`
	ResourcePackId                 string `properties:"resource-pack-id" json:"resource-pack-id"`
	ResourcePackPrompt             string `properties:"resource-pack-prompt" json:"resource-pack-prompt"`
	ResourcePackSha1               string `properties:"resource-pack-sha1" json:"resource-pack-sha1"`
	ServerIp                       string `properties:"server-ip" json:"server-ip"`
	ServerPort                     int    `properties:"server-port" json:"server-port"`
	SimulationDistance             int    `properties:"simulation-distance" json:"simulation-distance"`
	SpawnAnimals                   bool   `properties:"spawn-animals" json:"spawn-animals"`
	SpawnMonsters                  bool   `properties:"spawn-monsters" json:"spawn-monsters"`
	SpawnNpcs                      bool   `properties:"spawn-npcs" json:"spawn-npcs"`
	SpawnProtection                int32  `properties:"spawn-protection" json:"spawn-protection"`
	SyncChunkWrites                bool   `properties:"sync-chunk-writes" json:"sync-chunk-writes"`
	TextFilteringConfig            string `properties:"text-filtering-config" json:"text-filtering-config"`
	UseNativeTransport             bool   `properties:"use-native-transport" json:"use-native-transport"`
	ViewDistance                   int    `properties:"view-distance" json:"view-distance"`
	WhiteList                      bool   `properties:"white-list" json:"white-list"`
}

func (config *SpigotProprieties) Load(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return properties.Unmarshal(data, config)
}

func (config *SpigotProprieties) Save(filePath string) error {
	data, err := properties.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, os.FileMode(0o666))
}
