package extends

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/nmvinicius/duct/pkg/ductfile"
)

// Resolver handles EXTENDS directives
type Resolver struct {
	cacheDir string
}

// NewResolver creates a new extends resolver
func NewResolver(cacheDir string) *Resolver {
	return &Resolver{
		cacheDir: cacheDir,
	}
}

// Resolve fetches and parses a parent Ductfile
func (r *Resolver) Resolve(path string) (*ductfile.Ductfile, error) {
	// Local path
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "/") {
		return r.resolveLocal(path)
	}

	// URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return r.resolveURL(path)
	}

	// GitHub shorthand: github.com/user/repo//path/to/file
	if strings.Contains(path, "//") {
		return r.resolveGitHub(path)
	}

	// Try as local first, then remote
	if _, err := os.Stat(path); err == nil {
		return r.resolveLocal(path)
	}

	return nil, fmt.Errorf("cannot resolve extends: %s", path)
}

func (r *Resolver) resolveLocal(path string) (*ductfile.Ductfile, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	// TODO: Call parser.ParseFile
	_ = abs
	return nil, fmt.Errorf("local extends not yet implemented")
}

func (r *Resolver) resolveURL(urlStr string) (*ductfile.Ductfile, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	_ = u
	return nil, fmt.Errorf("URL extends not yet implemented")
}

func (r *Resolver) resolveGitHub(path string) (*ductfile.Ductfile, error) {
	// github.com/user/repo//path/to/file.duct
	parts := strings.SplitN(path, "//", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid github extends format: %s", path)
	}

	repo := parts[0] // github.com/user/repo
	file := parts[1] // path/to/file.duct

	// Convert to raw GitHub URL
	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s",
		strings.TrimPrefix(repo, "github.com/"),
		file)

	_ = rawURL
	return nil, fmt.Errorf("github extends not yet implemented: would fetch %s", rawURL)
}
