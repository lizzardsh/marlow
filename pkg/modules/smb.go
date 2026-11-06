package modules

import (
	"fmt"
	"net"
	"time"

	"marlow/pkg/scanner"
	"marlow/pkg/terminal"
)

type smbModule struct{}

func (smbModule) Name() string           { return "smb" }
func (smbModule) DefaultPorts() []string { return []string{"445"} }

func init() { scanner.Register(smbModule{}) }

func (smbModule) Run(config scanner.ModuleConfig) error {
	address := net.JoinHostPort(config.TargetIP, config.Port)
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return err
	}
	_ = conn.Close()

	message := fmt.Sprintf("[*] SMB reachable at %s; share listing is not implemented in this build", address)
	fmt.Println(terminal.Paint(terminal.Info, message))
	return nil
}
