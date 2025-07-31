package cmd

import (
	"testing"

	"github.com/kyokomi/gotodoist/internal/api"
)

func TestFilterProjectsByArchiveStatus(t *testing.T) {
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

	tests := []struct {
		name         string
		projects     []api.Project
		showArchived bool
		want         []api.Project
	}{
		{
			name:         "showArchived=true returns only archived projects",
			projects:     []api.Project{archivedProject, activeProject},
			showArchived: true,
			want:         []api.Project{archivedProject},
		},
		{
			name:         "showArchived=false returns only active projects",
			projects:     []api.Project{archivedProject, activeProject},
			showArchived: false,
			want:         []api.Project{activeProject},
		},
		{
			name:         "empty project list",
			projects:     []api.Project{},
			showArchived: false,
			want:         []api.Project{},
		},
		{
			name:         "all projects archived",
			projects:     []api.Project{archivedProject},
			showArchived: false,
			want:         []api.Project{},
		},
		{
			name:         "all projects active",
			projects:     []api.Project{activeProject},
			showArchived: true,
			want:         []api.Project{},
		},
		{
			name:         "multiple archived projects",
			projects:     []api.Project{archivedProject, archivedProject},
			showArchived: true,
			want:         []api.Project{archivedProject, archivedProject},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterProjectsByArchiveStatus(tt.projects, tt.showArchived)

			if len(got) != len(tt.want) {
				t.Errorf("got %d projects, want %d projects", len(got), len(tt.want))
				return
			}

			for i, project := range got {
				if project.ID != tt.want[i].ID {
					t.Errorf("project[%d]: got ID %s, want ID %s", i, project.ID, tt.want[i].ID)
				}
			}
		})
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
