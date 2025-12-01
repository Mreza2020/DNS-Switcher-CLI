package platform_all

import "fmt"

// init: sets up a mock implementation of NetshExec
// Intercepts network commands and returns predictable mock output for testing
func init() {
	NetshExec = func(args ...string) (string, error) {
		return fmt.Sprintf("MOCK: %v", args), nil
	}
}
