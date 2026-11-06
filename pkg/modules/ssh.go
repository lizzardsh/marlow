package modules

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"marlow/pkg/scanner"
	"marlow/pkg/terminal"
)

type sshModule struct{}

func (sshModule) Name() string           { return "ssh" }
func (sshModule) DefaultPorts() []string { return []string{"22"} }

func init() { scanner.Register(sshModule{}) }

func (sshModule) Run(config scanner.ModuleConfig) error {
	address := net.JoinHostPort(config.TargetIP, config.Port)
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	reader := bufio.NewReaderSize(conn, 256)
	for attempts := 0; attempts < 50; attempts++ {
		line, err := reader.ReadSlice('\n')
		if err != nil {
			return fmt.Errorf("read SSH identification from %s: %w", address, err)
		}
		identification := strings.TrimSpace(string(line))
		if strings.HasPrefix(identification, "SSH-") {
			fmt.Println(terminal.Paint(terminal.Success, fmt.Sprintf("[+] SSH %s version=%s", address, identification)))
			return nil
		}
	}
	return fmt.Errorf("SSH identification not found from %s", address)
}
