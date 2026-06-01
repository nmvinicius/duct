package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/nmvinicius/duct/pkg/ductfile"
)

// ParseFile reads and parses a Ductfile
func ParseFile(path string) (*ductfile.Ductfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read Ductfile: %w", err)
	}

	return Parse(string(data))
}

// Parse parses Ductfile content
func Parse(content string) (*ductfile.Ductfile, error) {
	df := &ductfile.Ductfile{
		Version: "1.0",
		Globals: make(map[string]string),
		Steps:   []ductfile.Step{},
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	var currentStep *ductfile.Step
	var inStep bool

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		df.RawLines = append(df.RawLines, line)

		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Detect indentation level
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		// Top-level commands (no indent)
		if indent == 0 {
			// Close current step if open
			if inStep && currentStep != nil {
				df.Steps = append(df.Steps, *currentStep)
				currentStep = nil
				inStep = false
			}

			parts := strings.Fields(trimmed)
			if len(parts) == 0 {
				continue
			}

			cmd := strings.ToUpper(parts[0])

			switch cmd {
			case "VERSION":
				if len(parts) >= 2 {
					df.Version = strings.Trim(parts[1], `"`)
				}

			case "PROJECT":
				if len(parts) >= 2 {
					df.Project = strings.Join(parts[1:], " ")
				}

			case "TEAM":
				if len(parts) >= 2 {
					df.Team = strings.Join(parts[1:], " ")
				}

			case "GLOBAL":
				if len(parts) >= 2 {
					kv := strings.Join(parts[1:], " ")
					if idx := strings.Index(kv, "="); idx > 0 {
						key := strings.TrimSpace(kv[:idx])
						val := strings.Trim(strings.TrimSpace(kv[idx+1:]), `"`)
						df.Globals[key] = val
					}
				}

			case "EXTENDS":
				if len(parts) >= 2 {
					df.Extends = append(df.Extends, strings.Join(parts[1:], " "))
				}

			case "CACHE":
				if len(parts) >= 2 {
					cacheName := parts[1]
					cachePath := ""
					if idx := strings.Index(cacheName, ":"); idx >= 0 {
						cachePath = cacheName[idx+1:]
						cacheName = cacheName[:idx]
					} else if len(parts) >= 3 {
						cachePath = strings.Trim(parts[2], `"`)
					}
					df.Caches = append(df.Caches, ductfile.CacheDef{
						Name: cacheName,
						Path: cachePath,
					})
				}

			case "REQUIRE":
				if len(parts) >= 3 && strings.ToUpper(parts[1]) == "SECRET" {
					df.Secrets = append(df.Secrets, parts[2])
				}

			case "STEP":
				if len(parts) >= 2 {
					currentStep = &ductfile.Step{
						Name:  parts[1],
						Runs:  []string{},
						Needs: []string{},
						Uses:  []string{},
					}
					inStep = true
				}

			default:
				return nil, ductfile.ParseError{
					Line:    lineNum,
					Message: fmt.Sprintf("unknown top-level command: %s", cmd),
					Raw:     trimmed,
				}
			}

			continue
		}

		// Step-level commands (indented)
		if inStep && currentStep != nil && indent > 0 {
			parts := strings.Fields(trimmed)
			if len(parts) == 0 {
				continue
			}

			cmd := strings.ToUpper(parts[0])

			switch cmd {
			case "USE":
				if len(parts) >= 2 {
					currentStep.Uses = append(currentStep.Uses, parts[1])
				}

			case "RUN":
				currentStep.Runs = append(currentStep.Runs, strings.Join(parts[1:], " "))

			case "NEEDS":
				if len(parts) >= 2 {
					deps := strings.Split(strings.Join(parts[1:], " "), ",")
					for _, d := range deps {
						currentStep.Needs = append(currentStep.Needs, strings.TrimSpace(d))
					}
				}

			case "WHEN":
				if len(parts) >= 2 {
					cond := parseCondition(strings.Join(parts[1:], " "))
					currentStep.When = cond
				}

			case "CACHE":
				if len(parts) >= 2 {
					currentStep.Caches = append(currentStep.Caches, parts[1])
				}

			case "ARTIFACTS":
				if len(parts) >= 2 {
					currentStep.Artifacts = append(currentStep.Artifacts, parts[1])
				}

			case "ROLLBACK":
				currentStep.Rollback = strings.Join(parts[1:], " ")

			case "ALLOW_FAIL":
				currentStep.AllowFail = true

			case "NOTIFY":
				if len(parts) >= 2 {
					quoted := parseQuotedArgs(trimmed)
					channel := ""
					message := ""
					if len(quoted) >= 1 {
						channel = quoted[0]
					}
					if len(quoted) >= 2 {
						message = quoted[1]
					}
					currentStep.Notify = &ductfile.NotifyConfig{
						Channel: channel,
						Message: message,
					}
				}

			default:
				return nil, ductfile.ParseError{
					Line:    lineNum,
					Message: fmt.Sprintf("unknown step command: %s", cmd),
					Raw:     trimmed,
				}
			}

			continue
		}

		// If we get here, there's an indentation issue
		if indent > 0 && !inStep {
			return nil, ductfile.ParseError{
				Line:    lineNum,
				Message: "indented command outside of STEP block",
				Raw:     trimmed,
			}
		}
	}

	// Close last step
	if inStep && currentStep != nil {
		df.Steps = append(df.Steps, *currentStep)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	return df, nil
}

func parseQuotedArgs(input string) []string {
	re := regexp.MustCompile(`"([^"]*)"`)
	matches := re.FindAllStringSubmatch(input, -1)
	args := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			args = append(args, m[1])
		}
	}
	return args
}

func parseCondition(raw string) *ductfile.Condition {
	cond := &ductfile.Condition{Raw: raw}

	// Parse: branch == "main", tag =~ "^v[0-9]", etc.
	re := regexp.MustCompile(`(branch|tag|platform)\s*(==|=~|!=)\s*"([^"]*)"`)
	matches := re.FindAllStringSubmatch(raw, -1)

	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		field := m[1]
		operator := m[2]
		value := m[3]

		pattern := operator + ` "` + value + `"`

		switch field {
		case "branch":
			cond.Branch = pattern
		case "tag":
			cond.Tag = pattern
		case "platform":
			cond.Platform = pattern
		}
	}

	return cond
}

// ValidateSteps checks step dependencies and returns execution order
func ValidateSteps(steps []ductfile.Step) ([]ductfile.Step, error) {
	return ductfile.TopologicalSort(steps)
}

// GetExecutionOrder returns steps sorted by dependencies
func GetExecutionOrder(steps []ductfile.Step) ([]ductfile.Step, error) {
	return ductfile.TopologicalSort(steps)
}
