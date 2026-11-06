# Installation

## Kali Linux

Download the latest release binary for a standard 64-bit Kali system:

```sh
curl -fL -o /tmp/marlow https://github.com/lizzardsh/marlow/releases/latest/download/marlow-linux-amd64
sudo install -m 755 /tmp/marlow /usr/local/bin/marlow
marlow -h
```

For an ARM64 Kali device (`aarch64` from `uname -m`), use `marlow-linux-arm64` instead.

## Available release files

Every published release includes:

- `marlow-linux-amd64` — standard 64-bit Linux, including most Kali installations
- `marlow-linux-arm64` — ARM64 Linux
- `marlow-windows-amd64.exe` — 64-bit Windows
- `SHA256SUMS` — checksums for the binaries

## Publish a release

Push a version tag to start the GitHub Actions workflow. It checks the Go packages, builds the release files, and publishes them to the repository’s Releases page:

```sh
git tag -a v0.1.0 -m "MARLOW v0.1.0"
git push origin v0.1.0
```

Use a new version tag for each release. The first run appears under the repository’s **Actions** tab; after it succeeds, the assets appear under **Releases**.

## Build from source

To build locally, install Go 1.22 or newer, clone the repository, and run:

```sh
git clone https://github.com/lizzardsh/marlow.git
cd marlow
go build -o marlow ./cmd
./marlow -h
```

To install that local build as a command:

```sh
sudo install -m 755 marlow /usr/local/bin/marlow
```

If the shell is already running as root, omit `sudo`.
