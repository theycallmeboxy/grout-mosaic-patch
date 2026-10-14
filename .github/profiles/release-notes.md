Patched Grout **{{VERSION}}** for spruceOS: each handheld sharing one SD card gets its own save-sync settings. That means its own RomM device registration, its own save-folder (core) mapping and its own save-slot choices. Without this, a Brick on mGBA and a Flip on gpSP would sync as one device.

> **Pre-release:** builds and unit tests pass, but this hasn't been tested on hardware yet.

### Install
1. You need **stock Grout {{VERSION}}** already on the card. Back up `App/Grout/grout/config.json` and `save_slots.json`.
2. Extract `grout-profiles-{{VERSION}}.zip` to the SD card root and overwrite when asked. It only replaces `grout64` and `grout32`.
3. The first device you launch keeps the existing registration. Each other device shows **Register Device** once.

### Good to know
- Only use this zip with Grout {{VERSION}}.
- A Grout update turns the patch off. Every device then shows as unregistered and nothing syncs. Don't register while it's off; copy the matching zip again and everything comes back.
- Two identical handhelds share a profile. Set `GROUT_PROFILE` on one of them to keep them apart.
- Off spruce, without `GROUT_PROFILE`, behaviour is the same as stock.

Full notes are in `README.txt` inside the zip. Source: branch [`{{BRANCH}}`](https://github.com/theycallmeboxy/grout-mosaic-patch/tree/{{BRANCH}}); the same change is attached as a single `.patch`.

SHA-256 (zip): `{{SHA256}}`

---
Unofficial build. It is not made, endorsed or supported by the [Grout](https://github.com/rommapp/grout), [RomM](https://github.com/rommapp/romm) or [spruceOS](https://github.com/spruceUI/spruceOS) projects, so please report problems [here](https://github.com/theycallmeboxy/grout-mosaic-patch/issues) rather than upstream. Grout is © 2025 Brandon T. Kowalski and the Grout contributors and is MIT-licensed; its license ships in the zip as `App/Grout/LICENSE`. The patch is released under the same license.
