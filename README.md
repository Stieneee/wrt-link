# wrt-link

wrt-link is a Go application for the collection of bandwidth, connection and device stats from DD-WRT routers.
Originally built to report to logmy.io the application reports periodically to an HTTP and could report to independently hosted endpoints.

## Features

- [x] Bandwidth Logging
  - [x] iptables (wrtbwmon)
  - [x] ipconntrack scrapping for bandwidth
- [x] Device connection reporting scrapping ipconntrack
  - [ ] Configurable private devices
  - [ ] Black list checking
- [ ] Router Stats
  - [x] Device type, CPU type
  - [x] DD-wrt version, kernel version
  - [x] CPU
  - [x] Memory
  - [ ] Space, NVRAM, CIFS, JFFS2
  - [x] Active IP Connections
  - [x] Device Hostnames
  - [ ] Wireless Settings
  - [ ] Wireless Clients
  - [ ] Nearby Wireless Devices
- [ ] ISP Stats
  - [x] Ping
  - [ ] Speedtest
  - [ ] DNS response time
- [ ] Simple Endpoint Example - (Probably Nodejs and Mongodb in a Docker)

## Prerequisites

A DD-WRT router with ssh access enabled.
A reporting endpoint or an account with logmy.io.
A 256 Byte RSA private key.

In the event this is being used in combination with logmy.io please follow the installation steps provided by logmy.io.
There is no need to read any of the information here.

## Compiling

```bash
make all
```

## Installing

Ideally installation should occur via a startup script saved in the router's administration settings.
For testing purposes the script can be deployed on the router manually.
A full example will be made available in the future.

### Manual Setup Example

SSH into your router and run the following commands.
Generating a new rsa key and user account on the remote server to be used by the script is advised.

Alternatively scp could be used to retrieve the file securely from the ssh server.

### Generate a New Private Key

```bash
# Location: Desktop
ssh-keygen -t rsa -f ~/wrt-link.id_rsa
scp ~/wrt-link-id_rsa root@192.168.1.1:/tmp/wrt-link.id_rsa
```

### SCP Gateway

A public scp gateway is available for easy retrieval of the latest binary.
The host key of the gateway must be added to the know_hosts.

```bash
# On Router
get.logmy.io ssh-rsa "AAAAB3NzaC1yc2EAAAADAQABAAABAQC0zauQB43Zn2xReW3ULrP09ckJxK6rZ+V45SFIQ9J88AnjMhaZ/YVjr8FBRXsBWk3Mqgx38D4WfOpvpMTWieaA3xJoLvVVBWKp5Sm+hfZdsDoJFwI23POG2cJvsM08bvq7ifnXcQs5uncTR26sa60ZEfmWKvw7GXvXnbjb2XsnPzzJytVcVAblH4piaQzt6iLlb436iEBgMqzJaxemDQsX47uZhbcfKG+YCZEr/uyJMUWZbnhfpkme1YpW4Ob1cNf1Ff/aijUnir6qooVVMybRg8HmWkgV6gqzDGKn+yAEcSFXcZks39bwnM/ffzVe1qvvMQR55NcJ0jZihyVhFlpF" >> /tmp/root/.ssh/known_hosts
scp -P 222 public@get.logmy.io:latest/wrt-link-mips /tmp/wrt-link
```

The 'latest' directory on the remote host can be replaced with the version tag i.e `public@get.logmy.io:0.2.0/wrt-link-mips`.

### Calling

The binary is called with the following three arguments.

- API_ENDPOINT - The reporting HTTP endpoint.
- ROUTER_ID - The unique ID of the router.
- KEY_FILE - The file of the private key. `/tmp/wrt-link.id_rsa` from the example above.

```bash
/tmp/wrt-link {API_ENDPOINT} {ROUTER_ID} {KEY_FILE}
```

## Contributing

Issues and pull requests are welcome.

## License

This project is licensed under the GPL-3.0 license - see the LICENSE file for details.

## Acknowledgments

This script was based on work from wrtbwmon.
