# 🔁 DNS Switcher CLI
Windows DNS profile manager, latency tester, and automatic fastest DNS selector.
DNS Switcher is a command‑line utility that gives you control over DNS configuration on Windows. It lets you manage DNS profiles, measure latency, apply DNS settings, view current configuration, rollback to previous settings, and automatically select the fastest DNS resolver.

### 📜 Licenses & Third‑Party Libraries

This project uses the following open‑source libraries:

1- **cobra**  
License  ==  [License](https://github.com/spf13/cobra?tab=Apache-2.0-1-ov-file "License cobra")
Apache-2.0 license

---------------------------------------

2- **Viper**

License  ==  [License](https://github.com/spf13/viper?tab=MIT-1-ov-file "License Viper")
Copyright (c) 2014 Steve Francia

---------------------------------------

3- **dns**

License  ==  [License](https://github.com/miekg/dns?tab=BSD-3-Clause-1-ov-file "License dns")
Copyright (c) 2009, The Go Authors. Extensions copyright (c) 2011, Miek Gieben.
All rights reserved.

---------------------------------------

## 🚀 Features
### DNS Profile Management
- Create and list DNS profiles
- Multiple servers per profile
- Add new profiles via add-profile command
###  Performance Testing
- Measure RTT for a single server or an entire profile
- Error‑aware reporting
###  DNS Control (Windows only)
- Apply a chosen DNS profile with apply
- Show current DNS settings with status
###  Rollback & Recovery
- Restore previous DNS settings with rollback
###  Auto Mode
- Measure latency across all profiles
- Automatically select and apply the fastest profile
- Show benchmark results
### Network Interface Selection
- Explicitly choose which network adapter to apply DNS changes
- Useful for multi‑adapter systems (Wi‑Fi, Ethernet)

## 🎯 Audience
Perfect for Windows developers, gamers, network admins, and privacy‑conscious users.

> [!IMPORTANT]
>Note: The DNS profiles provided are optional and serve only as examples. You can create and use your own profiles as needed

## 📖 Command Reference – DNS Switcher CLI
Below is the full list of available commands with explanations and usage examples.
Run dns-switcher [command] --help for detailed options.

> [!IMPORTANT]
>Be sure to set the local variable before using the package:
> 
> set DNS_SWITCHER_PATH=Address
> 
> set DomainTesting=example.com

## 1. add-profile (Add a new DNS profile)
Create and store a custom DNS profile with a name and one or more servers.
Usage:
```
dns-switcher add-profile --name mydns --servers 1.1.1.1,1.0.0.1
```

Example Output:
```
Profile 'mydns' added: [1.1.1.1 1.0.0.1]
```


### 2. apply (Apply a DNS profile)
Switch the system DNS settings to a specific profile.
Usage:
```
dns-switcher apply cloudflare
dns-switcher apply cloudflare -f   # force reapply even if already active
```

Example Output:
```
Applied profile 'cloudflare'
```

### 3. auto (Auto-select the fastest DNS profile)
Automatically tests all profiles and applies the one with the lowest latency.
Usage:
```
dns-switcher auto
dns-switcher auto -r 5   # repeat each test 5 times
dns-switcher auto -a     # apply fastest profile automatically
```

Example Output:
```
Applied fastest profile 'cloudflare'
1.1.1.1 -> RTT[1]: 20ms
1.0.0.1 -> RTT[1]: 22ms
```

### 4. completion (Generate shell autocompletion scripts)
Generate autocompletion scripts for Bash, Zsh, Fish, or PowerShell.
This improves usability by enabling tab-completion for commands and flags.
Usage:
```
dns-switcher completion bash
dns-switcher completion zsh
dns-switcher completion fish
dns-switcher completion powershell
```

### 5. delete-profile (Delete one or more DNS profiles)
Delete one or more stored DNS profiles by name. Supports backup and confirmation options.
Usage:
```
dns-switcher delete-profile google
dns-switcher delete-profile cloudflare mydns -y   # skip confirmation
dns-switcher delete-profile google -f             # force delete even if active
dns-switcher delete-profile google -j             # JSON output
dns-switcher delete-profile google -q             # suppress success message
```

Example Output: 
```
Profile 'google' deleted
```

### 6. list (List available DNS profiles)
Show all stored DNS profiles with their associated servers.
Usage:
```
dns-switcher list
dns-switcher list -v   # verbose, show servers
```

Example Output:
```
Available profiles:
- google : [8.8.8.8 8.8.4.4]
- cloudflare : [1.1.1.1 1.0.0.1]
```

### 7. rollback (Rollback to previous DNS settings)
Restore the DNS configuration to the state before the last change.
Usage:
```
dns-switcher rollback
dns-switcher rollback -q   # suppress success message
```

Example Output:
```
Rollback successful – DNS restored to automatic (DHCP)
```

### 8. status (Show current DNS settings)
Display the DNS servers currently applied on the system.
Usage:
```
dns-switcher status
dns-switcher status -j   # JSON output
```

Example Output:

```
Current DNS servers:
- 1.1.1.1
- 1.0.0.1
```

### 9. test (Test latency for a profile or server)
Measure round-trip time (RTT) for either a profile’s servers or a single server.
Usage:

```
dns-switcher test google
dns-switcher test 8.8.8.8 -r 3   # repeat test 3 times
```

Example Output:
```
Testing profile 'google'
8.8.8.8 -> RTT[1]: 34ms
8.8.4.4 -> RTT[1]: 40ms
```

### 🎯 Flags (Global & Common)

- -h, --help → Show help for any command.
- -v, --verbose → Verbose output (list).
- -r, --repeat → Number of times to repeat RTT test (test, auto).
- -f, --force → Force apply/delete even if active (apply, delete-profile).
- -q, --quiet → Suppress success message (rollback, delete-profile).
- -j, --json → Output result in JSON format (status, delete-profile).
- -n, --name → Profile name (add-profile).
- -s, --servers → Comma-separated DNS servers (add-profile).
- -a, --apply → Apply fastest profile automatically (auto).
- -y, --yes → Skip confirmation prompt (delete-profile).
- -i, --iface → Specify network interface (The user must explicitly select a network interface (e.g., Wi‑Fi, Ethernet, VPN) before applying or testing DNS profiles.
  Without specifying an interface, the tool will not proceed.)
- --no-backup → Do not create a backup before deletion (delete-profile).