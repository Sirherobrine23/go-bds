package microsoft

const (
	GithubVersions string = "https://github.com/actions/setup-java/raw/main/src/distributions/microsoft/microsoft-openjdk-versions.json"
)

type msVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []struct {
		File     string `json:"filename"`
		FileUrl  string `json:"download_url"`
		NodeArch string `json:"arch"`
		NodeOs   string `json:"platform"`
	} `json:"files"`
}
