package modules

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"

	"marlow/pkg/scanner"
	"marlow/pkg/terminal"
)

type ftpModule struct{}

func (ftpModule) Name() string           { return "ftp" }
func (ftpModule) DefaultPorts() []string { return []string{"21"} }

func init() { scanner.Register(ftpModule{}) }

func (ftpModule) Run(config scanner.ModuleConfig) error {
	address := net.JoinHostPort(config.TargetIP, config.Port)
	if allowed, listing, err := ftpLogin(address, "anonymous", "anonymous@"); err == nil && allowed {
		fmt.Println(terminal.Paint(terminal.Success, "[+] SUCCESS: Anonymous FTP login allowed on "+address))
		for _, entry := range listing {
			fmt.Printf("    %s\n", entry)
		}
		return nil
	}
	if len(config.UserLogins) == 0 || len(config.PassLogins) == 0 {
		fmt.Println(terminal.Paint(terminal.Info, "[*] Anonymous FTP login denied on "+address))
		return nil
	}

	type credential struct{ username, password string }
	jobs := make(chan credential)
	var workers sync.WaitGroup
	var found sync.Once
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for item := range jobs {
				checkCredential := func() {
					allowed, _, err := ftpLogin(address, item.username, item.password)
					if err == nil && allowed {
						found.Do(func() {
							fmt.Println(terminal.Paint(terminal.Success, fmt.Sprintf("[+] SUCCESS: FTP credentials accepted at %s (username: %s)", address, item.username)))
						})
					}
				}
				func() {
					defer func() {
						if recovered := recover(); recovered != nil {
							fmt.Printf("[!] FTP worker recovered: %v\n", recovered)
						}
					}()
					checkCredential()
				}()
			}
		}()
	}
	for _, username := range config.UserLogins {
		for _, password := range config.PassLogins {
			jobs <- credential{username: username, password: password}
		}
	}
	close(jobs)
	workers.Wait()
	fmt.Println(terminal.Paint(terminal.Warning, "[!] Finished supplied FTP credential checks for "+address))
	return nil
}

func ftpLogin(address, username, password string) (bool, []string, error) {
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return false, nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	server := textproto.NewConn(conn)
	if _, _, err := server.ReadResponse(220); err != nil {
		return false, nil, err
	}
	if _, err := server.Cmd("USER %s", username); err != nil {
		return false, nil, err
	}
	code, _, err := server.ReadResponse(0)
	if err != nil {
		return false, nil, err
	}
	if code != 230 && code != 331 {
		return false, nil, nil
	}
	if code == 331 {
		if _, err := server.Cmd("PASS %s", password); err != nil {
			return false, nil, err
		}
		code, _, err = server.ReadResponse(0)
		if err != nil || (code != 230 && code != 202) {
			return false, nil, err
		}
	}

	if _, err := server.Cmd("TYPE I"); err == nil {
		_, _, _ = server.ReadResponse(200)
	}
	_, err = server.Cmd("EPSV")
	if err != nil {
		return true, nil, nil
	}
	_, message, err := server.ReadResponse(229)
	if err != nil {
		return true, nil, nil
	}
	port, ok := passivePort(message)
	if !ok {
		return true, nil, nil
	}
	host, _, _ := net.SplitHostPort(address)
	data, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 3*time.Second)
	if err != nil {
		return true, nil, nil
	}
	defer data.Close()
	_ = data.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := server.Cmd("LIST"); err != nil {
		return true, nil, nil
	}
	_, _, _ = server.ReadResponse(150)
	var listing []string
	scanner := bufio.NewScanner(data)
	for scanner.Scan() && len(listing) < 5 {
		listing = append(listing, scanner.Text())
	}
	_, _, _ = server.ReadResponse(226)
	return true, listing, nil
}

func passivePort(message string) (string, bool) {
	start := strings.LastIndex(message, "(")
	end := strings.LastIndex(message, ")")
	if start < 0 || end <= start {
		return "", false
	}
	fields := strings.Split(message[start+1:end], "|")
	if len(fields) < 4 {
		return "", false
	}
	return fields[len(fields)-2], true
}
