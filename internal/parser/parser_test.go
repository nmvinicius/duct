package parser

import (
	"testing"

	"github.com/nmvinicius/duct/pkg/ductfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMinimal(t *testing.T) {
	input := `VERSION 1.0

PROJECT test-project

STEP build
    RUN echo hello
`

	df, err := Parse(input)
	require.NoError(t, err)

	assert.Equal(t, "1.0", df.Version)
	assert.Equal(t, "test-project", df.Project)
	assert.Len(t, df.Steps, 1)
	assert.Equal(t, "build", df.Steps[0].Name)
	assert.Equal(t, []string{"echo hello"}, df.Steps[0].Runs)
}

func TestParseFullDuctfile(t *testing.T) {
	input := `VERSION 1.0

PROJECT my-api
TEAM backend

GLOBAL NODE_VERSION=20
GLOBAL REGISTRY=registry.io

EXTENDS github.com/org/templates//node/base

CACHE node_modules
CACHE docker_layers:/var/lib/docker

REQUIRE SECRET AWS_KEY
REQUIRE SECRET AWS_SECRET

STEP lint
    USE node
    RUN npm ci
    RUN npm run lint
    CACHE node_modules

STEP test
    USE node
    NEEDS lint
    RUN npm test
    ARTIFACTS coverage/

STEP build
    USE docker
    NEEDS test
    RUN docker build -t app .
    WHEN branch == "main"
    ROLLBACK docker rmi app
    ALLOW_FAIL

STEP notify
    NEEDS build
    NOTIFY slack "#deploys" "Build done"
`

	df, err := Parse(input)
	require.NoError(t, err)

	// Metadata
	assert.Equal(t, "1.0", df.Version)
	assert.Equal(t, "my-api", df.Project)
	assert.Equal(t, "backend", df.Team)

	// Globals
	assert.Equal(t, "20", df.Globals["NODE_VERSION"])
	assert.Equal(t, "registry.io", df.Globals["REGISTRY"])

	// Extends
	assert.Equal(t, []string{"github.com/org/templates//node/base"}, df.Extends)

	// Caches
	require.Len(t, df.Caches, 2)
	assert.Equal(t, "node_modules", df.Caches[0].Name)
	assert.Equal(t, "", df.Caches[0].Path)
	assert.Equal(t, "docker_layers", df.Caches[1].Name)
	assert.Equal(t, "/var/lib/docker", df.Caches[1].Path)

	// Secrets
	assert.Equal(t, []string{"AWS_KEY", "AWS_SECRET"}, df.Secrets)

	// Steps
	require.Len(t, df.Steps, 4)

	// Step: lint
	lint := df.Steps[0]
	assert.Equal(t, "lint", lint.Name)
	assert.Equal(t, []string{"node"}, lint.Uses)
	assert.Equal(t, []string{"npm ci", "npm run lint"}, lint.Runs)
	assert.Equal(t, []string{"node_modules"}, lint.Caches)

	// Step: test
	test := df.Steps[1]
	assert.Equal(t, "test", test.Name)
	assert.Equal(t, []string{"lint"}, test.Needs)
	assert.Equal(t, []string{"coverage/"}, test.Artifacts)

	// Step: build
	build := df.Steps[2]
	assert.Equal(t, "build", build.Name)
	assert.Equal(t, []string{"docker"}, build.Uses)
	assert.Equal(t, []string{"test"}, build.Needs)
	assert.Equal(t, "docker build -t app .", build.Runs[0])
	assert.True(t, build.AllowFail)
	assert.Equal(t, "docker rmi app", build.Rollback)
	require.NotNil(t, build.When)
	assert.Equal(t, `== "main"`, build.When.Branch)

	// Step: notify
	notify := df.Steps[3]
	assert.Equal(t, "notify", notify.Name)
	require.NotNil(t, notify.Notify)
	assert.Equal(t, "#deploys", notify.Notify.Channel)
	assert.Equal(t, "Build done", notify.Notify.Message)
}

func TestParseCommentsAndEmptyLines(t *testing.T) {
	input := `# This is a comment

VERSION 1.0
# Another comment

PROJECT test

# Before step
STEP build
    # Inside step
    RUN echo hello
`

	df, err := Parse(input)
	require.NoError(t, err)
	assert.Equal(t, "test", df.Project)
	assert.Len(t, df.Steps, 1)
}

func TestParseMultipleNeeds(t *testing.T) {
	input := `VERSION 1.0

PROJECT test

STEP a
    RUN echo a

STEP b
    RUN echo b

STEP c
    NEEDS a, b
    RUN echo c
`

	df, err := Parse(input)
	require.NoError(t, err)

	c := df.Steps[2]
	assert.Equal(t, "c", c.Name)
	assert.Equal(t, []string{"a", "b"}, c.Needs)
}

func TestParseErrorUnknownCommand(t *testing.T) {
	input := `VERSION 1.0

PROJECT test

UNKNOWN_COMMAND something
`

	_, err := Parse(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown top-level command")
}

func TestParseErrorIndentedOutsideStep(t *testing.T) {
	input := `VERSION 1.0

PROJECT test

    RUN echo hello
`

	_, err := Parse(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside of STEP block")
}

func TestParseConditionComplex(t *testing.T) {
	input := `VERSION 1.0

PROJECT test

STEP deploy
    WHEN branch == "main" AND tag =~ "^v[0-9]"
    RUN echo deploy
`

	df, err := Parse(input)
	require.NoError(t, err)

	step := df.Steps[0]
	require.NotNil(t, step.When)
	assert.Equal(t, `== "main"`, step.When.Branch)
	assert.Equal(t, `=~ "^v[0-9]"`, step.When.Tag)
}

func TestParseFileNotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/Ductfile")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read Ductfile")
}

func TestValidateSteps(t *testing.T) {
	steps := []ductfile.Step{
		{Name: "lint", Needs: []string{}},
		{Name: "build", Needs: []string{"lint"}},
		{Name: "deploy", Needs: []string{"build"}},
	}

	sorted, err := ValidateSteps(steps)
	require.NoError(t, err)
	assert.Equal(t, "lint", sorted[0].Name)
	assert.Equal(t, "build", sorted[1].Name)
	assert.Equal(t, "deploy", sorted[2].Name)
}

func TestGetExecutionOrder(t *testing.T) {
	steps := []ductfile.Step{
		{Name: "b", Needs: []string{"a"}},
		{Name: "a", Needs: []string{}},
	}

	sorted, err := GetExecutionOrder(steps)
	require.NoError(t, err)
	assert.Equal(t, "a", sorted[0].Name)
	assert.Equal(t, "b", sorted[1].Name)
}

func TestParseGlobalWithQuotes(t *testing.T) {
	input := `VERSION 1.0

PROJECT test

GLOBAL MESSAGE="hello world"
GLOBAL UNQUOTED=value

STEP build
    RUN echo $MESSAGE
`

	df, err := Parse(input)
	require.NoError(t, err)
	assert.Equal(t, "hello world", df.Globals["MESSAGE"])
	assert.Equal(t, "value", df.Globals["UNQUOTED"])
}

func TestParseNoSteps(t *testing.T) {
	input := `VERSION 1.0

PROJECT empty
`

	df, err := Parse(input)
	require.NoError(t, err)
	assert.Empty(t, df.Steps)
}