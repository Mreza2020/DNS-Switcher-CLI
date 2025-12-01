package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Mreza2020/DNS-Switcher/internal/config"
	platformall "github.com/Mreza2020/DNS-Switcher/internal/platform-all"
	"github.com/Mreza2020/DNS-Switcher/internal/resolver"
	"github.com/spf13/cobra"
)

var (
	DomainTesting string
)

func init() {
	DomainTesting = os.Getenv("DomainTesting")

	if DomainTesting == "" {
		log.Fatal("Environment variable DomainTesting is not set. Please set it before running.")
		return
	}
}

func normalizeDNS(list []string) []string {
	var ips []string
	for _, entry := range list {
		trimmed := strings.TrimSpace(entry)
		if net.ParseIP(trimmed) != nil {
			ips = append(ips, trimmed)
		}
	}
	return ips
}

// equalDNS
func equalDNS(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	amap := make(map[string]bool)
	for _, ip := range a {
		amap[ip] = true
	}
	for _, ip := range b {
		if !amap[ip] {
			return false
		}
	}
	return true
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "dns-switcher",
		Short: "DNS switcher",
		Long:  "DNS Switcher lets you manage DNS profiles, test latency, apply settings, and rollback safely.",
	}

	// List Command
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List available DNS profiles",
		Run: func(cmd *cobra.Command, args []string) {
			profiles := config.LoadProfilesDns()
			if len(profiles) == 0 {
				fmt.Println("No profiles found")
				return
			}
			fmt.Println("Available profiles:")
			verbose, _ := cmd.Flags().GetBool("verbose")

			for _, p := range profiles {
				if verbose {
					fmt.Printf(" - %s : %v\n", p.Name, p.Servers)
				} else {
					fmt.Printf(" - %s\n", p.Name)
				}
			}
		},
	}
	listCmd.Flags().BoolP("verbose", "v", false, "Show servers in list output")

	// Test Command
	var testCmd = &cobra.Command{
		Use:   "test [profile|server]",
		Short: "Test latency for a profile or server",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// خواندن فلگ repeat
			repeat, _ := cmd.Flags().GetInt("repeat")
			if repeat <= 0 {
				repeat = 1
			}

			target := args[0]
			profiles := config.LoadProfilesDns()

			if p, ok := config.FindProfile(profiles, target); ok {
				fmt.Printf("Testing profile '%s'\n", p.Name)
				for _, server := range p.Servers {
					var sum time.Duration
					var success int

					for i := 0; i < repeat; i++ {
						r := resolver.MeasureRTT(server, DomainTesting, 2*time.Second)
						if r.Error != nil {
							fmt.Printf("%s -> error: %v\n", server, r.Error)
							continue
						}
						fmt.Printf("%s -> RTT[%d]: %v\n", server, i+1, r.RTT)
						sum += r.RTT
						success++
					}

					if success > 0 {
						avg := sum / time.Duration(success)
						fmt.Printf("%s -> average RTT: %v\n", server, avg)
					}
				}
			} else {
				fmt.Printf("Testing server '%s'\n", target)
				var sum time.Duration
				var success int

				for i := 0; i < repeat; i++ {
					r := resolver.MeasureRTT(target, DomainTesting, 2*time.Second)
					if r.Error != nil {
						fmt.Printf("%s -> error: %v\n", target, r.Error)
						continue
					}
					fmt.Printf("%s -> RTT[%d]: %v\n", target, i+1, r.RTT)
					sum += r.RTT
					success++
				}

				if success > 0 {
					avg := sum / time.Duration(success)
					fmt.Printf("%s -> average RTT: %v\n", target, avg)
				}
			}
		},
	}
	testCmd.Flags().IntP("repeat", "r", 1, "Number of times to repeat RTT test")

	// Apply Command
	var applyCmd = &cobra.Command{
		Use:   "apply [profile]",
		Short: "Apply a DNS profile",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			force, _ := cmd.Flags().GetBool("force")
			iface, _ := cmd.Flags().GetString("iface")

			profileName := args[0]
			profiles := config.LoadProfilesDns()
			p, ok := config.FindProfile(profiles, profileName)
			if !ok {
				fmt.Printf("Profile '%s' not found\n", profileName)
				return
			}

			if iface == "" {
				interfaces, err := platformall.GetNetworkInterfaces()
				if err != nil || len(interfaces) == 0 {
					fmt.Println("No network interfaces detected.")
					return
				}

				fmt.Println("Available network interfaces:")
				for i, name := range interfaces {
					fmt.Printf(" [%d] %s\n", i+1, name)
				}

				fmt.Print("Select interface number: ")
				var choice int
				fmt.Scanln(&choice)

				if choice < 1 || choice > len(interfaces) {
					fmt.Println("Invalid choice")
					return
				}

				iface = interfaces[choice-1]
			}

			p.Interface = iface

			if !force {
				current, err := platformall.GetCurrentDNS(iface)
				if err == nil && len(current) > 0 {
					normalized := normalizeDNS(current)
					if equalDNS(normalized, p.Servers) {
						fmt.Printf("Profile '%s' is already active on interface '%s'. Use -f to force reapply.\n", profileName, iface)
						return
					}
				}
			}

			res, err := platformall.ApplyProfile(*p)
			if err != nil {
				fmt.Printf("Error applying profile: %v\n", err)
			} else {
				fmt.Println(res.Message)
			}
		},
	}
	applyCmd.Flags().BoolP("force", "f", false, "Force apply even if already active")
	applyCmd.Flags().StringP("iface", "i", "", "Specify network interface")

	// Status Command
	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show current DNS settings",
		Run: func(cmd *cobra.Command, args []string) {

			jsonOutput, _ := cmd.Flags().GetBool("json")
			iface, _ := cmd.Flags().GetString("iface")

			if iface == "" {
				interfaces, err := platformall.GetNetworkInterfaces()
				if err != nil || len(interfaces) == 0 {
					fmt.Println("No network interfaces detected.")
					return
				}

				fmt.Println("Available network interfaces:")
				for i, name := range interfaces {
					fmt.Printf(" [%d] %s\n", i+1, name)
				}

				fmt.Print("Select interface number: ")
				var choice int
				fmt.Scanln(&choice)

				if choice < 1 || choice > len(interfaces) {
					fmt.Println("Invalid choice")
					return
				}

				iface = interfaces[choice-1]
			}

			dnsList, err := platformall.GetCurrentDNS(iface)
			if err != nil {
				fmt.Printf("Error getting current DNS: %v\n", err)
				return
			}

			normalized := normalizeDNS(dnsList)

			if jsonOutput {
				data, _ := json.MarshalIndent(normalized, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Println("Current DNS servers:")
				for _, d := range normalized {
					fmt.Println(" -", d)
				}
			}
		},
	}

	statusCmd.Flags().StringP("iface", "i", "", "Select network interface")
	statusCmd.Flags().BoolP("json", "j", false, "Output status in JSON format")

	// Rollback Command
	var rollbackCmd = &cobra.Command{
		Use:   "rollback",
		Short: "Rollback to previous DNS settings",
		Run: func(cmd *cobra.Command, args []string) {
			quiet, _ := cmd.Flags().GetBool("quiet")
			iface, _ := cmd.Flags().GetString("iface")

			if iface == "" {
				interfaces, err := platformall.GetNetworkInterfaces()
				if err != nil || len(interfaces) == 0 {
					fmt.Println("No network interfaces detected.")
					return
				}

				fmt.Println("Available network interfaces:")
				for i, name := range interfaces {
					fmt.Printf(" [%d] %s\n", i+1, name)
				}

				fmt.Print("Select interface number: ")
				var choice int
				fmt.Scanln(&choice)

				if choice < 1 || choice > len(interfaces) {
					fmt.Println("Invalid choice")
					return
				}

				iface = interfaces[choice-1]
			}

			res, err := platformall.Rollback(iface)
			if err != nil {
				fmt.Printf("Rollback error: %v\n", err)
			} else {
				if !quiet {
					fmt.Println(res.Message)
				}
			}
		},
	}
	rollbackCmd.Flags().BoolP("quiet", "q", false, "Suppress success message")
	rollbackCmd.Flags().StringP("iface", "i", "", "Select network interface")

	// Add-profile Command
	var addProfileCmd = &cobra.Command{
		Use:   "add-profile",
		Short: "Add a new DNS profile",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			servers, _ := cmd.Flags().GetString("servers")
			if name == "" || servers == "" {
				fmt.Println("Please provide --name and --servers")
				return
			}
			serverList := strings.Split(servers, ",")
			err := config.AddProfile(name, serverList)
			if err != nil {
				fmt.Printf("Error adding profile: %v\n", err)
			} else {
				fmt.Printf("Profile '%s' added: %v\n", name, serverList)
			}
		},
	}
	addProfileCmd.Flags().StringP("name", "n", "", "Profile name")
	addProfileCmd.Flags().StringP("servers", "s", "", "Comma-separated DNS servers")

	// Auto Command
	var autoCmd = &cobra.Command{
		Use:   "auto",
		Short: "Auto-select the fastest DNS profile",
		Run: func(cmd *cobra.Command, args []string) {
			repeat, _ := cmd.Flags().GetInt("repeat")
			apply, _ := cmd.Flags().GetBool("apply")
			iface, _ := cmd.Flags().GetString("iface")

			if repeat <= 0 {
				repeat = 5
			}

			if iface == "" {
				interfaces, err := platformall.GetNetworkInterfaces()
				if err != nil || len(interfaces) == 0 {
					fmt.Println("No network interfaces detected.")
					return
				}

				fmt.Println("Available network interfaces:")
				for i, name := range interfaces {
					fmt.Printf(" [%d] %s\n", i+1, name)
				}

				fmt.Print("Select interface number: ")
				var choice int
				fmt.Scanln(&choice)

				if choice < 1 || choice > len(interfaces) {
					fmt.Println("Invalid choice")
					return
				}

				iface = interfaces[choice-1]
			}

			profiles := config.LoadProfilesDns()
			if len(profiles) == 0 {
				fmt.Println("No profiles found")
				return
			}

			var bestProfile *config.Profile
			bestAvg := time.Duration(1<<63 - 1)

			for _, p := range profiles {
				fmt.Printf("Testing profile '%s'\n", p.Name)

				var total time.Duration
				var count int

				for _, server := range p.Servers {
					var sum time.Duration
					var success int

					for i := 0; i < repeat; i++ {
						r := resolver.MeasureRTT(server, DomainTesting, 2*time.Second)
						if r.Error != nil {
							fmt.Printf("%s -> error: %v\n", server, r.Error)
							continue
						}
						fmt.Printf("%s -> RTT[%d]: %v\n", server, i+1, r.RTT)
						sum += r.RTT
						success++
					}

					if success > 0 {
						avg := sum / time.Duration(success)
						fmt.Printf("%s -> average RTT: %v\n", server, avg)
						total += avg
						count++
					}
				}

				if count > 0 {
					profileAvg := total / time.Duration(count)
					fmt.Printf("Profile '%s' average RTT: %v\n\n", p.Name, profileAvg)

					if profileAvg < bestAvg {
						bestAvg = profileAvg
						bestProfile = &p
					}
				}
			}

			if bestProfile == nil {
				fmt.Println("No valid servers found")
				return
			}

			if apply {

				bestProfile.Interface = iface
				res, err := platformall.ApplyProfile(*bestProfile)
				if err != nil {
					fmt.Printf("Error applying profile: %v\n", err)
					return
				}
				fmt.Println(res.Message)
				fmt.Printf("Applied fastest profile '%s' with average RTT %v on interface '%s'\n", bestProfile.Name, bestAvg, iface)
			} else {
				fmt.Printf("Fastest profile is '%s' with average RTT %v (not applied)\n", bestProfile.Name, bestAvg)
			}
		},
	}

	autoCmd.Flags().IntP("repeat", "r", 5, "Number of times to test each server")
	autoCmd.Flags().BoolP("apply", "a", false, "Apply fastest profile automatically")
	autoCmd.Flags().StringP("iface", "i", "", "Select network interface")

	// Delete-profile Command
	var deleteProfileCmd = &cobra.Command{
		Use:   "delete-profile [name1 name2 ...]",
		Short: "Delete one or more DNS profiles by name",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			yes, _ := cmd.Flags().GetBool("yes")
			force, _ := cmd.Flags().GetBool("force")
			jsonOut, _ := cmd.Flags().GetBool("json")
			quiet, _ := cmd.Flags().GetBool("quiet")
			iface, _ := cmd.Flags().GetString("iface")

			if iface == "" {
				interfaces, err := platformall.GetNetworkInterfaces()
				if err != nil || len(interfaces) == 0 {
					fmt.Println("No network interfaces detected.")
					return
				}

				fmt.Println("Available network interfaces:")
				for i, name := range interfaces {
					fmt.Printf(" [%d] %s\n", i+1, name)
				}

				fmt.Print("Select interface number: ")
				var choice int
				fmt.Scanln(&choice)

				if choice < 1 || choice > len(interfaces) {
					fmt.Println("Invalid choice")
					return
				}

				iface = interfaces[choice-1]
			}

			for _, name := range args {

				profiles := config.LoadProfilesDns()
				p, ok := config.FindProfile(profiles, name)
				if !ok {
					fmt.Printf("Profile '%s' not found\n", name)
					continue
				}

				current, err := platformall.GetCurrentDNS(iface)
				normalized := normalizeDNS(current)
				isActive := err == nil && len(normalized) > 0 && equalDNS(normalized, p.Servers)

				if isActive && !force {
					fmt.Printf("Profile '%s' is currently active. Use --force to delete.\n", name)
					continue
				}

				if !yes {
					fmt.Printf("Are you sure you want to delete profile '%s'? [y/N]: ", name)
					var reply string
					fmt.Scanln(&reply)
					reply = strings.ToLower(strings.TrimSpace(reply))
					if reply != "y" && reply != "yes" {
						fmt.Printf("Aborted deletion of '%s'.\n", name)
						continue
					}
				}

				if err := config.DeleteProfile(name); err != nil {
					fmt.Printf("Error deleting profile '%s': %v\n", name, err)
					continue
				}

				if jsonOut {
					payload := map[string]interface{}{
						"deleted": true,
						"name":    name,
						"forced":  force,
						"active":  isActive,
					}
					b, _ := json.MarshalIndent(payload, "", "  ")
					fmt.Println(string(b))
					continue
				}

				if !quiet {
					fmt.Printf("Profile '%s' deleted\n", name)
				}
			}
		},
	}

	deleteProfileCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	deleteProfileCmd.Flags().BoolP("force", "f", false, "Delete even if profile is active")
	deleteProfileCmd.Flags().BoolP("json", "j", false, "Output result in JSON")
	deleteProfileCmd.Flags().BoolP("quiet", "q", false, "Suppress success message")
	deleteProfileCmd.Flags().StringP("iface", "i", "", "Select network interface")

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(listCmd, testCmd, applyCmd, statusCmd, rollbackCmd, addProfileCmd, autoCmd, deleteProfileCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
