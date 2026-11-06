# Roadmap

MARLOW aims to put multiple network discovery and enumeration workflows behind one CLI and a shared module registry. Port scanning is the first module family, not the full product scope.

## Available now

- TCP connect scans for selected ports, ranges, or all TCP ports
- FTP anonymous-access checks and optional checks against supplied wordlists
- HTTP status and server-header collection
- SSH identification-banner collection
- SMB TCP reachability check

## Planned module families

### Port discovery

- Additional scan modes beyond TCP connect
- Clear scan profiles and Nmap-like option compatibility where practical
- UDP support as a separately selected mode

### DNS enumeration

- DNS record and hostname discovery workflows
- Target and resolver options specific to DNS modules

### Web discovery

- HTTP path and directory discovery similar in purpose to dirsearch
- URL, wordlist, status-code, and concurrency options scoped to web modules

### Service enumeration

- Deeper protocol-specific checks for services already detected by the port scanner
- Independent modules that can be registered without changing the core engine

## CLI direction

As module families grow, commands can be grouped by task while sharing common output and configuration:

```text
marlow ports <target> [options]
marlow services <target> [options]
marlow dns <domain> [options]
marlow web <url> [options]
```

These command groups are a design direction, not current commands. The current CLI remains `marlow [OPTIONS] <TARGET>`; see [Usage](usage.md).

The module configuration will need to grow beyond the current IP-and-port shape so DNS modules can accept domains and web modules can accept URLs, each with task-specific options. That keeps new enumeration families from being forced into port-scanner settings.
