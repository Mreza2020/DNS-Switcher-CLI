package platform_all

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Mreza2020/DNS-Switcher/internal/config"
)

type ApplyResult struct {
	Ok      bool
	Message string
}

var NetshExec = runNetsh

// runNetsh executes a Windows `netsh` command with given arguments
// and returns combined stdout/stderr output as string.
// It is a helper used by other networking functions.
func runNetsh(args ...string) (string, error) {
	cmd := exec.Command("netsh", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

// getActiveInterface detects and returns the name of the currently
// active (Connected) network interface using `netsh interface show interface`.
// Returns error if no connected interface is found.
func getActiveInterface() (string, error) {
	out, err := runNetsh("interface", "show", "interface")
	if err != nil {
		return "", fmt.Errorf("failed to run netsh: %v, output: %s", err, out)
	}

	lines := strings.Split(out, "\n")
	for _, l := range lines {
		if strings.Contains(l, "Connected") {
			fields := strings.Fields(l)
			if len(fields) >= 4 {
				iface := strings.Join(fields[3:], " ")
				return iface, nil
			}
		}
	}
	return "", fmt.Errorf("no active network interface found")
}

// ApplyProfile applies static DNS settings (primary + optional secondary servers)
// to the currently active network interface using netsh.
// Returns ApplyResult with success status and descriptive message.
func ApplyProfile(p config.Profile) (ApplyResult, error) {
	if runtime.GOOS != "windows" {
		return ApplyResult{Ok: false}, fmt.Errorf("only supported on Windows")
	}

	iface := p.Interface
	if iface == "" {
		return ApplyResult{Ok: false}, fmt.Errorf("no interface specified")
	}

	if len(p.Servers) == 0 {
		return ApplyResult{Ok: false}, fmt.Errorf("no DNS servers provided")
	}

	// apply primary DNS server
	out, err := NetshExec("interface", "ip", "set", "dns",
		fmt.Sprintf("name=%s", iface),
		"source=static",
		fmt.Sprintf("address=%s", p.Servers[0]))
	if err != nil {
		return ApplyResult{Ok: false}, fmt.Errorf("netsh error: %v, output: %s", err, out)
	}

	// apply secondary DNS servers
	for i := 1; i < len(p.Servers); i++ {
		out1, err1 := NetshExec("interface", "ip", "add", "dns",
			fmt.Sprintf("name=%s", iface),
			fmt.Sprintf("addr=%s", p.Servers[i]),
			fmt.Sprintf("index=%d", i+1))
		if err1 != nil {
			return ApplyResult{Ok: false}, fmt.Errorf("netsh error: %v, output: %s", err1, out1)
		}
	}

	return ApplyResult{Ok: true, Message: fmt.Sprintf("DNS applied to %s: %v", iface, p.Servers)}, nil
}

// Rollback restores DNS mode back to automatic DHCP on the active interface.
// Used when reverting applied DNS profiles.
func Rollback(iface string) (ApplyResult, error) {
	if runtime.GOOS != "windows" {
		return ApplyResult{Ok: false}, fmt.Errorf("only supported on Windows")
	}

	if iface == "" {
		var err error
		iface, err = getActiveInterface()
		if err != nil {
			return ApplyResult{Ok: false}, err
		}
	}

	out, err := runNetsh("interface", "ip", "set", "dns", fmt.Sprintf("name=%s", iface), "source=dhcp")
	if err != nil {
		return ApplyResult{Ok: false}, fmt.Errorf("netsh error: %v, output: %s", err, out)
	}
	return ApplyResult{Ok: true, Message: fmt.Sprintf("Rollback successful – DNS restored to automatic (DHCP) on interface %s", iface)}, nil
}

// GetCurrentDNS reads and returns the currently configured DNS servers
// (static or enumerated) for the active network interface.
// Output is parsed from `netsh interface ip show dns name=<iface>`.
func GetCurrentDNS(iface string) ([]string, error) {
	out, err := NetshExec("interface", "ip", "show", "dns",
		fmt.Sprintf("name=%s", iface))
	if err != nil {
		return nil, err
	}

	var dnsList []string
	var capture bool
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "Statically Configured DNS Servers") ||
			strings.HasPrefix(l, "DNS Servers") {
			parts := strings.Split(l, ":")
			if len(parts) == 2 {
				dnsList = append(dnsList, strings.TrimSpace(parts[1]))
			}
			capture = true
			continue
		}
		if capture && l != "" {
			dnsList = append(dnsList, l)
		}
	}
	return dnsList, nil
}

func GetNetworkInterfaces() ([]string, error) {
	out, err := NetshExec("interface", "show", "interface")
	if err != nil {
		return nil, fmt.Errorf("failed to run netsh: %v, output: %s", err, out)
	}

	var names []string
	lines := strings.Split(out, "\n")

	for _, l := range lines {
		l = strings.TrimSpace(l)

		if strings.Contains(l, "Connected") {
			fields := strings.Fields(l)
			if len(fields) >= 4 {
				iface := strings.Join(fields[3:], " ")
				names = append(names, iface)
			}
		}
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no active interfaces found")
	}

	return names, nil
}
