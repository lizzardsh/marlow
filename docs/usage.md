# Usage

```text
marlow [OPTIONS] <TARGET_IP_OR_DOMAIN>
```

## Options

| Option | Description |
| --- | --- |
| `-p <ports>`, `--ports <ports>` | Comma-separated ports or ranges, for example `21,80,443` or `1-1024`. Defaults to common ports. |
| `-p-`, `--all-ports` | Scan TCP ports `1-65535`. This takes precedence over `-p`. |
| `-sV <modules>`, `--modules <modules>` | Run explicit modules, one per specified port and in the same order. Supported: `ftp`, `http`, `smb`, `ssh`. `--services` is also accepted. |
| `-s-`, `--auto-services` | Automatically run a registered module when an open port matches that module’s default port. |
| `-version-detect`, `--version-detect` | Probe open ports for service banners and basic version details. HTTP probes report status and the `Server` header; HTTPS probes also report the negotiated TLS version. |
| `-Uw <path>`, `--userlist <path>` | Username wordlist for FTP checks. |
| `-Pw <path>`, `--passlist <path>` | Password wordlist for FTP checks. Must be supplied with a username list. |
| `-timeout <duration>`, `--timeout <duration>` | TCP connection timeout per port. Default: `3s`. |
| `-workers <count>`, `--workers <count>` | Concurrent TCP scan workers. Default: `256`. |
| `-h` | Print help. |

## Examples

Scan selected ports and automatically run matching service modules:

```sh
marlow -p 21,22,80,443,445 -s- 192.0.2.10
```

Scan every TCP port:

```sh
marlow -p- -s- example.org
marlow --all-ports --auto-services example.org
```

Collect basic service and version details from open ports:

```sh
marlow -p 21,22,25,80,443,8080 -version-detect example.org
```

Version detection currently reads greetings from FTP, SSH, SMTP, POP3, and IMAP, and sends an HTTP `HEAD` request on common web ports. It reports the server's own banner or `Server` header where available; results depend on what each service reveals and are not a full Nmap service fingerprint database. `-sV` remains the existing option for explicitly mapping service modules to ports.

Run explicit modules. The service list must have one entry for each specified port, in matching order:

```sh
marlow -p 21,80,445 -sV ftp,http,smb 192.0.2.10
marlow -p 22 -sV ssh 192.0.2.10
marlow --ports 22 --modules ssh 192.0.2.10
```

Run FTP checks with local wordlists:

```sh
marlow -p 21 -sV ftp -Uw users.txt -Pw passwords.txt 192.0.2.10
```

Adjust concurrency and timeout:

```sh
marlow -p 1-1024 -timeout 1s -workers 512 192.0.2.10
```

## Service modules

- **FTP:** Checks anonymous access. If both wordlists are supplied, checks their username/password combinations with up to eight workers.
- **HTTP:** Reports the HTTP status and `Server` header. Port `443` uses HTTPS.
- **SSH:** Reads the server identification banner; it does not attempt authentication.
- **SMB:** Checks whether TCP port `445` is reachable. SMB authentication and share listing are not implemented yet.

Use MARLOW only on systems and networks you are authorized to assess.

The current engine uses TCP connect scans. Nmap-style scan modes such as SYN or UDP are not available yet. See the [Roadmap](roadmap.md) for planned scan and enumeration families.
