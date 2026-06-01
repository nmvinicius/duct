package platform

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectGitHub(t *testing.T) {
	os.Setenv("GITHUB_ACTIONS", "true")
	defer os.Unsetenv("GITHUB_ACTIONS")

	p := Detect()
	assert.Equal(t, GitHub, p)
	assert.Equal(t, "github", p.String())
}

func TestDetectBitbucket(t *testing.T) {
	os.Setenv("BITBUCKET_PIPELINE_UUID", "uuid-123")
	defer os.Unsetenv("BITBUCKET_PIPELINE_UUID")

	p := Detect()
	assert.Equal(t, Bitbucket, p)
	assert.Equal(t, "bitbucket", p.String())
}

func TestDetectGitLab(t *testing.T) {
	os.Setenv("GITLAB_CI", "true")
	defer os.Unsetenv("GITLAB_CI")

	p := Detect()
	assert.Equal(t, GitLab, p)
	assert.Equal(t, "gitlab", p.String())
}

func TestDetectLocal(t *testing.T) {
	// Ensure no CI env vars are set
	os.Unsetenv("GITHUB_ACTIONS")
	os.Unsetenv("BITBUCKET_PIPELINE_UUID")
	os.Unsetenv("GITLAB_CI")

	p := Detect()
	assert.Equal(t, Local, p)
	assert.Equal(t, "local", p.String())
}

func TestGitVarsGitHub(t *testing.T) {
	os.Setenv("GITHUB_SHA", "abc123")
	os.Setenv("GITHUB_REF_NAME", "main")
	os.Setenv("GITHUB_REF_TYPE", "branch")
	defer func() {
		os.Unsetenv("GITHUB_SHA")
		os.Unsetenv("GITHUB_REF_NAME")
		os.Unsetenv("GITHUB_REF_TYPE")
	}()

	vars := GitHub.GitVars()
	assert.Equal(t, "abc123", vars["GIT_COMMIT"])
	assert.Equal(t, "main", vars["GIT_BRANCH"])
	assert.Empty(t, vars["GIT_TAG"])
}

func TestGitVarsGitHubTag(t *testing.T) {
	os.Setenv("GITHUB_SHA", "abc123")
	os.Setenv("GITHUB_REF_NAME", "v1.0.0")
	os.Setenv("GITHUB_REF_TYPE", "tag")
	defer func() {
		os.Unsetenv("GITHUB_SHA")
		os.Unsetenv("GITHUB_REF_NAME")
		os.Unsetenv("GITHUB_REF_TYPE")
	}()

	vars := GitHub.GitVars()
	assert.Equal(t, "v1.0.0", vars["GIT_TAG"])
}

func TestGitVarsBitbucket(t *testing.T) {
	os.Setenv("BITBUCKET_COMMIT", "def456")
	os.Setenv("BITBUCKET_BRANCH", "develop")
	os.Setenv("BITBUCKET_TAG", "v2.0.0")
	defer func() {
		os.Unsetenv("BITBUCKET_COMMIT")
		os.Unsetenv("BITBUCKET_BRANCH")
		os.Unsetenv("BITBUCKET_TAG")
	}()

	vars := Bitbucket.GitVars()
	assert.Equal(t, "def456", vars["GIT_COMMIT"])
	assert.Equal(t, "develop", vars["GIT_BRANCH"])
	assert.Equal(t, "v2.0.0", vars["GIT_TAG"])
}

func TestGitVarsGitLab(t *testing.T) {
	os.Setenv("CI_COMMIT_SHA", "ghi789")
	os.Setenv("CI_COMMIT_REF_NAME", "feature/x")
	os.Setenv("CI_COMMIT_TAG", "")
	defer func() {
		os.Unsetenv("CI_COMMIT_SHA")
		os.Unsetenv("CI_COMMIT_REF_NAME")
		os.Unsetenv("CI_COMMIT_TAG")
	}()

	vars := GitLab.GitVars()
	assert.Equal(t, "ghi789", vars["GIT_COMMIT"])
	assert.Equal(t, "feature/x", vars["GIT_BRANCH"])
	assert.Empty(t, vars["GIT_TAG"])
}

func TestGitVarsLocal(t *testing.T) {
	vars := Local.GitVars()
	// Local returns empty, populated by runner script
	assert.Empty(t, vars["GIT_COMMIT"])
}

func TestPlatformStringUnknown(t *testing.T) {
	var p Platform = 99
	assert.Equal(t, "unknown", p.String())
}

func TestDetectPriority(t *testing.T) {
	// GitHub should win if multiple are set
	os.Setenv("GITHUB_ACTIONS", "true")
	os.Setenv("GITLAB_CI", "true")
	defer func() {
		os.Unsetenv("GITHUB_ACTIONS")
		os.Unsetenv("GITLAB_CI")
	}()

	p := Detect()
	assert.Equal(t, GitHub, p)
}