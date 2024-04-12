package powernukkitx

import (
	"encoding/json"
	"fmt"
	"time"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
)

const (
	ProjectVersions string = "https://search.maven.org/solrsearch/select?q=g:cn.powernukkitx+AND+a:powernukkitx&core=gav&wt=json&start=%d"
)

type mavenDocs struct {
	ID        string   `json:"id"`
	G         string   `json:"g"`
	A         string   `json:"a"`
	V         string   `json:"v"`
	P         string   `json:"p"`
	Timestamp int64    `json:"timestamp"`
	Ec        []string `json:"ec"`
	Tags      []string `json:"tags"`
}

type mevenVersions struct {
	ResponseHeader struct {
		Status int `json:"status"`
		QTime  int `json:"QTime"`
		Params struct {
			Q       string `json:"q"`
			Core    string `json:"core"`
			Indent  string `json:"indent"`
			Fl      string `json:"fl"`
			Start   string `json:"start"`
			Sort    string `json:"sort"`
			Rows    string `json:"rows"`
			Wt      string `json:"wt"`
			Version string `json:"version"`
		} `json:"params"`
	} `json:"responseHeader"`
	Response struct {
		NumFound int         `json:"numFound"`
		Start    int         `json:"start"`
		Docs     []mavenDocs `json:"docs"`
	} `json:"response"`
}

type Version struct {
	Version string    `json:"version"`
	Release time.Time `json:"releaseTime"`
	FileUrl string    `json:"url"`
}

func MapFiles() ([]Version, error) {
	docs := []mavenDocs{}
	releasesVersions := []Version{}
	NumFound := -1

	for {
		if NumFound != -1 && NumFound >= len(docs) {
			break
		}

		res, err := request.Request(request.RequestOptions{HttpError: true, Url: fmt.Sprintf(ProjectVersions, len(docs))})
		if err != nil {
			return releasesVersions, err
		}

		defer res.Body.Close()
		var decs mevenVersions
		if err = json.NewDecoder(res.Body).Decode(&decs); err != nil {
			return releasesVersions, err
		}

		NumFound = decs.Response.NumFound
		docs = append(docs, decs.Response.Docs...)
	}

	for _, doc := range docs {
		if artefactExtension, ext := internal.ArrayStringIncludes(doc.Ec, "-shaded.jar"); ext {
			releasesVersions = append(releasesVersions, Version{
				doc.V,
				time.UnixMilli(doc.Timestamp),
				fmt.Sprintf("https://search.maven.org/remotecontent?filepath=cn/powernukkitx/powernukkitx/%s/powernukkitx-%s%s", doc.V, doc.V, artefactExtension),
			})
		}
	}

	return releasesVersions, nil
}
