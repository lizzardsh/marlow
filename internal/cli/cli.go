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
	ports, err := parsePorts(portSpec)
	if err != nil {
		return err
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

	fmt.Println(terminal.Paint(terminal.Info, fmt.Sprintf("[*] Scanning %s (%d TCP ports; timeout %s)", target, len(ports), opts.timeout)))
	results := (scanner.Engine{Workers: opts.workers, Timeout: opts.timeout}).Scan(context.Background(), target, ports)
	openPorts := make([]string, 0)
	for _, result := range results {
		if !result.Open {
			continue
		}
		openPorts = append(openPorts, result.Port)
		fmt.Println(terminal.Paint(terminal.Success, "[+] OPEN: "+result.Port+"/tcp"))
	}
	if len(openPorts) == 0 {
		fmt.Println(terminal.Paint(terminal.Error, "[-] No open TCP ports found"))
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
	opts := options{}
	var showHelp bool
	defaultPorts := "21,22,25,53,80,110,143,443,445,3306,5432,6379,8080"
	fs.StringVar(&opts.ports, "ports", defaultPorts, "TCP ports or ranges, for example 22,80,443 or 1-1024")
	fs.BoolVar(&opts.allPorts, "all-ports", false, "Scan TCP ports 1-65535")
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
	fs.StringVar(&opts.ports, "p", defaultPorts, "")
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
		fmt.Fprintln(fs.Output(), "\nOptions:\n  --ports <list>         TCP ports or ranges (default: common ports)\n  --all-ports            Scan TCP ports 1-65535\n  --versions             Collect service banners and basic version details\n  --service-checks       Run built-in checks for recognized services\n  --modules <list>       Limit checks to ftp,http,smb,ssh\n  --ftp-userlist <path>  FTP username wordlist\n  --ftp-passlist <path>  FTP password wordlist\n  --timeout <duration>   Connection timeout (default: 3s)\n  --workers <count>      Concurrent port scan workers (default: 256)\n  -h, --help             Show this help")
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

func moveTargetLast(args []string) []string {
	valueFlags := map[string]bool{
		"--ports": true, "-ports": true, "-p": true,
		"--modules": true, "-modules": true, "--services": true, "-services": true,
		"--ftp-userlist": true, "-ftp-userlist": true, "--ftp-passlist": true, "-ftp-passlist": true,
		"--userlist": true, "-userlist": true, "--passlist": true, "-passlist": true,
		"-Uw": true, "-Pw": true, "--timeout": true, "-timeout": true,
		"--workers": true, "-workers": true,
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
