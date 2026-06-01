package platform

import "os"

// Platform represents the CI/CD platform
type Platform int

const (
	Unknown Platform = iota
	GitHub
	Bitbucket
	GitLab
	Local
)

func (p Platform) String() string {
	switch p {
	case GitHub:
		return "github"
	case Bitbucket:
		return "bitbucket"
	case GitLab:
		return "gitlab"
	case Local:
		return "local"
	default:
		return "unknown"
	}
}

// Detect identifies the current CI platform from environment variables
func Detect() Platform {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		return GitHub
	}
	if os.Getenv("BITBUCKET_PIPELINE_UUID") != "" {
		return Bitbucket
	}
	if os.Getenv("GITLAB_CI") != "" {
		return GitLab
	}
	return Local
}

// GitVars returns git-related variables for the current platform
func (p Platform) GitVars() map[string]string {
	vars := make(map[string]string)

	switch p {
	case GitHub:
		vars["GIT_COMMIT"] = os.Getenv("GITHUB_SHA")
		vars["GIT_BRANCH"] = os.Getenv("GITHUB_REF_NAME")
		if os.Getenv("GITHUB_REF_TYPE") == "tag" {
			vars["GIT_TAG"] = os.Getenv("GITHUB_REF_NAME")
		}
	case Bitbucket:
		vars["GIT_COMMIT"] = os.Getenv("BITBUCKET_COMMIT")
		vars["GIT_BRANCH"] = os.Getenv("BITBUCKET_BRANCH")
		vars["GIT_TAG"] = os.Getenv("BITBUCKET_TAG")
	case GitLab:
		vars["GIT_COMMIT"] = os.Getenv("CI_COMMIT_SHA")
		vars["GIT_BRANCH"] = os.Getenv("CI_COMMIT_REF_NAME")
		vars["GIT_TAG"] = os.Getenv("CI_COMMIT_TAG")
	case Local:
		// Will be populated by shelling out to git
	}

	return vars
}
