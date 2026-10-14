package settings

import (
	"encoding/json"
	"errors"
	"grout/files"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

// A device profile lets one card carry grout between handhelds that run
// different emulator cores for the same platform. Hosts, credentials and every
// other setting stay in config.json and are shared by all of them; what
// identifies a device to RomM and which folder its saves live in are kept per
// device in profiles/<name>.json.
//
// Profiles are off unless a profile name is known, so every firmware other
// than spruce behaves exactly as before.

// ProfileEnvVar names the device profile. Without it, spruce's PLATFORM
// (Brick, SmartPro, Flip, A30, ...) names it.
const ProfileEnvVar = "GROUT_PROFILE"

const profilesDirName = "profiles"

type DeviceProfile struct {
	// Hosts holds this device's registration with each host, keyed by host URL.
	Hosts                 map[string]HostDevice `json:"hosts,omitempty"`
	SaveDirectoryMappings map[string]string     `json:"save_directory_mappings,omitempty"`
}

type HostDevice struct {
	// ClientDeviceID is how RomM recognises a device when it pairs again, so a
	// shared one would hand one device another's registration.
	ClientDeviceID      string `json:"client_device_identifier,omitempty"`
	DeviceID            string `json:"device_id,omitempty"`
	DeviceName          string `json:"device_name,omitempty"`
	DeviceClientVersion string `json:"device_client_version,omitempty"`
}

// ActiveProfile is the name of this device's profile, or "" when grout runs
// without one.
func ActiveProfile() string {
	name := os.Getenv(ProfileEnvVar)
	if name == "" && strings.EqualFold(os.Getenv("CFW"), "SPRUCE") {
		name = os.Getenv("PLATFORM")
	}
	// The name becomes a file name.
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		}
		return -1
	}, name)
}

func profilePathFor(configPath, name string) string {
	return filepath.Join(filepath.Dir(configPath), profilesDirName, name+".json")
}

// profileSlotPreferencesPath is where the active profile keeps its save slot
// choices, or "" without a profile. A slot choice outranks the slot a device
// last synced, so a shared one would move every device onto the slot picked
// on any of them.
func profileSlotPreferencesPath(configPath string) string {
	name := ActiveProfile()
	if name == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(configPath), profilesDirName, name+"."+SlotPreferencesFileName)
}

// applyProfile replaces the device's own settings in config with its profile's.
//
// A device seen for the first time takes over the registration, save mappings
// and slot choices in config.json only when no device has a profile yet, which
// is the card's first launch with profiles. Every later device starts
// unregistered, on the firmware's default save folders and slots, so it never
// syncs as another device or into another device's core folder or slot.
func applyProfile(config *Config, configPath string) {
	name := ActiveProfile()
	if name == "" {
		return
	}
	path := profilePathFor(configPath, name)

	profile, err := readProfile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		if !anyProfileExists(configPath) {
			profile = profileOf(*config)
			claimSlotPreferences(configPath)
		}
		if err := writeProfile(path, profile); err != nil {
			slog.Default().Warn("Could not create device profile", "path", path, "error", err)
		}
	case err != nil:
		// Set aside rather than overwritten on the next save, so the device's
		// registration can still be recovered from it.
		slog.Default().Warn("Could not read device profile, starting it fresh", "path", path, "error", err)
		if err := os.Rename(path, path+".bad"); err != nil {
			slog.Default().Warn("Could not set aside unreadable device profile", "path", path, "error", err)
		}
	}

	for i := range config.Hosts {
		device := profile.Hosts[config.Hosts[i].URL()]
		config.Hosts[i].ClientDeviceID = device.ClientDeviceID
		config.Hosts[i].DeviceID = device.DeviceID
		config.Hosts[i].DeviceName = device.DeviceName
		config.Hosts[i].DeviceClientVersion = device.DeviceClientVersion
	}
	config.SaveDirectoryMappings = maps.Clone(profile.SaveDirectoryMappings)
}

// splitProfile writes the device's own settings to its profile and returns
// what is left to write to config.json.
func splitProfile(config *Config, configPath string) (*Config, error) {
	name := ActiveProfile()
	if name == "" {
		return config, nil
	}
	if err := writeProfile(profilePathFor(configPath, name), profileOf(*config)); err != nil {
		return nil, err
	}

	shared := *config
	shared.Hosts = make([]Host, len(config.Hosts))
	for i, host := range config.Hosts {
		host.ClientDeviceID, host.DeviceID, host.DeviceName, host.DeviceClientVersion = "", "", "", ""
		shared.Hosts[i] = host
	}
	shared.SaveDirectoryMappings = nil
	return &shared, nil
}

// claimSlotPreferences copies the card's slot choices to the first profile.
// The shared file is left in place for a grout without profiles.
func claimSlotPreferences(configPath string) {
	data, err := os.ReadFile(filepath.Join(filepath.Dir(configPath), SlotPreferencesFileName))
	if err != nil {
		return
	}
	path := profileSlotPreferencesPath(configPath)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err == nil {
		err = files.WriteFileAtomic(path, data, 0644)
	}
	if err != nil {
		slog.Default().Warn("Could not copy save slot choices to device profile", "path", path, "error", err)
	}
}

func profileOf(config Config) DeviceProfile {
	profile := DeviceProfile{SaveDirectoryMappings: maps.Clone(config.SaveDirectoryMappings)}
	for _, host := range config.Hosts {
		device := HostDevice{
			ClientDeviceID:      host.ClientDeviceID,
			DeviceID:            host.DeviceID,
			DeviceName:          host.DeviceName,
			DeviceClientVersion: host.DeviceClientVersion,
		}
		if device == (HostDevice{}) {
			continue
		}
		if profile.Hosts == nil {
			profile.Hosts = make(map[string]HostDevice)
		}
		profile.Hosts[host.URL()] = device
	}
	return profile
}

func readProfile(path string) (DeviceProfile, error) {
	var profile DeviceProfile
	data, err := os.ReadFile(path)
	if err != nil {
		return profile, err
	}
	err = json.Unmarshal(data, &profile)
	return profile, err
}

func writeProfile(path string, profile DeviceProfile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	pretty, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}
	return files.WriteFileAtomic(path, pretty, 0644)
}

func anyProfileExists(configPath string) bool {
	found, _ := filepath.Glob(filepath.Join(filepath.Dir(configPath), profilesDirName, "*.json"))
	return len(found) > 0
}
