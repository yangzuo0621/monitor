package vstspat

import (
	"context"
	"os"
)

type PATEnvBackend struct {
	envKey    string
	envSource func(s string) string
}

var _ PATProvider = (*PATEnvBackend)(nil)

// NewPATEnvBackend creates PATEnvBackend instance.
func NewPATEnvBackend(envKey string) *PATEnvBackend {
	return &PATEnvBackend{
		envKey:    envKey,
		envSource: os.Getenv,
	}
}

func (b *PATEnvBackend) getEnv(s string) string {
	if b.envSource == nil {
		return os.Getenv(s)
	}

	return b.envSource(s)
}

func (b *PATEnvBackend) GetPAT(ctx context.Context) (string, error) {
	return b.getEnv(b.envKey), nil
}
