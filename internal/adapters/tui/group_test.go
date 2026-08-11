package tui

import (
	"testing"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

func TestVisibleTreeNodes_Grouping(t *testing.T) {
	m := NewModel(nil, "docker")
	m.containers = []containerItem{
		{Container: domain.Container{ID: "1", Name: "app-web", ComposeProject: "my-stack", Status: domain.StatusRunning}},
		{Container: domain.Container{ID: "2", Name: "app-db", ComposeProject: "my-stack", Status: domain.StatusRunning}},
		{Container: domain.Container{ID: "3", Name: "standalone-nginx", Status: domain.StatusRunning}},
	}

	nodes := m.VisibleTreeNodes()
	if len(nodes) != 4 {
		t.Fatalf("got %d nodes, want 4", len(nodes))
	}

	if nodes[0].Type != NodeProjectHeader || nodes[0].ProjectName != "my-stack" {
		t.Errorf("node 0 = %+v, want ProjectHeader my-stack", nodes[0])
	}
	if nodes[0].TotalCount != 2 || nodes[0].RunningCount != 2 {
		t.Errorf("node 0 counts = (%d, %d), want (2, 2)", nodes[0].TotalCount, nodes[0].RunningCount)
	}

	if nodes[1].Type != NodeContainerItem || nodes[1].Container.Name != "app-web" || nodes[1].IsLastInGroup {
		t.Errorf("node 1 = %+v, want app-web (IsLastInGroup=false)", nodes[1])
	}

	if nodes[2].Type != NodeContainerItem || nodes[2].Container.Name != "app-db" || !nodes[2].IsLastInGroup {
		t.Errorf("node 2 = %+v, want app-db (IsLastInGroup=true)", nodes[2])
	}

	if nodes[3].Type != NodeContainerItem || nodes[3].Container.Name != "standalone-nginx" || nodes[3].ProjectName != "" {
		t.Errorf("node 3 = %+v, want standalone-nginx", nodes[3])
	}
}

func TestVisibleTreeNodes_Collapse(t *testing.T) {
	m := NewModel(nil, "docker")
	m.containers = []containerItem{
		{Container: domain.Container{ID: "1", Name: "app-web", ComposeProject: "my-stack", Status: domain.StatusRunning}},
		{Container: domain.Container{ID: "2", Name: "app-db", ComposeProject: "my-stack", Status: domain.StatusRunning}},
		{Container: domain.Container{ID: "3", Name: "standalone-nginx", Status: domain.StatusRunning}},
	}

	m.toggleProjectExpanded("my-stack")

	nodes := m.VisibleTreeNodes()
	if len(nodes) != 2 {
		t.Fatalf("got %d nodes, want 2 when collapsed", len(nodes))
	}

	if nodes[0].Type != NodeProjectHeader || nodes[0].Expanded {
		t.Errorf("node 0 = %+v, want collapsed ProjectHeader my-stack", nodes[0])
	}

	if nodes[1].Type != NodeContainerItem || nodes[1].Container.Name != "standalone-nginx" {
		t.Errorf("node 1 = %+v, want standalone-nginx", nodes[1])
	}
}

func TestVisibleTreeNodes_AutoExpandOnFilter(t *testing.T) {
	m := NewModel(nil, "docker")
	m.containers = []containerItem{
		{Container: domain.Container{ID: "1", Name: "app-web", ComposeProject: "my-stack", Status: domain.StatusRunning}},
	}

	m.toggleProjectExpanded("my-stack")
	m.filterQuery = "web"

	nodes := m.VisibleTreeNodes()
	if len(nodes) != 2 {
		t.Fatalf("got %d nodes when filtering, expected auto-expansion to show 2", len(nodes))
	}
}
