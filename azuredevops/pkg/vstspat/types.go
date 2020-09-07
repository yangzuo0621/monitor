package vstspat

import "context"

// PATProvider provides VSTS personal access token.
type PATProvider interface {
	// GetPAT gets a personal access token.
	GetPAT(ctx context.Context) (string, error)
}
