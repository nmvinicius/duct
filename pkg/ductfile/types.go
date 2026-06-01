package ductfile

import (
	"fmt"
	"regexp"
	"strings"
)

// Ductfile represents the parsed pipeline definition
type Ductfile struct {
	Version  string            `json:"version"`
	Project  string            `json:"project"`
	Team     string            `json:"team,omitempty"`
	Globals  map[string]string `json:"globals"`
	Extends  []string          `json:"extends,omitempty"`
	Caches   []CacheDef        `json:"caches,omitempty"`
	Secrets  []string          `json:"secrets,omitempty"`
	Steps    []Step            `json:"steps"`
	RawLines []string          `json:"-"`
}

// Step represents a single pipeline step
type Step struct {
	Name      string        `json:"name"`
	Uses      []string      `json:"uses,omitempty"`
	Runs      []string      `json:"runs"`
	Needs     []string      `json:"needs,omitempty"`
	When      *Condition    `json:"when,omitempty"`
	Caches    []string      `json:"caches,omitempty"`
	Artifacts []string      `json:"artifacts,omitempty"`
	Rollback  string        `json:"rollback,omitempty"`
	AllowFail bool          `json:"allow_fail,omitempty"`
	Notify    *NotifyConfig `json:"notify,omitempty"`
}

// Condition represents a WHEN clause
type Condition struct {
	Raw      string `json:"raw"`
	Branch   string `json:"branch,omitempty"`
	Tag      string `json:"tag,omitempty"`
	Platform string `json:"platform,omitempty"`
}

// NotifyConfig represents notification settings
type NotifyConfig struct {
	Channel string `json:"channel"`
	Message string `json:"message"`
}

// CacheDef represents a cache definition
type CacheDef struct {
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
}

// ParseError represents a parsing error with line number
type ParseError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Raw     string `json:"raw"`
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d: %s (near: %q)", e.Line, e.Message, e.Raw)
}

// Evaluate checks if a condition is met given the current context
func (c *Condition) Evaluate(branch, tag, platform string) (bool, error) {
	if c == nil {
		return true, nil
	}

	if c.Branch != "" {
		matched, err := matchCondition(c.Branch, branch)
		if err != nil || !matched {
			return matched, err
		}
	}

	if c.Tag != "" {
		matched, err := matchCondition(c.Tag, tag)
		if err != nil || !matched {
			return matched, err
		}
	}

	if c.Platform != "" {
		matched, err := matchCondition(c.Platform, platform)
		if err != nil || !matched {
			return matched, err
		}
	}

	return true, nil
}

func matchCondition(pattern, value string) (bool, error) {
	pattern = strings.TrimSpace(pattern)
	value = strings.TrimSpace(value)

	if strings.HasPrefix(pattern, `==`) {
		expected := strings.Trim(strings.TrimPrefix(pattern, `==`), ` "`)
		return value == expected, nil
	}

	if strings.HasPrefix(pattern, `=~`) {
		regexStr := strings.Trim(strings.TrimPrefix(pattern, `=~`), ` "`)
		re, err := regexp.Compile(regexStr)
		if err != nil {
			return false, fmt.Errorf("invalid regex %q: %w", regexStr, err)
		}
		return re.MatchString(value), nil
	}

	if strings.HasPrefix(pattern, `!=`) {
		expected := strings.Trim(strings.TrimPrefix(pattern, `!=`), ` "`)
		return value != expected, nil
	}

	return value == pattern, nil
}

// StepGraph builds a dependency graph from steps
func StepGraph(steps []Step) (map[string][]string, error) {
	graph := make(map[string][]string)
	stepNames := make(map[string]bool)

	for _, s := range steps {
		stepNames[s.Name] = true
		graph[s.Name] = s.Needs
	}

	for stepName, deps := range graph {
		for _, dep := range deps {
			if !stepNames[dep] {
				return nil, fmt.Errorf("step %q depends on non-existent step %q", stepName, dep)
			}
		}
	}

	return graph, nil
}

// TopologicalSort returns steps in dependency order
func TopologicalSort(steps []Step) ([]Step, error) {
	graph, err := StepGraph(steps)
	if err != nil {
		return nil, err
	}

	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	order := make([]string, 0)

	var visit func(string) error
	visit = func(name string) error {
		if visiting[name] {
			return fmt.Errorf("circular dependency detected involving step %q", name)
		}
		if visited[name] {
			return nil
		}

		visiting[name] = true
		for _, dep := range graph[name] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[name] = false
		visited[name] = true
		order = append(order, name)
		return nil
	}

	for _, s := range steps {
		if err := visit(s.Name); err != nil {
			return nil, err
		}
	}

	stepMap := make(map[string]Step)
	for _, s := range steps {
		stepMap[s.Name] = s
	}

	result := make([]Step, 0, len(order))
	for _, name := range order {
		result = append(result, stepMap[name])
	}

	return result, nil
}
