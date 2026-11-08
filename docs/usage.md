# Usage

All current network operations use the `scan` command:

```text
marlow scan <target> [options]
```

The target can be an IP address or a hostname. Options can appear before or after the target.

## Common workflows

Scan the 1,000 most common ports for TCP and UDP:

```sh
marlow scan 192.0.2.10
```

Scan selected ports, show basic service and version details, and run the built-in checks for recognized services:

```sh
marlow scan example.org --ports 21,22,80,443,445 --versions --service-checks
```

Scan all TCP and UDP ports and collect TCP service details:

```sh
marlow scan 192.0.2.10 --all-ports --versions
```

Run only selected built-in modules. Module selection automatically enables service checks and applies each module to its recognized default port:

```sh
marlow scan 192.0.2.10 --modules http,ssh
```

Run FTP checks with supplied wordlists:

```sh
marlow scan 192.0.2.10 --modules ftp --ftp-userlist users.txt --ftp-passlist passwords.txt
```

## Options

| Option | Description |
| --- | --- |
| `--ports <list>` | Comma-separated ports or ranges applied to selected protocols. Default: top 1,000 ports per protocol. Example: `22,80,443` or `1-1024`. |
| `--protocol <name>` | `tcp`, `udp`, or `both`. Default: `both`. |
| `--all-ports` | Scan ports `1-65535` for selected protocols; overrides `--ports`. |
| `--versions` | Read service banners and collect basic version details from open ports. HTTP checks report status and the `Server` header; HTTPS checks also report negotiated TLS version. |
| `--service-checks` | Run built-in checks for recognized open services. |
| `--modules <list>` | Limit built-in checks to a comma-separated list: `ftp,http,smb,ssh`. Implies `--service-checks`. Modules are matched to their default ports. |
| `--ftp-userlist <path>` | FTP username wordlist. Requires `--ftp-passlist`. |
| `--ftp-passlist <path>` | FTP password wordlist. Requires `--ftp-userlist`. |
| `--timeout <duration>` | Per-connection timeout. Default: `3s`. |
| `--workers <count>` | Concurrent scan workers per protocol. Default: `256`. |
| `-h`, `--help` | Show scan help. Use `marlow help` for the top-level command. |

Version details depend on the protocol's greeting or HTTP headers. The current detector covers FTP, SSH, SMTP, POP3, IMAP, and HTTP(S) on common web ports; it is a lightweight fingerprinting feature, not a broad service-probe database.

## Service checks

- **FTP:** Checks anonymous access. When both wordlists are supplied, it also tries their username/password combinations.
- **HTTP:** Reports the HTTP status and `Server` header. Port `443` uses HTTPS.
- **SSH:** Reads the server identification banner; it does not attempt authentication.
- **SMB:** Checks whether TCP port `445` is reachable.

TCP uses connect scans. UDP sends a generic datagram; a reply means `OPEN`, an ICMP port-unreachable response means `CLOSED`, and silence is reported as `OPEN|FILTERED` because the service may ignore the probe or filtering may prevent a response. UDP service fingerprinting is not implemented. Use MARLOW only on systems and networks you are authorized to assess.

The default TCP and UDP port rankings are bundled from the frequency fields in Nmap's [`nmap-services`](https://github.com/nmap/nmap/blob/master/nmap-services) dataset; each protocol has its own top 1,000 list.
