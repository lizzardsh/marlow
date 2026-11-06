package scanner

import (
	"sort"
	"strings"
	"sync"
)

type ModuleConfig struct {
	TargetIP   string
	Port       string
	UserLogins []string
	PassLogins []string
}

type ScannerModule interface {
	Name() string
	DefaultPorts() []string
	Run(config ModuleConfig) error
}

var registry = struct {
	sync.RWMutex
	modules map[string]ScannerModule
}{modules: make(map[string]ScannerModule)}

func Register(module ScannerModule) {
	if module == nil || strings.TrimSpace(module.Name()) == "" {
		panic("scanner: invalid module")
	}
	registry.Lock()
	defer registry.Unlock()
	key := strings.ToLower(module.Name())
	if _, exists := registry.modules[key]; exists {
		panic("scanner: duplicate module " + key)
	}
	registry.modules[key] = module
}

func Module(name string) (ScannerModule, bool) {
	registry.RLock()
	defer registry.RUnlock()
	m, ok := registry.modules[strings.ToLower(name)]
	return m, ok
}

func Modules() []ScannerModule {
	registry.RLock()
	defer registry.RUnlock()
	out := make([]ScannerModule, 0, len(registry.modules))
	for _, m := range registry.modules {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

func ServiceForPort(port string) string {
	for _, module := range Modules() {
		for _, defaultPort := range module.DefaultPorts() {
			if defaultPort == port {
				return module.Name()
			}
		}
	}
	return ""
}
