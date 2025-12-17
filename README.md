
# netcfgdiff
**netcfgdiff** is a lightweight, context-aware configuration diff tool for network devices (e.g., Cisco IOS), written in Go.  
Unlike standard Unix `diff`, `netcfgdiff` understands the hierarchical structure of network configurations (indentation-based). It allows you to see exactly which parent block a change belongs to, making it easier to review changes before deployment.

![License](https://img.shields.io/badge/license-GPL--3.0--or--later-blue.svg)
![Go](https://img.shields.io/badge/language-Go-00ADD8.svg)

## Features

* **Context-Aware Diff:** Detects parent-child relationships using indentation. If a child line changes, the parent line is displayed for context.
* **Colorized Output:** Clearly highlights additions (green) and removals (red).
* **Ignore Noise:** Filter out irrelevant lines (e.g., timestamps, encrypted passwords) using regex.
* **Target Scope:** Focus comparison on specific blocks (e.g., `router ospf`) and ignore the rest.
* **Single Binary:** Easy to distribute to bastion hosts or servers without Python dependencies.

## Installation
### From Source

```bash
git clone https://github.com/ksaegusa/netcfgdiff.git
cd netcfgdiff
go build -o netcfgdiff ./cmd/netcfgdiff
```

## Usage
Basic comparison between running config and candidate config:

```bash
./netcfgdiff running.conf candidate.conf
```

### Options

| Flag | Description |
| --- | --- |
| `-i`, `--ignore` | Regex pattern to ignore lines (can be used multiple times). |
| `-t`, `--target` | Target block prefix to limit the scope of comparison. |
| `-h`, `--help` | Help for netcfgdiff. |

### Examples

**1. Ignore specific lines (e.g., NTP clock or timestamps):**

```bash
./netcfgdiff old.conf new.conf -i "^ntp clock-period" -i "^! Last configuration"
```

**2. Compare only a specific block (e.g., OSPF configuration):**

```bash
./netcfgdiff old.conf new.conf --target "router ospf"
```

## Output Example

```diff
interface GigabitEthernet1
- description Management Interface
+ description Management Interface (Updated)
  ip address 10.0.0.1 255.255.255.0
+ interface GigabitEthernet2
+   description New Interface
+   no shutdown
router bgp 65000
- neighbor 10.0.0.2 remote-as 65001
```

* **White lines:** Context (parent)
* **Red lines (-):** Removed lines
* **Green lines (+):** Added lines

## License & Acknowledgements

This project is a Go port/derivative work based on the configuration parsing logic found in
[ansible-collections/ansible.netcommon](https://github.com/ansible-collections/ansible.netcommon).

* Original Source: ansible-collections/ansible.netcommon
* Original License: GNU General Public License v3.0 or later

Consequently, this project (netcfgdiff) is released under the **GNU General Public License v3.0 or later**.
See the [LICENSE](LICENSE) file for details.