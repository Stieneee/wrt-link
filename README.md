# wrt-link

wrt-link is a Go application for collection bandwidth, conneciton and device stats from DD-WRT routers.
Orignally built to report to logmy.io the application reports perodically via an HTTP and could report to independantly hosted endpoints.

## Features

- [ ] Bandwidth Logging
- [ ]   iptables (wrtbwmon)
- [ ]   ipconntrack scrapping for bandwidth
- [ ] Device connection reporting scrapping ip conntack
- [ ]   Configurable private devices
- [ ]   Black list checking
- [ ] Router Stats
- [ ]   Device type, CPU type
- [ ]   DD-wrt version, kernel version
- [ ]   CPU
- [ ]   Memory
- [ ]   Space, NVRAM, CIFS, JFFS2
- [ ]   Active IP Connections
- [ ]   Wireless Settings
- [ ]   Wireless Clients
- [ ]   Nearby Wireless Devices
- [ ] ISP Stats
- [ ]   Ping
- [ ]   Speedtest
- [ ]   DNS response time


- [ ] Simple Endpoint Example - (Probably Nodejs and Mongodb in a Docker)

## Prerequisites

A DD-WRT router with ssh access enabled.
A reporting endpoint or an account with logmy.io.
A 256 Byte RSA private key.

In the event this is being used in combination with logmy.io please follow the installation steps provided by that service.
There is no need to read any of the inforamtion here.

## Compiling

// TODO

## Installing

Ideally installation should occur via a startup script saved in the router's administraiton settings.
For testing purposes the script can be deployed on the router manually.

### Manual Setup Example

SSH into your router and run the following commands.
Generating a new rsa key and user accout on the remote server to be used by the script is advised.

//TODO

Alternatively scp could be used to retrieve the file securely from the ssh server.

### Configure

### Sentry Error Reporting

## Contributing
Issues and pull requests are welcome.

## License
This project is licensed under the GPL-3.0 license - see the LICENSE file for details.

## Acknowledgments
This script was based on work from wrtbwmon.
