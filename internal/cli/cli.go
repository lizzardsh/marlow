package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"marlow/pkg/fingerprint"
	"marlow/pkg/scanner"
	"marlow/pkg/terminal"
)

type options struct {
	ports         string
	protocol      string
	services      string
	allPorts      bool
	serviceChecks bool
	versionInfo   bool
	userPath      string
	passPath      string
	timeout       time.Duration
	workers       int
}

func Run(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printRootUsage()
		return nil
	}
	if args[0] == "scan" {
		args = args[1:]
	}
	opts, target, err := parseArgs(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if err := validateTarget(target); err != nil {
		return err
	}
	portSpec := opts.ports
	if opts.allPorts {
		portSpec = "1-65535"
	}
	var ports []string
	if portSpec != "" {
		ports, err = parsePorts(portSpec)
		if err != nil {
			return err
		}
	}
	if err := validateServices(opts.services); err != nil {
		return err
	}
	users, err := readLines(opts.userPath)
	if err != nil {
		return err
	}
	passwords, err := readLines(opts.passPath)
	if err != nil {
		return err
	}

	protocols := selectedProtocols(opts.protocol)
	fmt.Println(terminal.Paint(terminal.Info, fmt.Sprintf("[*] Scanning %s (%s; timeout %s)", target, strings.Join(protocols, ","), opts.timeout)))
	engine := scanner.Engine{Workers: opts.workers, Timeout: opts.timeout}
	openPorts := make([]string, 0)
	anyFound := false
	for _, protocol := range protocols {
		protocolPorts := ports
		if len(protocolPorts) == 0 {
			protocolPorts = scanner.DefaultPorts(protocol)
		}
		fmt.Println(terminal.Paint(terminal.Info, fmt.Sprintf("[*] %s: scanning %d ports", strings.ToUpper(protocol), len(protocolPorts))))
		results := engine.ScanProtocol(context.Background(), target, protocolPorts, protocol)
		for _, result := range results {
			if result.State == "OPEN" || result.State == "OPEN|FILTERED" {
				anyFound = true
				color := terminal.Success
				if result.State == "OPEN|FILTERED" {
					color = terminal.Warning
				}
				fmt.Println(terminal.Paint(color, fmt.Sprintf("[+] %s: %s/%s", result.State, result.Port, protocol)))
			}
			if result.Open && protocol == "tcp" {
				openPorts = append(openPorts, result.Port)
			}
		}
	}
	if !anyFound {
		fmt.Println(terminal.Paint(terminal.Error, "[-] No responsive ports found"))
	}
	if len(openPorts) == 0 {
		return nil
	}
	if opts.versionInfo {
		for _, port := range openPorts {
			result := fingerprint.Detect(context.Background(), target, port, opts.timeout)
			fmt.Println(terminal.Paint(terminal.Info, fmt.Sprintf("[*] SERVICE: %s:%s service=%s version=%s", target, port, result.Service, result.Version)))
		}
	}
	return runModules(target, openPorts, opts, users, passwords)
}

func printRootUsage() {
	fmt.Println("MARLOW — network discovery and service enumeration")
	fmt.Println("\nUsage:\n  marlow scan <target> [options]\n  marlow help")
	fmt.Println("\nRun 'marlow scan --help' for scan options.")
}

func parseArgs(args []string) (options, string, error) {
	fs := flag.NewFlagSet("marlow", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	opts := options{protocol: "both"}
	var showHelp bool
	fs.StringVar(&opts.ports, "ports", "", "Ports or ranges, for example 22,80,443 or 1-1024; default: top 1,000 per protocol")
	fs.StringVar(&opts.protocol, "protocol", "both", "Protocol to scan: tcp, udp, or both (default: both)")
	fs.BoolVar(&opts.allPorts, "all-ports", false, "Scan ports 1-65535 for selected protocols")
	fs.BoolVar(&opts.versionInfo, "versions", false, "Collect service banners and basic version details from open ports")
	fs.BoolVar(&opts.serviceChecks, "service-checks", false, "Run built-in checks for recognized open services")
	fs.StringVar(&opts.services, "modules", "", "Limit service checks to a comma-separated list: ftp,http,smb,ssh")
	fs.StringVar(&opts.services, "services", "", "Alias for --modules")
	fs.StringVar(&opts.userPath, "ftp-userlist", "", "FTP username wordlist")
	fs.StringVar(&opts.passPath, "ftp-passlist", "", "FTP password wordlist")
	fs.DurationVar(&opts.timeout, "timeout", 3*time.Second, "Connection timeout, for example 3s")
	fs.IntVar(&opts.workers, "workers", 256, "Concurrent port scan workers")
	fs.BoolVar(&showHelp, "help", false, "Show scan help")
	fs.BoolVar(&showHelp, "h", false, "Show scan help")

	// Legacy spellings stay accepted for existing scripts. The help output only
	// shows the clear long-form options above.
	fs.StringVar(&opts.ports, "p", "", "")
	fs.BoolVar(&opts.allPorts, "p-", false, "")
	fs.BoolVar(&opts.versionInfo, "sV", false, "")
	fs.BoolVar(&opts.versionInfo, "version-detect", false, "")
	fs.BoolVar(&opts.serviceChecks, "s-", false, "")
	fs.StringVar(&opts.userPath, "Uw", "", "")
	fs.StringVar(&opts.passPath, "Pw", "", "")
	fs.BoolVar(&opts.serviceChecks, "auto-services", false, "")
	fs.BoolVar(&opts.versionInfo, "service-version", false, "")
	fs.StringVar(&opts.userPath, "userlist", "", "")
	fs.StringVar(&opts.passPath, "passlist", "", "")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: marlow scan <target> [options]")
		fmt.Fprintln(fs.Output(), "\nOptions:\n  --ports <list>         Ports or ranges (default: top 1,000 per protocol)\n  --protocol <name>      tcp, udp, or both (default: both)\n  --all-ports            Scan ports 1-65535 for selected protocols\n  --versions             Collect service banners and basic version details\n  --service-checks       Run built-in checks for recognized TCP services\n  --modules <list>       Limit checks to ftp,http,smb,ssh\n  --ftp-userlist <path>  FTP username wordlist\n  --ftp-passlist <path>  FTP password wordlist\n  --timeout <duration>   Connection timeout (default: 3s)\n  --workers <count>      Concurrent port scan workers (default: 256)\n  -h, --help             Show this help")
		fmt.Fprintln(fs.Output(), "\nExamples:\n  marlow scan example.org\n  marlow scan example.org --ports 21,22,80,443 --versions --service-checks\n  marlow scan 192.0.2.10 --all-ports --versions")
	}

	args = moveTargetLast(args)
	if err := fs.Parse(args); err != nil {
		return options{}, "", err
	}
	if showHelp {
		fs.Usage()
		return options{}, "", flag.ErrHelp
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return options{}, "", fmt.Errorf("scan needs exactly one target")
	}
	if opts.timeout <= 0 || opts.workers < 1 {
		return options{}, "", fmt.Errorf("timeout and worker count must be positive")
	}
	if opts.protocol != "tcp" && opts.protocol != "udp" && opts.protocol != "both" {
		return options{}, "", fmt.Errorf("--protocol must be tcp, udp, or both")
	}
	if (opts.userPath == "") != (opts.passPath == "") {
		return options{}, "", fmt.Errorf("--ftp-userlist and --ftp-passlist must be supplied together")
	}
	if len(splitNames(opts.services)) > 0 && !opts.serviceChecks {
		opts.serviceChecks = true
	}
	return opts, strings.TrimSpace(fs.Arg(0)), nil
}

func validateTarget(target string) error {
	if target == "" {
		return fmt.Errorf("target cannot be empty")
	}
	if net.ParseIP(target) == nil {
		if _, err := net.LookupHost(target); err != nil {
			return fmt.Errorf("cannot resolve target %q: %w", target, err)
		}
	}
	return nil
}

func runModules(target string, openPorts []string, opts options, users, passwords []string) error {
	serviceNames := splitNames(opts.services)
	for _, port := range openPorts {
		if !opts.serviceChecks {
			continue
		}
		service := scanner.ServiceForPort(port)
		if service == "" || (len(serviceNames) > 0 && !contains(serviceNames, service)) {
			continue
		}

		module, ok := scanner.Module(service)
		if !ok {
			continue
		}
		fmt.Println(terminal.Paint(terminal.Info, fmt.Sprintf("[*] Running %s module on %s:%s", service, target, port)))
		config := scanner.ModuleConfig{TargetIP: target, Port: port, UserLogins: users, PassLogins: passwords}
		if err := module.Run(config); err != nil {
			fmt.Println(terminal.Paint(terminal.Warning, fmt.Sprintf("[!] %s on %s:%s: %v", service, target, port, err)))
		}
	}
	return nil
}

func validateServices(value string) error {
	services := splitNames(value)
	for _, name := range services {
		if _, ok := scanner.Module(name); !ok {
			return fmt.Errorf("unknown service %q; available: %s", name, availableModules())
		}
	}
	return nil
}

func parsePorts(value string) ([]string, error) {
	ports := make(map[int]struct{})
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if strings.Contains(item, "-") {
			parts := strings.SplitN(item, "-", 2)
			first, firstErr := strconv.Atoi(parts[0])
			last, lastErr := strconv.Atoi(parts[1])
			if firstErr != nil || lastErr != nil || first < 1 || last > 65535 || first > last {
				return nil, fmt.Errorf("invalid port range %q", item)
			}
			for port := first; port <= last; port++ {
				ports[port] = struct{}{}
			}
			continue
		}
		port, err := strconv.Atoi(item)
		if err != nil || port < 1 || port > 65535 {
			return nil, fmt.Errorf("invalid port %q", item)
		}
		ports[port] = struct{}{}
	}
	if len(ports) == 0 {
		return nil, fmt.Errorf("no ports specified")
	}
	result := make([]string, 0, len(ports))
	for port := range ports {
		result = append(result, strconv.Itoa(port))
	}
	sort.Slice(result, func(i, j int) bool {
		left, _ := strconv.Atoi(result[i])
		right, _ := strconv.Atoi(result[j])
		return left < right
	})
	return result, nil
}

func readLines(path string) ([]string, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read wordlist %q: %w", path, err)
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

func splitNames(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var names []string
	for _, part := range strings.Split(strings.ToLower(value), ",") {
		if name := strings.TrimSpace(part); name != "" {
			names = append(names, name)
		}
	}
	return names
}

func availableModules() string {
	modules := scanner.Modules()
	names := make([]string, 0, len(modules))
	for _, module := range modules {
		names = append(names, module.Name())
	}
	return strings.Join(names, ", ")
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func selectedProtocols(value string) []string {
	switch value {
	case "tcp":
		return []string{"tcp"}
	case "udp":
		return []string{"udp"}
	default:
		return []string{"tcp", "udp"}
	}
}

func moveTargetLast(args []string) []string {
	valueFlags := map[string]bool{
		"--ports": true, "-ports": true, "-p": true,
		"--modules": true, "-modules": true, "--services": true, "-services": true,
		"--ftp-userlist": true, "-ftp-userlist": true, "--ftp-passlist": true, "-ftp-passlist": true,
		"--userlist": true, "-userlist": true, "--passlist": true, "-passlist": true,
		"-Uw": true, "-Pw": true, "--timeout": true, "-timeout": true,
		"--workers": true, "-workers": true,
		"--protocol": true, "-protocol": true,
	}
	var flags []string
	target := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			if target != "" {
				return args
			}
			target = arg
			continue
		}
		flags = append(flags, arg)
		if valueFlags[arg] && i+1 < len(args) {
			flags = append(flags, args[i+1])
			i++
		}
	}
	if target == "" {
		return args
	}
	return append(flags, target)
}
