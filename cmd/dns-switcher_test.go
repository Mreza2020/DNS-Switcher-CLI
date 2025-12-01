package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Helper: Running a command and getting the output
func executeCommand(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	_, err := root.ExecuteC()
	return buf.String(), err
}

// Mock Profiles و RootCmd test
var mockProfiles = []struct {
	Name    string
	Servers []string
	Active  bool
}{
	{"google", []string{"8.8.8.8", "8.8.4.4"}, false},
	{"cloudflare", []string{"1.1.1.1", "1.0.0.1"}, true},
}

func resetMockProfiles() {
	mockProfiles = []struct {
		Name    string
		Servers []string
		Active  bool
	}{
		{"google", []string{"8.8.8.8", "8.8.4.4"}, false},
		{"cloudflare", []string{"1.1.1.1", "1.0.0.1"}, true},
	}
}

// Builds a mock Root Command including list/apply/status/delete
func buildTestRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{Use: "dns-switcher"}

	// list command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available DNS profiles",
		Run: func(cmd *cobra.Command, args []string) {
			if len(mockProfiles) == 0 {
				cmd.Println("No profiles found")
				return
			}
			cmd.Println("Available profiles:")
			for _, p := range mockProfiles {
				cmd.Printf(" - %s : %v\n", p.Name, p.Servers)
			}
		},
	}

	// apply command
	applyCmd := &cobra.Command{
		Use:   "apply [profile]",
		Short: "Apply a DNS profile",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			found := false
			for i, p := range mockProfiles {
				if p.Name == name {
					found = true
					mockProfiles[i].Active = true
					cmd.Printf("Applied profile '%s'\n", name)
				}
			}
			if !found {
				cmd.Printf("Profile '%s' not found\n", name)
			}
		},
	}

	// status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show current DNS settings",
		Run: func(cmd *cobra.Command, args []string) {
			activeProfiles := []string{}
			for _, p := range mockProfiles {
				if p.Active {
					activeProfiles = append(activeProfiles, p.Name)
				}
			}
			cmd.Printf("Active DNS profiles: %v\n", activeProfiles)
		},
	}

	// delete-profile command (چند پروفایل)
	deleteProfileCmd := &cobra.Command{
		Use:   "delete-profile [name1 name2 ...]",
		Short: "Delete one or more DNS profiles by name",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			yes, _ := cmd.Flags().GetBool("yes")
			force, _ := cmd.Flags().GetBool("force")
			jsonOut, _ := cmd.Flags().GetBool("json")

			for _, name := range args {
				found := false
				for i, p := range mockProfiles {
					if p.Name == name {
						found = true
						if p.Active && !force {
							cmd.Printf("Profile '%s' is currently active. Use --force to delete.\n", name)
							continue
						}
						if !yes {
							cmd.Printf("Are you sure you want to delete profile '%s'? [y/N]: y\n", name)
						}
						mockProfiles = append(mockProfiles[:i], mockProfiles[i+1:]...)
						if jsonOut {
							cmd.Printf("{\"deleted\":true,\"name\":\"%s\"}\n", name)
						} else {
							cmd.Printf("Profile '%s' deleted\n", name)
						}
						break
					}
				}
				if !found {
					cmd.Printf("Profile '%s' not found\n", name)
				}
			}
		},
	}
	deleteProfileCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	deleteProfileCmd.Flags().BoolP("force", "f", false, "Delete even if profile is active")
	deleteProfileCmd.Flags().BoolP("json", "j", false, "Output result in JSON")

	rootCmd.AddCommand(listCmd, applyCmd, statusCmd, deleteProfileCmd)
	return rootCmd
}

// ---------------- TESTS ----------------

// Test list command
func TestListCommand(t *testing.T) {
	resetMockProfiles()
	root := buildTestRootCmd()
	out, _ := executeCommand(root, "list")
	if !strings.Contains(out, "google") || !strings.Contains(out, "cloudflare") {
		t.Errorf("expected profiles in list, got: %s", out)
	}
}

// Test apply command
func TestApplyCommand(t *testing.T) {
	resetMockProfiles()
	root := buildTestRootCmd()
	out, _ := executeCommand(root, "apply", "google")
	if !strings.Contains(out, "Applied profile 'google'") {
		t.Errorf("expected apply success, got: %s", out)
	}
}

// Test status command
func TestStatusCommand(t *testing.T) {
	resetMockProfiles()
	root := buildTestRootCmd()
	_, _ = executeCommand(root, "apply", "google")
	out, _ := executeCommand(root, "status")
	if !strings.Contains(out, "google") {
		t.Errorf("expected active profile in status, got: %s", out)
	}
}

// Test delete single profile
func TestDeleteProfileCommand(t *testing.T) {
	resetMockProfiles()
	root := buildTestRootCmd()
	out, _ := executeCommand(root, "delete-profile", "google", "--yes")
	if !strings.Contains(out, "Profile 'google' deleted") {
		t.Errorf("expected delete success, got: %s", out)
	}
}

// Test delete multiple profiles
func TestDeleteMultipleProfiles(t *testing.T) {
	resetMockProfiles()
	root := buildTestRootCmd()
	out, _ := executeCommand(root, "delete-profile", "google", "cloudflare", "--yes", "--force")
	if !strings.Contains(out, "Profile 'google' deleted") || !strings.Contains(out, "Profile 'cloudflare' deleted") {
		t.Errorf("expected multiple deletes, got: %s", out)
	}
}

// Test delete profile with JSON output
func TestDeleteProfileJSON(t *testing.T) {
	resetMockProfiles()
	root := buildTestRootCmd()
	out, _ := executeCommand(root, "delete-profile", "google", "--yes", "--json")
	if !strings.Contains(out, "\"deleted\":true") || !strings.Contains(out, "\"name\":\"google\"") {
		t.Errorf("expected JSON delete output, got: %s", out)
	}
}

// Test delete non-existing profile
func TestDeleteProfileNotFound(t *testing.T) {
	resetMockProfiles()
	root := buildTestRootCmd()
	out, _ := executeCommand(root, "delete-profile", "nonexistent", "--yes")
	if !strings.Contains(out, "Profile 'nonexistent' not found") {
		t.Errorf("expected not found message, got: %s", out)
	}
}
