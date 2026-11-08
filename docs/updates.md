# Updates

## Unreleased

### Options

- Standardized usage around `marlow scan <target> [options]`.
- Added readable options: `--ports`, `--all-ports`, `--versions`, `--service-checks`, and `--modules`.
- Kept legacy scan flags accepted for existing scripts.
- Module selection now names a set of checks and matches them against their default ports; modules no longer need to be listed one-for-one with scanned ports.

### Current scan scope

MARLOW currently performs TCP connect scans. `--all-ports` scans all TCP ports; it does not enable SYN, UDP, or other scan modes.

The longer-term project scope is a shared enumeration toolkit: port discovery and service checks first, followed by DNS enumeration and HTTP path discovery. See the [Roadmap](roadmap.md); planned items are not available in the current release.

## Updating an installed Kali binary

After a release is published, download and replace the installed binary:

```sh
curl -fL -o /tmp/marlow https://github.com/lizzardsh/marlow/releases/latest/download/marlow-linux-amd64
sudo install -m 755 /tmp/marlow /usr/local/bin/marlow
marlow -h
```

For ARM64, replace `marlow-linux-amd64` with `marlow-linux-arm64`. Release generation steps are in [Installation](installation.md#publish-a-release).
