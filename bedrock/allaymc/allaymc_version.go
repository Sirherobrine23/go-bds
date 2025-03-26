package allaymc

type Version struct {
	Version     string `json:"version"`      // Server version
	MCVersion   string `json:"mc_version"`   // Minecraft bedrock version
	ServerURL   string `json:"download"`     // Server file to download
	JavaVersion uint   `json:"java_version"` // Java Version, example: 21
}
