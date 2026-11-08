<div align="center">
  <img src="docs/assets/marlow-banner.svg" alt="MARLOW banner" width="600">

  <h1>MARLOW</h1>

  <p>Open-source, MIT-licensed network reconnaissance and enumeration toolkit written in Go.</p>
</div>

---

## Vision

MARLOW is a modular command-line tool for authorized network surface discovery. Its workflow is designed to move from finding reachable TCP and UDP ports, to identifying TCP services, to running focused protocol checks. The same command structure will grow to cover DNS and web discovery, with each area using options suited to that task.

## Features

### Available now

- Concurrent TCP connect and UDP datagram scanning with configurable workers and timeout
- Default top 1,000 ports per protocol, explicit ports and ranges, or all ports
- Optional service banner and basic version detection on open ports
- FTP, HTTP, SSH, and SMB reachability/service checks
- Optional FTP checks with user-provided wordlists

### Planned

- DNS enumeration
- HTTP path and directory discovery
- Additional scan and service-enumeration modes

## Quick start

Download a prebuilt binary from [Releases](../../releases/latest), or build from source:

```sh
go build -o marlow ./cmd
./marlow -h
```

Scan common ports, identify basic service versions, and run matching service checks:

```sh
./marlow scan 192.0.2.10 --ports 21,22,80,443,445 --versions --service-checks
```

The default command scans the 1,000 most common TCP ports and the 1,000 most common UDP ports, ranked from Nmap's `nmap-services` frequency data. UDP silence is reported as `OPEN|FILTERED`, since a UDP service may not reply to a generic probe. Use `--protocol tcp` or `--protocol udp` to select one protocol, `--ports` to choose ports, or `--all-ports` to scan the full port range. Add `--versions` for TCP service details or `--service-checks` for built-in TCP checks.

See [Installation](docs/installation.md) for release downloads and publishing, [Usage](docs/usage.md) for commands and options, [Roadmap](docs/roadmap.md) for the project direction, and [Updates](docs/updates.md) for changes.

## Project layout

```text
cmd/                 CLI entry point
internal/cli/        flags and orchestration
pkg/scanner/         worker pool and module registry
pkg/modules/         current FTP, HTTP, SSH, and SMB modules
pkg/banner/          shared ASCII banner
pkg/terminal/        shared color palette and formatting
docs/                installation, usage, update, and roadmap guides
```

## License

MIT
