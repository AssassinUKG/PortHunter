# PortHunter - The Ultimate Port Checker

PortHunter is a fast and efficient port scanning tool that automates the process of running Nmap or Masscan, storing results, and comparing them to detect network changes over time. 

## Features

- 🔍 **Supports both Nmap & Masscan** - Automates scanning with your preferred tool.
- 📊 **Scan comparison** - Detects added or removed ports between scans.
- 🗄️ **JSON output & storage** - Saves scan results in structured JSON format.
- ⏳ **Timestamp tracking** - Tracks when each scan was performed.
- 🚀 **Multi-threaded processing** - Handles large-scale scans efficiently.
- 🔕 **Silent mode** - Runs without verbose console output.

## Installation

### Prerequisites
- [Go](https://go.dev/doc/install) (1.18+ recommended)
- Nmap or Masscan installed and accessible in `PATH`

### Clone the Repository
```sh
git clone https://github.com/yourusername/porthunter.git
cd porthunter
```

### Build & Install
```sh
go build -o porthunter
```

## Usage

### Basic Scan
```sh
./porthunter -c "nmap -p- -T4" -t "192.168.1.1"
```


### Silent Mode
```sh
./porthunter -c "nmap -p- -T4" -t "192.168.1.1" -s
```

### Scan Comparison
If previous scan data exists, PortHunter will automatically compare the new scan results and display differences (added/removed ports).

## Example Output
```
--- Checking Previous Scan Data (Last scan was 2 hours ago) ---

Changes for 192.168.1.1:
  [+] Added Ports:
    - 443 (https)
  [-] Removed Ports:
    - 80 (http)

Summary: 1 new port added, 1 removed.
```

## Roadmap & Future Improvements
- 🔄 **Resume feature** - Restart from the last interrupted scan.
- 🔔 **Notification support** - Send alerts via Slack or email when changes are detected.
- 📡 **Multi-target scanning** - Support for scanning multiple IPs in a single execution.

## Contributions
Contributions are welcome! If you have ideas or improvements, feel free to submit an issue or a pull request.

## License
This project is licensed under the MIT License.

## Author
Created by Richard Jones 🚀