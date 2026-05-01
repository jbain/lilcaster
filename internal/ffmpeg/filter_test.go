package ffmpeg

import (
	"regexp"
	"testing"

	"lilcaster/internal/config"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		name    string
		filters []config.FilterEntry
		want    string
	}{
		{
			name:    "empty",
			filters: nil,
			want:    "",
		},
		{
			name:    "scale",
			filters: []config.FilterEntry{{Filter: &config.ScaleFilter{Width: "1280", Height: "720"}}},
			want:    "scale=1280:720",
		},
		{
			name:    "timestamp",
			filters: []config.FilterEntry{{Filter: &config.TimestampFilter{}}},
			want:    timestampFilter,
		},
		{
			name:    "custom verbatim",
			filters: []config.FilterEntry{{Filter: &config.CustomFilter{String: "eq=brightness=0.1"}}},
			want:    "eq=brightness=0.1",
		},
		{
			name: "all three combined",
			filters: []config.FilterEntry{
				{Filter: &config.ScaleFilter{Width: "-2", Height: "480"}},
				{Filter: &config.TimestampFilter{}},
				{Filter: &config.CustomFilter{String: "eq=brightness=0.1"}},
			},
			want: "scale=-2:480," + timestampFilter + ",eq=brightness=0.1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Build(tc.filters)
			if got != tc.want {
				t.Errorf("Build():\n got  %q\n want %q", got, tc.want)
			}
		})
	}
}

func TestBuildPhrase(t *testing.T) {
	got := Build([]config.FilterEntry{{Filter: &config.PhraseFilter{}}})
	re := regexp.MustCompile(`^drawtext=text='[a-z]+-[a-z]+-[a-z]+':x=10:y=H-th-10:fontsize=28:fontcolor=[a-z]+:box=1:boxcolor=black@0\.5$`)
	if !re.MatchString(got) {
		t.Errorf("Build() phrase = %q, want match pattern %q", got, re.String())
	}
}

func TestBuildPhraseInChain(t *testing.T) {
	got := Build([]config.FilterEntry{
		{Filter: &config.ScaleFilter{Width: "1280", Height: "720"}},
		{Filter: &config.PhraseFilter{}},
	})
	re := regexp.MustCompile(`^scale=1280:720,drawtext=text='[a-z]+-[a-z]+-[a-z]+':x=10:y=H-th-10:fontsize=28:fontcolor=[a-z]+:box=1:boxcolor=black@0\.5$`)
	if !re.MatchString(got) {
		t.Errorf("Build() scale+phrase = %q, want match pattern %q", got, re.String())
	}
}
