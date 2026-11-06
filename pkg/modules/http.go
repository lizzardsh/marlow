package modules

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"marlow/pkg/scanner"
	"marlow/pkg/terminal"
)

type httpModule struct{}

func (httpModule) Name() string           { return "http" }
func (httpModule) DefaultPorts() []string { return []string{"80", "443", "8080"} }

func init() { scanner.Register(httpModule{}) }

func (httpModule) Run(config scanner.ModuleConfig) error {
	scheme := "http"
	if config.Port == "443" {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(config.TargetIP, config.Port))
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	response, err := client.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	message := fmt.Sprintf("[+] HTTP %s status=%d server=%q", url, response.StatusCode, response.Header.Get("Server"))
	fmt.Println(terminal.Paint(terminal.Success, message))
	return nil
}
