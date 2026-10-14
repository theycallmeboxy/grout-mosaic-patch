package saves

import "grout/settings"

// deviceSaveFolders narrows a platform's save folders to the one this device
// syncs when grout runs with a device profile: the mapped folder, or the
// firmware's default. Devices sharing a card can run different cores for the
// same platform, and scanning every core's folder would sync another device's
// saves as this one's.
func deviceSaveFolders(config *settings.Config, rommFSSlug string, dirs []string) []string {
	if config == nil || len(dirs) < 2 || settings.ActiveProfile() == "" {
		return dirs
	}
	return []string{chosenEmulator(*config, rommFSSlug, dirs)}
}
