### tplink-hs1x-cli

A simple app to control TPLink HS100 and HS110 devices.

Sometimes it is just easier to use a simple cli tool to turn the lights on and off.  This is that tool.


## Background

There are two ways to control TP-Link HS1x0 devices, firstly by the TP-Links cloud and secondly by sending command directly to the devices.

The cloud method is slower and relies on internet access, sending commands directly is quicker, but insecure.

The TP-Link devices have port 9999/tcp open which allows properly constructed and encrypted json to issue commands and get responses.  The encryption method is simple and well documented, further details are available in the Credit section.

## Installation

## Configuration

Create a configuration file like the below.  It comprises of a human readable name and the IP address of the device.  You will use the human readable name when issuing commands from the tool.

```
---
devices:
  small: 192.168.10.44
  large: 192.168.10.127
```

## Usage


## Credit, High Fives & Useful Links
- https://github.com/softScheck/tplink-smartplug
- https://www.softscheck.com/en/reverse-engineering-tp-link-hs110/
- https://github.com/softScheck/tplink-smartplug/blob/master/tplink-smarthome-commands.txt
- https://github.com/sausheong/hs1xxplug
