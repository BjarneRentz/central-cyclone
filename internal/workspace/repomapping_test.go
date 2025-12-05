package workspace

import "testing"

func TestDefaultRepoMapper_GetFolderName(t *testing.T) {
	tests := []struct {
		name    string
		repoURL string
		want    string
		wantErr bool
	}{
		{
			name:    "GitHub simple",
			repoURL: "https://github.com/org/repo",
			want:    "org_repo",
			wantErr: false,
		},
		{
			name:    "GitHub with .git suffix",
			repoURL: "https://github.com/org/repo.git",
			want:    "org_repo",
			wantErr: false,
		},
		{
			name:    "Azure DevOps",
			repoURL: "https://dev.azure.com/my-org/my-project/_git/my-repo",
			want:    "my-org_my-project_my-repo",
			wantErr: false,
		},
		{
			name:    "Invalid GitHub URL - missing repo",
			repoURL: "https://github.com/org",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid Azure DevOps URL - missing _git",
			repoURL: "https://dev.azure.com/org/project/repo",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Unsupported git host",
			repoURL: "https://gitlab.com/org/repo",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid URL",
			repoURL: "://invalid-url",
			want:    "",
			wantErr: true,
		},
	}

	mapper := DefaultRepoMapper{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapper.GetFolderName(tt.repoURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFolderName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetFolderName() = %v, want %v", got, tt.want)
			}
		})
	}
}
