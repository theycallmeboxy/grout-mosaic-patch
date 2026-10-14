package saves

import (
	"slices"
	"testing"

	"grout/settings"
)

func TestDeviceSaveFolders(t *testing.T) {
	dirs := []string{"gpSP", "mGBA"}
	mapped := &settings.Config{SaveDirectoryMappings: map[string]string{"gba": "mGBA"}}

	t.Setenv(settings.ProfileEnvVar, "")
	t.Setenv("CFW", "")
	if got := deviceSaveFolders(mapped, "gba", dirs); !slices.Equal(got, dirs) {
		t.Errorf("without a profile, folders = %v, want all of %v", got, dirs)
	}

	t.Setenv(settings.ProfileEnvVar, "Flip")
	if got := deviceSaveFolders(mapped, "gba", dirs); !slices.Equal(got, []string{"mGBA"}) {
		t.Errorf("with a profile, folders = %v, want only the mapped mGBA", got)
	}
	if got := deviceSaveFolders(&settings.Config{}, "gba", dirs); !slices.Equal(got, []string{"gpSP"}) {
		t.Errorf("with a profile and no mapping, folders = %v, want the default gpSP", got)
	}
	if got := deviceSaveFolders(mapped, "gba", []string{"GBA"}); !slices.Equal(got, []string{"GBA"}) {
		t.Errorf("a single folder must be left alone, got %v", got)
	}
}
