Grout per-device profiles (patched build for Grout {{VERSION}})
=============================================================

WHAT THIS IS
  One spruceOS SD card often moves between several handhelds. Each one may
  use a different emulator core for the same system. This patch gives each
  device its own save-sync settings, so they no longer share one RomM device
  registration, one save-folder mapping or one set of save slots.

REQUIREMENTS
  - spruceOS with stock Grout {{VERSION}} already installed on the card.
  - Do not use a zip built for a different Grout version. Only the binaries
    are replaced, so they must match your stock launch.sh and libraries.

INSTALL
  1. Back up App/Grout/grout/config.json and save_slots.json.
  2. Extract this zip to the SD card root.
  3. Overwrite when asked. It only replaces grout64 and grout32.
  4. Optional: delete README.txt from the card root afterwards.

HOW IT WORKS
  The profile name is the GROUT_PROFILE environment variable if set.
  Otherwise it is spruce's PLATFORM (Brick, SmartPro, Flip, A30, Pixel2...).
  Off spruce and without GROUT_PROFILE, nothing changes.

FIRST LAUNCH ON EACH DEVICE
  - The first device you launch keeps the card's existing registration,
    save-folder mappings and slot choices. They move into its profile.
  - Every other device starts unregistered, with default folders and slots.
    It shows the normal "Register Device" prompt once. Register it there.
  - Hosts, credentials and other settings stay shared in config.json.
  - Pairing login (QR / approve on server) gives each device its own pairing
    identity. The token is still shared.
  - Save scan only reads this device's chosen core folder per system, under
    /mnt/SDCARD/Saves/saves/<CoreName>/.

AFTER A GROUT UPDATE
  An update puts the stock binaries back. Then every device shows as
  unregistered and Save Sync is hidden. Nothing syncs. This is safe.
  Copy the matching patched zip again and everything returns from profiles/.
  Do not register in stock Grout meanwhile. It adds an extra RomM device.
  It is harmless, but it clutters the server.

UNINSTALL / BACK TO STOCK
  Restore the stock grout64 and grout32 (reinstall Grout). Stock Grout
  ignores profiles/, so you must register once more. To go back fully, you
  can also restore your backed-up config.json and save_slots.json.

KNOWN LIMITS
  - Two identical handhelds share a profile (same PLATFORM). Set
    GROUT_PROFILE on one of them to separate them, e.g. add
    "export GROUT_PROFILE=Brick2" near the top of App/Grout/launch.sh.
    A Grout update replaces launch.sh, so add it again afterwards.
  - If you change a device's core, re-pick its save folder mapping there.
  - To keep different cores' saves apart on RomM, pick different save slots
    per device for that game.
  - If a host's URL is changed, other devices lose their registration for
    it and must register again.
  - An unreadable profile file is renamed to <name>.json.bad and that device
    starts fresh. You can recover the registration by hand from the .bad
    file.

FILES IT CREATES (in App/Grout/grout/)
  profiles/<name>.json              device ID/name, pairing identifier,
                                    save-folder mappings
  profiles/<name>.save_slots.json   save slot choices
  profiles/<name>.json.bad          only if a profile was unreadable
  Device fields are removed from config.json.
