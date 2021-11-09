### tplink-hs1x-cli

A simple app to control TPLink HS100 and HS110 devices.

Sometimes it is just easier to use a simple cli tool to turn the lights on and off.  This is that tool.


## Background

There are two ways to control TP-Link HS1x0 devices, firstly by the TP-Links cloud and secondly by sending command directly to the devices.

The cloud method is slower and relies on internet access, sending commands directly is quicker, but insecure.

The TP-Link devices have port 9999/tcp open which allows properly constructed and encrypted json to issue commands and get responses.  The encryption method is simple and well documented, further details are available in the Credit section.

## Installation

You can install a few ways:

1. Download the binary for your OS from https://github.com/smford/tplink-hs1x-cli/releases
1. or use `go install`
   ```
   go install -v github.com/smford/tplink-hs1x-cli@latest
   ```
1. or clone the git repo and build
   ```
   git clone git@github.com:smford/tplink-hs1x-cli.git
   cd tplink-hs1x-cli
   go get -v
   go build
   ```

## Configuration

Create a configuration file called `config.yaml` an example is available below:
```
---
devices:
  small: 192.168.10.44
  large: 192.168.10.127
```

The configuration file has a list of devices, a human readable name, and the IP address of the device.  The human readable name is used to issue commands against the devices.

When tplink-hs1x-cli runs it checks the current directory for a `config.yaml`, if you wish to use a different configuration file use the command `--config /path/to/file.yaml`

## Command Line Options
```
      --config [file]       Configuration file: /path/to/file.yaml (default: "./config.yaml")
      --debug               Display debug information
      --device [string]     Device to apply "do action" against
      --displayconfig       Display configuration
      --do <action>         on, off, status, info, cloudinfo, ledon, ledoff, wifiscan, getaction, gettime, getrules, getaway, reboot, antitheft, factoryreset, energy (default: "on")
      --help                Display help
      --list                List devices
      --version             Display version
```

## Actions
| Action | Details |
|:--|:--|
| antitheft | Display anti-theft configuration |
| cloudinfo | Display TP-Link cloud information |
| energy | Display enegery information |
| factoryreset | Factory reset the device |
| getaction | Display actions |
| getaway | Display configurared away information |
| getrules | Display configured rules |
| gettime | Display configured time |
| info | Display detailed information on a device |
| ledon | Turn LED on (night mode) |
| ledff | Turn LED off (night mode) |
| off | Turn off |
| on | Turn on |
| reboot | Reboot device |
| status | Display current status of a device |
| wifiscan | Display wifi networks that the device can see |


##  Example Usage



## Credit, High Fives & Useful Links
- https://github.com/softScheck/tplink-smartplug
- https://www.softscheck.com/en/reverse-engineering-tp-link-hs110/
- https://github.com/softScheck/tplink-smartplug/blob/master/tplink-smarthome-commands.txt
- https://github.com/sausheong/hs1xxplug
