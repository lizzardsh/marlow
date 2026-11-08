// Package fingerprint performs small, protocol-aware probes against open TCP
// ports to collect service banners and basic version information.
package fingerprint

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	Service string
	Version string
}

// Detect uses passive greeting reads for banner-first protocols and HTTP HEAD
// requests on common web ports. It does not authenticate or send exploit data.
func Detect(ctx context.Context, target, port string, timeout time.Duration) Result {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	if isHTTPPort(port) {
		if result, ok := detectHTTP(ctx, target, port, timeout); ok {
			return result
		}
	}
	service := bannerService(port)
	if service == "" {
		return Result{Service: "unknown", Version: "unidentified"}
	}

	address := net.JoinHostPort(target, port)
	conn, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "tcp", address)
	if err != nil {
		return Result{Service: service, Version: "banner unavailable"}
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	line, err := bufio.NewReaderSize(conn, 1024).ReadString('\n')
	if err != nil && len(line) == 0 {
		return Result{Service: service, Version: "banner unavailable"}
	}
	line = strings.TrimSpace(line)
	if line == "" {
		line = "banner unavailable"
	}
	return Result{Service: service, Version: line}
}

func detectHTTP(ctx context.Context, target, port string, timeout time.Duration) (Result, bool) {
	secure := isHTTPSPort(port)
	scheme := "http"
	if secure {
		scheme = "https"
	}
	url := scheme + "://" + net.JoinHostPort(target, port) + "/"
	client := &http.Client{Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	request, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return Result{}, false
	}
	response, err := client.Do(request)
	if err != nil {
		return Result{}, false
	}
	defer response.Body.Close()
	server := strings.TrimSpace(response.Header.Get("Server"))
	version := "HTTP " + strconv.Itoa(response.StatusCode)
	if server != "" {
		version += "; Server: " + server
	}
	if secure && response.TLS != nil {
		version += "; TLS: " + tlsName(response.TLS.Version)
	}
	return Result{Service: scheme, Version: version}, true
}

func isHTTPPort(port string) bool {
	switch port {
	case "80", "443", "8000", "8008", "8080", "8081", "8443", "8888", "9000", "9443":
		return true
	default:
		return false
	}
}

func isHTTPSPort(port string) bool {
	switch port {
	case "443", "8443", "9443":
		return true
	default:
		return false
	}
}

func bannerService(port string) string {
	switch port {
	case "21":
		return "ftp"
	case "22":
		return "ssh"
	case "25", "587":
		return "smtp"
	case "110":
		return "pop3"
	case "143":
		return "imap"
	default:
		return ""
	}
}

func tlsName(version uint16) string {
	switch version {
	case 0x0301:
		return "TLS 1.0"
	case 0x0302:
		return "TLS 1.1"
	case 0x0303:
		return "TLS 1.2"
	case 0x0304:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("0x%04x", version)
	}
}
