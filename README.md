# wrt-link

wrt-link is a shell script for collecting statistcs and per device bandwidth monitoring with DD-WRT routers.
The script reports information from a number of sources including the nvram, iptables and conntrack via scp.

## Prerequisites

A DD-WRT router with ssh access enabled.
A seperate ssh server.

In the event this script is being used in combination with logmy.io please follow the installation steps provided by that service.

## Installing

Ideally installation should occur via a startup script saved in the router's administraiton settings.
For testing purposes the script can be deployed on the router manually.

### Manual Setup Example

SSH into your router and run the following commands.
Generating a new rsa key and user accout on the remote server to be used by the script is advised.

```
echo {{SSH ADDRESS}} {{REMOTE SERVER PUBLIC HOSTKEY}} > /tmp/root/.ssh/known_hosts
cat > /tmp/wrt-link.id_rsa <<- EOM
{{PRIVATE KEY}}
EOM
wget http://github.com/Stieneee/wrt-link/releases/download/latest/wrt-link.sh -O /tmp/wrt-link.sh
/tmp/wrt-link.sh {ROUTER ID} {SERVER ADDRESS} {SERVER PORT}
```

Alternatively scp could be used to retrieve the file securely from the ssh server.

### Handling Report File

The report file contains information in rows with several unique formats.
Each format has a two character identifier at the beginning of the line.

#### Version Information

Version information is reported in the frist report after the device restarts.

```
wl {WRT-LINK Version}
dv {DD-WRT Version}
se {SFE Enabled}
```

#### Ping Report

The result of a running ping command is concatinated and included with each report.
The results can be parsed to determine packet loss and average ping statistics.

```
pt {Time (ms)} 
pu # An unreachable host reponse.
po # A timeout response.
```

#### Iptables Report

MAC, IPs and iptables counters are reported.
One line is present for each client of the devices regardless of whether or not the counters are non-zero.

```
nf {MAC} {IP} {Download} ${Upload}
```

#### Conntrack

A row for each row from /proc/net/ip_conntrack.
These rows have been condensed to save required bandwidth.

```
ct {Protocol} {Source IP} {Destination IP} {Source Port} {Destination Port} {Download Bytes} {Upload Bytes}
```

### Calculating Usage Data
Byte counters on the router are reset during each cycle for iptables but not for conntrack.
To properly account bandwidth information for each device differences must be calculated for the conntrack byte counters between each report.

## Contributing
Issues and pull requests are welcome.

## License
This project is licensed under the GPL-3.0 license - see the LICENSE file for details.

## Acknowledgments
This script was based on work from wrtbwmon.
