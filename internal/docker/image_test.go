package docker

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/haloydev/haloy/internal/config"
)

func TestSelectImageTagsToRemove_IgnoreDeploymentCountsTowardKeepLimit(t *testing.T) {
	candidates := []removableImageTag{
		{Tag: "app:20260222010101", DeploymentID: "20260222010101", ImageID: "img-1"},
		{Tag: "app:20260222010102", DeploymentID: "20260222010102", ImageID: "img-2"},
		{Tag: "app:20260222010103", DeploymentID: "20260222010103", ImageID: "img-3"},
	}

	removals := selectImageTagsToRemove(candidates, map[string]struct{}{}, 2, "20260222010104")

	if len(removals) != 2 {
		t.Fatalf("len(removals) = %d, want 2", len(removals))
	}
	if removals[0].Tag != "app:20260222010102" {
		t.Fatalf("removals[0].Tag = %q, want %q", removals[0].Tag, "app:20260222010102")
	}
	if removals[1].Tag != "app:20260222010101" {
		t.Fatalf("removals[1].Tag = %q, want %q", removals[1].Tag, "app:20260222010101")
	}
}

func TestSelectImageTagsToRemove_KeepCurrentOnlyRemovesOlderTags(t *testing.T) {
	candidates := []removableImageTag{
		{Tag: "app:20260222010101", DeploymentID: "20260222010101", ImageID: "img-1"},
		{Tag: "app:20260222010102", DeploymentID: "20260222010102", ImageID: "img-2"},
	}

	removals := selectImageTagsToRemove(candidates, map[string]struct{}{}, 1, "20260222010103")

	if len(removals) != 2 {
		t.Fatalf("len(removals) = %d, want 2", len(removals))
	}
	if removals[0].Tag != "app:20260222010102" {
		t.Fatalf("removals[0].Tag = %q, want %q", removals[0].Tag, "app:20260222010102")
	}
	if removals[1].Tag != "app:20260222010101" {
		t.Fatalf("removals[1].Tag = %q, want %q", removals[1].Tag, "app:20260222010101")
	}
}

func TestSelectImageTagsToRemove_PreservesInUseImageWithoutKeptTag(t *testing.T) {
	candidates := []removableImageTag{
		{Tag: "app:20260222010101", DeploymentID: "20260222010101", ImageID: "img-1"},
	}

	removals := selectImageTagsToRemove(candidates, map[string]struct{}{"img-1": {}}, 0, "20260222010102")

	if len(removals) != 0 {
		t.Fatalf("len(removals) = %d, want 0", len(removals))
	}
}

func TestPlanImagePrune_ReturnsRunningDeploymentsAndRemovableTags(t *testing.T) {
	candidates := []removableImageTag{
		{Tag: "app:20260222010101", DeploymentID: "20260222010101", ImageID: "img-1"},
		{Tag: "app:20260222010102", DeploymentID: "20260222010102", ImageID: "img-2"},
		{Tag: "app:20260222010103", DeploymentID: "20260222010103", ImageID: "img-3"},
	}

	plan := planImagePrune(
		candidates,
		map[string]struct{}{"img-3": {}},
		"app",
		"",
		1,
		[]string{"20260222010103"},
	)

	if plan.AppName != "app" {
		t.Fatalf("plan.AppName = %q, want %q", plan.AppName, "app")
	}
	if plan.Keep != 1 {
		t.Fatalf("plan.Keep = %d, want %d", plan.Keep, 1)
	}
	if len(plan.RunningDeploymentIDs) != 1 || plan.RunningDeploymentIDs[0] != "20260222010103" {
		t.Fatalf("plan.RunningDeploymentIDs = %v, want [20260222010103]", plan.RunningDeploymentIDs)
	}
	if len(plan.Tags) != 2 {
		t.Fatalf("len(plan.Tags) = %d, want 2", len(plan.Tags))
	}
	if plan.Tags[0].Tag != "app:20260222010102" {
		t.Fatalf("plan.Tags[0].Tag = %q, want %q", plan.Tags[0].Tag, "app:20260222010102")
	}
	if plan.Tags[1].Tag != "app:20260222010101" {
		t.Fatalf("plan.Tags[1].Tag = %q, want %q", plan.Tags[1].Tag, "app:20260222010101")
	}
}

func TestRunningDeploymentIDs_DeduplicatesAndSortsDescending(t *testing.T) {
	containers := []container.Summary{
		{Labels: map[string]string{config.LabelDeploymentID: "20260222010102"}},
		{Labels: map[string]string{config.LabelDeploymentID: "20260222010101"}},
		{Labels: map[string]string{config.LabelDeploymentID: "20260222010102"}},
	}

	ids := runningDeploymentIDs(containers)

	if len(ids) != 2 {
		t.Fatalf("len(ids) = %d, want 2", len(ids))
	}
	if ids[0] != "20260222010102" || ids[1] != "20260222010101" {
		t.Fatalf("ids = %v, want [20260222010102 20260222010101]", ids)
	}
}
