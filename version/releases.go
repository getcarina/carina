package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Release is the minimal set of release data carina needs from the GitHub API
type Release struct {
	TagName string `json:"tag_name"`
}

func githubGet(uri string, rel *Release) error {
	resp, err := http.Get("https://api.github.com" + uri)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("could not fetch releases, %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github responded with a non 200 OK: %v", resp.Status)
	}

	var r io.Reader = resp.Body

	if err = json.NewDecoder(r).Decode(rel); err != nil {
		return fmt.Errorf("could not unmarshal JSON into Release struct, %v", err)
	}

	return nil
}

const latestReleaseURI = "/repos/%s/%s/releases/latest%s"

func latestReleaseAPI(user, repo, token string) (*Release, error) {
	if token != "" {
		token = "?access_token=" + token
	}
	var release Release
	return &release, githubGet(fmt.Sprintf(latestReleaseURI, user, repo, token), &release)
}

const ownerName = "getcarina"
const repoName = "carina"

// LatestRelease returns the most recent release of carina
func LatestRelease() (*Release, error) {
	return latestRelease(ownerName, repoName)
}

// Get the latest release currently
func latestRelease(owner, repo string) (*Release, error) {
	rel, err := latestReleaseAPI(owner, repo, "")

	if err != nil {
		return nil, err
	}

	return rel, nil
}
