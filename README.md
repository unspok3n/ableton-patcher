# Ableton Patcher

Customizable license key patcher for [Ableton Live](https://www.ableton.com/en/live/) with an auto installation search & keygen. Based on [devilAPI/abletonCracker](https://github.com/devilAPI/abletonCracker) and [rufoa/ableton](https://github.com/rufoa/ableton). Supports *Windows* & *MacOS*

![Screenshot](/screenshots/main.png?raw=true "Screenshot")

Setup
---
1. Download or build.

2. Run and select the desired option:
    - Patch - replaces the `original_public_key` with the public key from the `private_key`
    - Unpatch - replaces the public key from the `private_key` with the `original_public_key`
    - Deauthorize - removes `Unlock.json` for the selected Live installation
    - Generate license - creates an authorize file using the `private_key` based on the provided HWID
    - Generate DSA key pair - generates a new custom DSA key pair that can be used to patch installations and generate licenses

The config file (`ableton-patcher-config.yml`) is not required for running but can be created manually, it accepts 2 options:

| Option                | Description                | Default                     |
|-----------------------|----------------------------|-----------------------------|
| `private_key`         | DER-encoded DSA key pair   | *R2R Team Key*              |
| `original_public_key` | DER-encoded DSA public key | *Original Ableton Live Key* |

Note: on MacOS (tested on m1, 26.0), you will have to run `codesign --force --deep --sign - /Applications/Ableton\ Live\ 12\ Suite.app/` after patching, replacing 12 and Suite with your local version. Otherwise, the app will crash at launch, returning a codesigning error in the report.
