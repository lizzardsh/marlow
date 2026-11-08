# Project direction

## Vision

MARLOW is a modular CLI for authorized network surface discovery and focused service enumeration. Each workflow starts with a clear target and task command, then exposes only the options relevant to that task. Shared output, timeouts, and concurrency settings keep the tools consistent across modules.

The product is organized around four related tasks:

1. **Scan:** find reachable TCP and UDP ports and collect basic TCP service details.
2. **Services:** run protocol-specific checks against discovered services.
3. **DNS:** inspect DNS records and discover related hostnames.
4. **Web:** inspect HTTP services and discover paths using user-provided wordlists.

The CLI shape follows those tasks:

```text
marlow scan <target> [options]
marlow services <target> [options]
marlow dns <domain> [options]
marlow web <url> [options]
```

Only `scan` is implemented today. The other commands describe the intended structure; they are not available yet.

## Available today

- Concurrent TCP connect and UDP datagram scans of top ports, chosen ports, ranges, or all ports
- Optional banner and basic version collection for FTP, SSH, SMTP, POP3, IMAP, and HTTP(S) on common web ports
- Optional FTP, HTTP, SSH, and SMB checks
- Configurable scan timeout and worker count

## Next priorities

### 1. Make scan results dependable

- Broaden protocol fingerprints with a small, documented probe set
- Distinguish detected services from guesses based on port numbers
- Add structured output such as JSON for scripts and automation

### 2. Grow service checks

- Add protocol-specific checks that use results from the scan
- Keep authentication checks explicit and tied to user-provided inputs
- Let users map checks to nonstandard ports without making common usage harder

### 3. Add DNS and web workflows

- DNS record and hostname discovery with domain-specific options
- HTTP path discovery with URL, wordlist, status-code, and concurrency options
- Keep task-specific flags under their own commands

## Out of scope for the current scan engine

The scanner uses TCP connect probes and generic UDP datagrams. UDP silence is ambiguous and is reported as `OPEN|FILTERED`; service-specific UDP probes, SYN scanning, and operating-system fingerprinting are not implemented.
