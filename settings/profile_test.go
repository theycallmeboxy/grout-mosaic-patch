package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestConfig(t *testing.T, path string, config *Config) {
	t.Helper()
	if err := SaveConfigTo(config, path); err != nil {
		t.Fatalf("SaveConfigTo: %v", err)
	}
}

func loadTestConfig(t *testing.T, path string) *Config {
	t.Helper()
	config, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}
	return config
}

func registeredConfig() *Config {
	return &Config{
		Hosts: []Host{{
			RootURI:        "http://romm.local",
			Port:           8080,
			Token:          "shared-token",
			ClientDeviceID: "client-a",
			DeviceID:       "device-a",
			DeviceName:     "Brick",
		}},
		SaveDirectoryMappings: map[string]string{"gba": "mGBA"},
	}
}

func TestActiveProfile(t *testing.T) {
	t.Setenv(ProfileEnvVar, "")
	t.Setenv("CFW", "")
	t.Setenv("PLATFORM", "Brick")
	if got := ActiveProfile(); got != "" {
		t.Errorf("off spruce, ActiveProfile = %q, want none", got)
	}

	t.Setenv("CFW", "SPRUCE")
	if got := ActiveProfile(); got != "Brick" {
		t.Errorf("on spruce, ActiveProfile = %q, want Brick", got)
	}

	t.Setenv(ProfileEnvVar, "../my brick")
	if got := ActiveProfile(); got != "mybrick" {
		t.Errorf("ActiveProfile = %q, want the name made safe for a file", got)
	}
}

// Without a profile, config.json keeps everything, as it always has.
func TestNoProfile_KeepsDeviceSettingsInConfig(t *testing.T) {
	t.Setenv(ProfileEnvVar, "")
	t.Setenv("CFW", "")
	path := filepath.Join(t.TempDir(), ConfigFileName)

	writeTestConfig(t, path, registeredConfig())
	config := loadTestConfig(t, path)

	if config.Hosts[0].DeviceID != "device-a" || config.SaveDirectoryMappings["gba"] != "mGBA" {
		t.Errorf("device settings lost without a profile: %+v %v", config.Hosts[0], config.SaveDirectoryMappings)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), profilesDirName)); !os.IsNotExist(err) {
		t.Errorf("profiles directory created without a profile")
	}
}

// The first device to launch keeps the card's existing registration; a second
// device starts unregistered, and each keeps its own from then on.
func TestProfiles_SeparateDevicesSharingACard(t *testing.T) {
	t.Setenv("CFW", "")
	path := filepath.Join(t.TempDir(), ConfigFileName)

	// Written by grout before profiles existed.
	t.Setenv(ProfileEnvVar, "")
	writeTestConfig(t, path, registeredConfig())

	t.Setenv(ProfileEnvVar, "Brick")
	brick := loadTestConfig(t, path)
	if brick.Hosts[0].DeviceID != "device-a" || brick.Hosts[0].ClientDeviceID != "client-a" || brick.SaveDirectoryMappings["gba"] != "mGBA" {
		t.Fatalf("first device lost the existing registration: %+v %v", brick.Hosts[0], brick.SaveDirectoryMappings)
	}
	writeTestConfig(t, path, brick)

	t.Setenv(ProfileEnvVar, "Flip")
	flip := loadTestConfig(t, path)
	if flip.Hosts[0].DeviceID != "" || flip.Hosts[0].ClientDeviceID != "" || flip.SaveDirectoryMappings != nil {
		t.Fatalf("second device inherited the first's: %+v %v", flip.Hosts[0], flip.SaveDirectoryMappings)
	}
	if flip.Hosts[0].Token != "shared-token" {
		t.Errorf("second device lost the shared credentials")
	}
	flip.Hosts[0].ClientDeviceID = "client-b"
	flip.Hosts[0].DeviceID = "device-b"
	flip.Hosts[0].DeviceName = "Flip"
	flip.SaveDirectoryMappings = map[string]string{"gba": "gpSP"}
	writeTestConfig(t, path, flip)

	t.Setenv(ProfileEnvVar, "Brick")
	brick = loadTestConfig(t, path)
	if brick.Hosts[0].DeviceID != "device-a" || brick.SaveDirectoryMappings["gba"] != "mGBA" {
		t.Errorf("first device picked up the second's settings: %+v %v", brick.Hosts[0], brick.SaveDirectoryMappings)
	}

	t.Setenv(ProfileEnvVar, "Flip")
	flip = loadTestConfig(t, path)
	if flip.Hosts[0].DeviceID != "device-b" || flip.Hosts[0].ClientDeviceID != "client-b" || flip.SaveDirectoryMappings["gba"] != "gpSP" {
		t.Errorf("second device lost its settings: %+v %v", flip.Hosts[0], flip.SaveDirectoryMappings)
	}

	shared, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, leaked := range []string{"device-a", "device-b", "client-a", "client-b", "save_directory_mappings"} {
		if strings.Contains(string(shared), leaked) {
			t.Errorf("config.json still holds %q", leaked)
		}
	}
}

// A slot picked on one device must not move another device onto it, and the
// first device keeps the slots chosen before profiles existed.
func TestProfiles_SlotPreferencesPerDevice(t *testing.T) {
	t.Setenv("CFW", "")
	path := filepath.Join(t.TempDir(), ConfigFileName)

	t.Setenv(ProfileEnvVar, "")
	writeTestConfig(t, path, registeredConfig())
	before := &Config{}
	before.SetSlotPreference(7, "mgba")
	if err := SaveSlotPreferencesTo(before, slotPreferencesPathFor(path)); err != nil {
		t.Fatal(err)
	}

	t.Setenv(ProfileEnvVar, "Brick")
	if got := loadTestConfig(t, path).GetSlotPreference(7); got != "mgba" {
		t.Fatalf("first device slot = %q, want the existing mgba", got)
	}

	t.Setenv(ProfileEnvVar, "Flip")
	flip := loadTestConfig(t, path)
	if got := flip.GetSlotPreference(7); got != DefaultSaveSlot {
		t.Fatalf("second device slot = %q, want the default", got)
	}
	flip.SetSlotPreference(7, "gpsp")
	if err := SaveSlotPreferencesTo(flip, slotPreferencesPathFor(path)); err != nil {
		t.Fatal(err)
	}

	t.Setenv(ProfileEnvVar, "Brick")
	if got := loadTestConfig(t, path).GetSlotPreference(7); got != "mgba" {
		t.Errorf("first device slot = %q after the second picked gpsp, want mgba", got)
	}
	t.Setenv(ProfileEnvVar, "Flip")
	if got := loadTestConfig(t, path).GetSlotPreference(7); got != "gpsp" {
		t.Errorf("second device slot = %q, want gpsp", got)
	}
}

// An unreadable profile is set aside, not overwritten, so the registration in
// it can be recovered.
func TestProfiles_UnreadableProfileIsSetAside(t *testing.T) {
	t.Setenv("CFW", "")
	path := filepath.Join(t.TempDir(), ConfigFileName)

	t.Setenv(ProfileEnvVar, "")
	writeTestConfig(t, path, registeredConfig())

	t.Setenv(ProfileEnvVar, "Brick")
	writeTestConfig(t, path, loadTestConfig(t, path))
	profile := profilePathFor(path, "Brick")
	if err := os.WriteFile(profile, []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}

	brick := loadTestConfig(t, path)
	if brick.Hosts[0].DeviceID != "" {
		t.Errorf("unreadable profile gave device ID %q, want none", brick.Hosts[0].DeviceID)
	}
	if kept, err := os.ReadFile(profile + ".bad"); err != nil || string(kept) != "{not json" {
		t.Errorf("unreadable profile not set aside: %q, %v", kept, err)
	}
}
