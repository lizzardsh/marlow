# Updates

## v0.2.1

- Default scans now cover the top 1,000 TCP ports and top 1,000 UDP ports, ranked independently using Nmap `nmap-services` frequency data.
- Added `--protocol tcp|udp|both`; `--ports` and `--all-ports` apply to the selected protocols.
- UDP results distinguish replies (`OPEN`), ICMP unreachable (`CLOSED`), and silence (`OPEN|FILTERED`).

## v0.2.0 — Options

- Standardized usage around `marlow scan <target> [options]`.
- Added readable options: `--ports`, `--all-ports`, `--versions`, `--service-checks`, and `--modules`.
- Kept legacy scan flags accepted for existing scripts.
- Module selection now names a set of checks and matches them against their default ports; modules no longer need to be listed one-for-one with scanned ports.

### Current scan scope

MARLOW performs TCP connect scans and generic UDP datagram probes. `--all-ports` scans all ports for the selected protocol(s); SYN scanning is not implemented.

The longer-term project scope is a shared enumeration toolkit: port discovery and service checks first, followed by DNS enumeration and HTTP path discovery. See the [Roadmap](roadmap.md); planned items are not available in the current release.

## Updating an installed Kali binary

After a release is published, download and replace the installed binary:

```sh
curl -fL -o /tmp/marlow https://github.com/lizzardsh/marlow/releases/latest/download/marlow-linux-amd64
sudo install -m 755 /tmp/marlow /usr/local/bin/marlow
marlow -h
```

For ARM64, replace `marlow-linux-amd64` with `marlow-linux-arm64`. Release generation steps are in [Installation](installation.md#publish-a-release).
