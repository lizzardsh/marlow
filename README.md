<div align="center">
  <img src="docs/assets/marlow-banner.svg" alt="MARLOW banner" width="600">

  <h1>MARLOW</h1>

  <p>Open-source, MIT-licensed network reconnaissance and enumeration toolkit written in Go.</p>
</div>

---

## About

MARLOW is intended to bring network discovery and enumeration tasks under one modular CLI. The current release focuses on TCP port scanning and a small set of service checks; DNS enumeration and web content discovery are planned as separate module families.

## Features

### Available now

- Concurrent TCP connect scanning with configurable workers and timeout
- Explicit ports, ranges, or all TCP ports
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

Scan common ports and run the matching service modules:

```sh
./marlow -p 21,22,80,443,445 -s- 192.0.2.10
```

See [Installation](docs/installation.md) for release downloads and publishing, [Usage](docs/usage.md) for current CLI options, [Roadmap](docs/roadmap.md) for planned module families, and [Updates](docs/updates.md) for changes.

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
