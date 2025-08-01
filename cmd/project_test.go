package cmd

import (
	"testing"

	"github.com/kyokomi/gotodoist/internal/api"
)

func TestProjectFilters(t *testing.T) {
	archivedProject := api.Project{
		ID:         "proj1",
		Name:       "Archived Project",
		IsArchived: true,
	}
	activeProject := api.Project{
		ID:         "proj2",
		Name:       "Active Project",
		IsArchived: false,
	}

	t.Run("filterArchivedProjects", func(t *testing.T) {
		tests := []struct {
			name     string
			projects []api.Project
			want     []api.Project
		}{
			{
				name:     "returns only archived projects",
				projects: []api.Project{archivedProject, activeProject},
				want:     []api.Project{archivedProject},
			},
			{
				name:     "empty project list",
				projects: []api.Project{},
				want:     []api.Project{},
			},
			{
				name:     "all projects active",
				projects: []api.Project{activeProject},
				want:     []api.Project{},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := filterArchivedProjects(tt.projects)
				assertProjectsEqual(t, got, tt.want)
			})
		}
	})

	t.Run("filterActiveProjects", func(t *testing.T) {
		tests := []struct {
			name     string
			projects []api.Project
			want     []api.Project
		}{
			{
				name:     "returns only active projects",
				projects: []api.Project{archivedProject, activeProject},
				want:     []api.Project{activeProject},
			},
			{
				name:     "empty project list",
				projects: []api.Project{},
				want:     []api.Project{},
			},
			{
				name:     "all projects archived",
				projects: []api.Project{archivedProject},
				want:     []api.Project{},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := filterActiveProjects(tt.projects)
				assertProjectsEqual(t, got, tt.want)
			})
		}
	})
}

func assertProjectsEqual(t *testing.T, got, want []api.Project) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("got %d projects, want %d projects", len(got), len(want))
		return
	}

	for i, project := range got {
		if project.ID != want[i].ID {
			t.Errorf("project[%d]: got ID %s, want ID %s", i, project.ID, want[i].ID)
		}
	}
}

func TestGetProjectListTitle(t *testing.T) {
	tests := []struct {
		name          string
		showArchived  bool
		showFavorites bool
		wantTitle     string
		wantEmpty     string
	}{
		{
			name:          "default (active projects)",
			showArchived:  false,
			showFavorites: false,
			wantTitle:     "📁 Projects",
			wantEmpty:     "📁 No projects found",
		},
		{
			name:          "archived projects",
			showArchived:  true,
			showFavorites: false,
			wantTitle:     "📦 Archived Projects",
			wantEmpty:     "📦 No archived projects found",
		},
		{
			name:          "favorite projects",
			showArchived:  false,
			showFavorites: true,
			wantTitle:     "⭐ Favorite Projects",
			wantEmpty:     "⭐ No favorite projects found",
		},
		{
			name:          "archived takes precedence over favorites",
			showArchived:  true,
			showFavorites: true,
			wantTitle:     "📦 Archived Projects",
			wantEmpty:     "📦 No archived projects found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTitle, gotEmpty := getProjectListTitle(tt.showArchived, tt.showFavorites)

			if gotTitle != tt.wantTitle {
				t.Errorf("title: got %s, want %s", gotTitle, tt.wantTitle)
			}
			if gotEmpty != tt.wantEmpty {
				t.Errorf("empty message: got %s, want %s", gotEmpty, tt.wantEmpty)
			}
		})
	}
}
