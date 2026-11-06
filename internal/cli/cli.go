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

	"marlow/pkg/scanner"
	"marlow/pkg/terminal"
)

type options struct {
	ports       string
	services    string
	allPorts    bool
	defaultMods bool
	userPath    string
	passPath    string
	timeout     time.Duration
	workers     int
}

func Run(args []string) error {
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
	if err := validateServices(opts.services, len(ports)); err != nil {
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
	return runModules(target, ports, openPorts, opts, users, passwords)
}

func parseArgs(args []string) (options, string, error) {
	fs := flag.NewFlagSet("marlow", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	opts := options{}
	defaultPorts := "21,22,25,53,80,110,143,443,445,3306,5432,6379,8080"
	fs.StringVar(&opts.ports, "p", defaultPorts, "Ports or ranges; for example 22,80,443 or 1-1024")
	fs.StringVar(&opts.ports, "ports", defaultPorts, "Alias for -p")
	fs.BoolVar(&opts.allPorts, "p-", false, "Scan all TCP ports (1-65535)")
	fs.BoolVar(&opts.allPorts, "all-ports", false, "Alias for -p-")
	fs.StringVar(&opts.services, "sV", "", "One module per port in order: ftp,http,smb,ssh")
	fs.StringVar(&opts.services, "services", "", "Alias for -sV")
	fs.StringVar(&opts.services, "modules", "", "Alias for -sV")
	fs.BoolVar(&opts.defaultMods, "s-", false, "Run a registered module matching each open port")
	fs.BoolVar(&opts.defaultMods, "auto-services", false, "Alias for -s-")
	fs.StringVar(&opts.userPath, "Uw", "", "FTP username wordlist")
	fs.StringVar(&opts.userPath, "userlist", "", "Alias for -Uw")
	fs.StringVar(&opts.passPath, "Pw", "", "FTP password wordlist")
	fs.StringVar(&opts.passPath, "passlist", "", "Alias for -Pw")
	fs.DurationVar(&opts.timeout, "timeout", 3*time.Second, "Per-port scan timeout, for example 3s")
	fs.IntVar(&opts.workers, "workers", 256, "Concurrent TCP scan workers")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: marlow [OPTIONS] <TARGET>")
		fmt.Fprintln(fs.Output(), "\nOptions:")
		fs.PrintDefaults()
		fmt.Fprintln(fs.Output(), "\nExamples:\n  marlow -p 21,80,443 -s- example.org\n  marlow -p- -s- 192.0.2.10\n  marlow --ports 22 --modules ssh 192.0.2.10")
	}

	if err := fs.Parse(args); err != nil {
		return options{}, "", err
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return options{}, "", fmt.Errorf("exactly one target is required")
	}
	if opts.timeout <= 0 || opts.workers < 1 {
		return options{}, "", fmt.Errorf("timeout and worker count must be positive")
	}
	if (opts.userPath == "") != (opts.passPath == "") {
		return options{}, "", fmt.Errorf("-Uw/--userlist and -Pw/--passlist must be supplied together")
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

func runModules(target string, scannedPorts, openPorts []string, opts options, users, passwords []string) error {
	serviceNames := splitNames(opts.services)
	for _, port := range openPorts {
		service := ""
		if len(serviceNames) > 0 {
			if index := indexOf(scannedPorts, port); index >= 0 && index < len(serviceNames) {
				service = serviceNames[index]
			}
		}
		if service == "" && opts.defaultMods {
			service = scanner.ServiceForPort(port)
		}
		if service == "" {
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

func validateServices(value string, portCount int) error {
	services := splitNames(value)
	if len(services) == 0 {
		return nil
	}
	if len(services) != portCount {
		return fmt.Errorf("-sV/--modules needs one module per specified port: got %d modules for %d ports", len(services), portCount)
	}
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

func indexOf(values []string, value string) int {
	for i, candidate := range values {
		if candidate == value {
			return i
		}
	}
	return -1
}
