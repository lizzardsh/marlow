package scanner

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PortResult struct {
	Port     string
	Protocol string
	State    string
	Open     bool
	Err      error
}

type Engine struct {
	Workers int
	Timeout time.Duration
}

func (e Engine) Scan(ctx context.Context, target string, ports []string) []PortResult {
	return e.ScanProtocol(ctx, target, ports, "tcp")
}

// ScanProtocol probes the supplied ports using TCP connect or UDP datagrams.
// UDP silence is reported as open|filtered because an open UDP service may not
// reply to an empty probe.
func (e Engine) ScanProtocol(ctx context.Context, target string, ports []string, protocol string) []PortResult {
	workers := e.Workers
	if workers < 1 {
		workers = 128
	}
	if workers > len(ports) {
		workers = len(ports)
	}
	if workers == 0 {
		return nil
	}
	timeout := e.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	jobs := make(chan string)
	results := make(chan PortResult)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range jobs {
				result := func() (result PortResult) {
					result.Port = port
					result.Protocol = protocol
					defer func() {
						if recovered := recover(); recovered != nil {
							result.Err = fmt.Errorf("scan worker recovered: %v", recovered)
						}
					}()
					address := net.JoinHostPort(target, port)
					if protocol == "udp" {
						result.State, result.Open, result.Err = scanUDP(ctx, address, timeout)
					} else {
						conn, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "tcp", address)
						result.Open = err == nil
						result.Err = err
						if err == nil {
							result.State = "OPEN"
							_ = conn.Close()
						} else {
							result.State = "CLOSED|FILTERED"
						}
					}
					return result
				}()
				select {
				case results <- result:
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, port := range ports {
			select {
			case jobs <- port:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		wg.Wait()
		close(results)
	}()
	out := make([]PortResult, 0, len(ports))
	for result := range results {
		out = append(out, result)
	}
	sort.Slice(out, func(i, j int) bool {
		left, _ := strconv.Atoi(out[i].Port)
		right, _ := strconv.Atoi(out[j].Port)
		return left < right
	})
	return out
}

func scanUDP(ctx context.Context, address string, timeout time.Duration) (string, bool, error) {
	conn, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "udp", address)
	if err != nil {
		return "CLOSED|FILTERED", false, err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return "OPEN|FILTERED", false, err
	}
	if _, err := conn.Write([]byte{0}); err != nil {
		return "CLOSED|FILTERED", false, err
	}
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err == nil {
		return "OPEN", true, nil
	}
	if opErr, ok := err.(*net.OpError); ok && strings.Contains(strings.ToLower(opErr.Err.Error()), "refused") {
		return "CLOSED", false, err
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return "OPEN|FILTERED", false, nil
	}
	return "CLOSED|FILTERED", false, err
}
