package scanner

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

type PortResult struct {
	Port string
	Open bool
	Err  error
}

type Engine struct {
	Workers int
	Timeout time.Duration
}

func (e Engine) Scan(ctx context.Context, target string, ports []string) []PortResult {
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
					defer func() {
						if recovered := recover(); recovered != nil {
							result.Err = fmt.Errorf("scan worker recovered: %v", recovered)
						}
					}()
					address := net.JoinHostPort(target, port)
					conn, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "tcp", address)
					result.Open = err == nil
					result.Err = err
					if err == nil {
						_ = conn.Close()
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
