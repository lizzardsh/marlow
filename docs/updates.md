# Updates

## Unreleased

### Options

- Added readable aliases: `--ports`, `--all-ports`, `--modules`, `--services`, `--auto-services`, `--userlist`, and `--passlist`.
- Kept the existing short options working for compatibility.
- `-p-` is now parsed as a dedicated all-ports option and scans TCP ports `1-65535`.
- `-sV` still means an explicit module list in MARLOW. It requires one module per port, in the same order. `--modules` is the clearer long option for new scripts.

`-sV` currently differs from Nmap’s version-detection flag. If MARLOW adopts that Nmap behavior later, the old spelling will be deprecated through a documented version change; `--modules` is the forward-compatible spelling for explicit modules.

### Current scan scope

MARLOW currently performs TCP connect scans. The all-ports option scans all TCP ports; it does not enable SYN, UDP, or other Nmap scan modes. Those can be added later as separately named scan modes without changing the meaning of `-p`, `--timeout`, or `--workers`.

The longer-term project scope is a shared enumeration toolkit: port discovery and service checks first, followed by DNS enumeration and HTTP path discovery. See the [Roadmap](roadmap.md); planned items are not available in the current release.

## Updating an installed Kali binary

After a release is published, download and replace the installed binary:

```sh
curl -fL -o /tmp/marlow https://github.com/lizzardsh/marlow/releases/latest/download/marlow-linux-amd64
sudo install -m 755 /tmp/marlow /usr/local/bin/marlow
marlow -h
```

For ARM64, replace `marlow-linux-amd64` with `marlow-linux-arm64`. Release generation steps are in [Installation](installation.md#publish-a-release).
