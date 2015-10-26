package version

import "github.com/google/go-github/github"

const ownerName = "getcarina"
const repoName = "carina"

// LatestRelease returns the most recent release of carina
func LatestRelease() (*github.RepositoryRelease, error) {
	return latestRelease(ownerName, repoName)
}

// Get the latest release currently
func latestRelease(owner, repo string) (*github.RepositoryRelease, error) {
	client := github.NewClient(nil)

	rel, resp, err := client.Repositories.GetLatestRelease(owner, repo)

	if err != nil {
		return nil, err
	}

	return rel, nil
}
