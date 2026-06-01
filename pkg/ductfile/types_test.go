package ductfile_test

import (
	"testing"

	"github.com/nmvinicius/duct/pkg/ductfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConditionEvaluate(t *testing.T) {
	tests := []struct {
		name     string
		cond     *ductfile.Condition
		branch   string
		tag      string
		platform string
		want     bool
		wantErr  bool
	}{
		{
			name:     "no condition always true",
			cond:     nil,
			branch:   "main",
			want:     true,
			wantErr:  false,
		},
		{
			name: "branch exact match",
			cond: &ductfile.Condition{
				Branch: `== "main"`,
			},
			branch: "main",
			want:   true,
		},
		{
			name: "branch exact mismatch",
			cond: &ductfile.Condition{
				Branch: `== "main"`,
			},
			branch: "develop",
			want:   false,
		},
		{
			name: "branch regex match",
			cond: &ductfile.Condition{
				Branch: `=~ "^feature/"`,
			},
			branch: "feature/auth",
			want:   true,
		},
		{
			name: "branch regex no match",
			cond: &ductfile.Condition{
				Branch: `=~ "^feature/"`,
			},
			branch: "bugfix/crash",
			want:   false,
		},
		{
			name: "tag regex match",
			cond: &ductfile.Condition{
				Tag: `=~ "^v[0-9]+"`,
			},
			tag:  "v1.2.3",
			want: true,
		},
		{
			name: "tag regex no match",
			cond: &ductfile.Condition{
				Tag: `=~ "^v[0-9]+"`,
			},
			tag:  "release-2024",
			want: false,
		},
		{
			name: "branch not equal",
			cond: &ductfile.Condition{
				Branch: `!= "main"`,
			},
			branch: "develop",
			want:   true,
		},
		{
			name: "combined conditions all match",
			cond: &ductfile.Condition{
				Branch:   `== "main"`,
				Tag:      `=~ "^v"`,
				Platform: `== "github"`,
			},
			branch:   "main",
			tag:      "v1.0.0",
			platform: "github",
			want:     true,
		},
		{
			name: "combined conditions one fails",
			cond: &ductfile.Condition{
				Branch: `== "main"`,
				Tag:    `=~ "^v"`,
			},
			branch: "develop",
			tag:    "v1.0.0",
			want:   false,
		},
		{
			name: "invalid regex",
			cond: &ductfile.Condition{
				Branch: `=~ "[invalid"`,
			},
			branch:  "main",
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cond.Evaluate(tt.branch, tt.tag, tt.platform)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTopologicalSort(t *testing.T) {
	tests := []struct {
		name    string
		steps   []ductfile.Step
		want    []string
		wantErr bool
	}{
		{
			name: "linear dependency",
			steps: []ductfile.Step{
				{ Name: "build", Needs: []string{"lint"} },
				{ Name: "lint", Needs: []string{} },
				{ Name: "deploy", Needs: []string{"build"} },
			},
			want: []string{"lint", "build", "deploy"},
		},
		{
			name: "multiple independent",
			steps: []ductfile.Step{
				{ Name: "test", Needs: []string{} },
				{ Name: "lint", Needs: []string{} },
				{ Name: "build", Needs: []string{"test", "lint"} },
			},
			want: []string{"test", "lint", "build"},
		},
		{
			name: "complex graph",
			steps: []ductfile.Step{
				{ Name: "a", Needs: []string{} },
				{ Name: "b", Needs: []string{"a"} },
				{ Name: "c", Needs: []string{"a"} },
				{ Name: "d", Needs: []string{"b", "c"} },
			},
			want: []string{"a", "b", "c", "d"},
		},
		{
			name: "missing dependency",
			steps: []ductfile.Step{
				{ Name: "build", Needs: []string{"lint"} },
			},
			wantErr: true,
		},
		{
			name: "circular dependency",
			steps: []ductfile.Step{
				{ Name: "a", Needs: []string{"b"} },
				{ Name: "b", Needs: []string{"a"} },
			},
			wantErr: true,
		},
		{
			name: "self dependency",
			steps: []ductfile.Step{
				{ Name: "a", Needs: []string{"a"} },
			},
			wantErr: true,
		},
		{
			name: "empty steps",
			steps: []ductfile.Step{},
			want: []string{},
		},
		{
			name: "single step no deps",
			steps: []ductfile.Step{
				{ Name: "build", Needs: []string{} },
			},
			want: []string{"build"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ductfile.TopologicalSort(tt.steps)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			gotNames := make([]string, len(got))
			for i, s := range got {
				gotNames[i] = s.Name
			}
			assert.Equal(t, tt.want, gotNames)
		})
	}
}

func TestStepGraph(t *testing.T) {
	steps := []ductfile.Step{
		{Name: "a", Needs: []string{}},
		{Name: "b", Needs: []string{"a"}},
		{Name: "c", Needs: []string{"a"}},
	}

	graph, err := ductfile.StepGraph(steps)
	require.NoError(t, err)

	assert.Equal(t, []string{}, graph["a"])
	assert.Equal(t, []string{"a"}, graph["b"])
	assert.Equal(t, []string{"a"}, graph["c"])
}

func TestCacheDef(t *testing.T) {
	c := ductfile.CacheDef{
		Name: "node_modules",
		Path: "./node_modules",
	}
	assert.Equal(t, "node_modules", c.Name)
	assert.Equal(t, "./node_modules", c.Path)
}

func TestNotifyConfig(t *testing.T) {
	n := ductfile.NotifyConfig{
		Channel: "#deploys",
		Message: "Deployed!",
	}
	assert.Equal(t, "#deploys", n.Channel)
	assert.Equal(t, "Deployed!", n.Message)
}