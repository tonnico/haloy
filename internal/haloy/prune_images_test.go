package haloy

import (
	"errors"
	"strings"
	"testing"

	"github.com/haloydev/haloy/internal/config"
	"github.com/haloydev/haloy/internal/constants"
)

func TestDefaultPruneKeep(t *testing.T) {
	localKeep := 3
	registryKeep := 2

	tests := []struct {
		name   string
		target config.TargetConfig
		want   int
	}{
		{
			name:   "no image falls back to default",
			target: config.TargetConfig{},
			want:   int(constants.DefaultDeploymentsToKeep),
		},
		{
			name: "history none defaults to zero",
			target: config.TargetConfig{
				Image: &config.Image{
					History: &config.ImageHistory{Strategy: config.HistoryStrategyNone},
				},
			},
			want: 0,
		},
		{
			name: "local history uses configured count",
			target: config.TargetConfig{
				Image: &config.Image{
					History: &config.ImageHistory{Strategy: config.HistoryStrategyLocal, Count: &localKeep},
				},
			},
			want: 3,
		},
		{
			name: "registry history uses configured count",
			target: config.TargetConfig{
				Image: &config.Image{
					History: &config.ImageHistory{Strategy: config.HistoryStrategyRegistry, Count: &registryKeep},
				},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultPruneKeep(tt.target); got != tt.want {
				t.Fatalf("defaultPruneKeep() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestWithImagePruneHint_AddsTargetAwareHintsForDiskErrors(t *testing.T) {
	keep := 3
	target := config.TargetConfig{
		Name:       "app",
		TargetName: "staging",
		Server:     "haloy.example.com",
		Image: &config.Image{
			History: &config.ImageHistory{
				Strategy: config.HistoryStrategyLocal,
				Count:    &keep,
			},
		},
	}

	err := withImagePruneHint(errors.New("server disk space too low on /var/lib/haloy/tmp"), target)
	msg := err.Error()

	if !strings.Contains(msg, "haloy prune-images --keep 3 --yes") {
		t.Fatalf("msg = %q, want prune command hint", msg)
	}
	if !strings.Contains(msg, "image.history.count is 3") {
		t.Fatalf("msg = %q, want count hint", msg)
	}
	if !strings.Contains(msg, "haloy prune-images --keep 2 --yes") {
		t.Fatalf("msg = %q, want reduced keep hint", msg)
	}
}

func TestWithImagePruneHint_LeavesNonDiskErrorsUnchanged(t *testing.T) {
	target := config.TargetConfig{Name: "app", TargetName: "staging", Server: "haloy.example.com"}
	err := withImagePruneHint(errors.New("authentication failed"), target)
	if err.Error() != "authentication failed" {
		t.Fatalf("err = %q, want unchanged error", err.Error())
	}
}
