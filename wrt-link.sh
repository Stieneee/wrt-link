#!/bin/sh
#
# wrt-link - Reporting and traffic logging tool for routers
# Copyright (C) 2018 Tyler Stiene
#
# Based on work from wrtbwmon - Emmanuel Brucey (e.brucy AT qut.edu.au)
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

LAN_IFACE=$(nvram get lan_ifname)

echo "0" > /proc/sys/net/bridge/bridge-nf-call-iptables
echo "0" > /proc/sys/net/bridge/bridge-nf-call-ip6tables
echo "1" > /proc/sys/net/netfilter/nf_conntrack_acct

setupIPLogger() {
	#Create the WRTLINK CHAIN (it doesn't matter if it already exists).
	iptables -N WRTLINK 2> /dev/null

	#Add the WRTLINK CHAIN to the FORWARD chain (if non existing).
	iptables -L FORWARD --line-numbers -n | grep "WRTLINK" | grep "1" > /dev/null
	if [ $? -ne 0 ]; then
		iptables -L FORWARD -n | grep "WRTLINK" > /dev/null
		if [ $? -eq 0 ]; then
			echo "DEBUG : iptables chain misplaced, recreating it..."
			iptables -D FORWARD -j WRTLINK
		fi
		iptables -I FORWARD -j WRTLINK
	fi

	#For each host in the ARP table
	grep ${LAN_IFACE} /proc/net/arp | while read IP TYPE FLAGS MAC MASK IFACE
	do
		#Add iptable rules (if non existing).
		iptables -nL WRTLINK | grep "${IP} " > /dev/null
		if [ $? -ne 0 ]; then
			iptables -I WRTLINK -d ${IP} -j RETURN
			iptables -I WRTLINK -s ${IP} -j RETURN
		fi
	done

	echo "DEBUG: IP Logger Setup"
}

readIPLogger() {
	echo "Reading IP Logger ${1}"
	#Read and reset counters
	iptables -L WRTLINK -vnxZ > /tmp/traffic_$$.tmp

	cat /tmp/traffic_$$.tmp

	grep -v "0x0" /proc/net/arp  | while read IP TYPE FLAGS MAC MASK IFACE
	do
		#Add new data to the graph. Count in Kbs to deal with 16 bits signed values (up to 2G only)
		#Have to use temporary files because of crappy busybox shell TODO investigate
		echo 0 > /tmp/in_$$.tmp
		echo 0 > /tmp/out_$$.tmp
		grep ${IP} /tmp/traffic_$$.tmp | while read PKTS BYTES TARGET PROT OPT IFIN IFOUT SRC DST
		do
			[ "${DST}" = "${IP}" ] && echo $((${BYTES})) > /tmp/in_$$.tmp
			[ "${SRC}" = "${IP}" ] && echo $((${BYTES})) > /tmp/out_$$.tmp
		done
		IN=$(cat /tmp/in_$$.tmp)
		OUT=$(cat /tmp/out_$$.tmp)

		if [ ${IN} -gt 0 -o ${OUT} -gt 0 ];  then
			echo ${MAC} ${IP} ${IN} ${OUT} >> /tmp/wrt-link/${1}.bw.db
		fi
	done

	#Free some memory
	rm -f /tmp/*_$$.tmp
}

readConntrack() {
	echo "DEBUG: Read ip_conntrack"
	sed -e 's/\[UNREPLIED\]//' /proc/net/ip_conntrack | awk '
	/tcp/ { printf "%s %s %s %s %s %s %s \n", $1, $5, $6, $7, $8, $10, $16 }
	/udp/ { printf "%s %s %s %s %s %s %s \n", $1, $4, $5, $6, $7, $9, $15 }
	' | sed -e 's/src=//' -e 's/dst=//' -e 's/sport=//' -e 's/dport=//' -e 's/bytes=//g' >> /tmp/wrt-link/${1}.bw.db

	sh /proc/net/ip_conntrack_flush
}

# Always try to send all files waiting to go
sendFiles() {
	for FILE in $(ls /tmp/wrt-link/)
	do
		echo "DEBUG: scp -i /tmp/wrt-link.id_rsa -P ${3} ${FILE} ${1}@${2}:${FILE}"
		scp -i /tmp/wrt-link.id_rsa -P ${3} /tmp/wrt-link/${FILE} ${1}@${2}:${FILE}
		if [ $? -eq 0 ]; then
		  echo "DEGUB: scp success removing file"
			rm /tmp/wrt-link/${FILE}
		else
		  echo "ERROR: scp failed!"
			break
		fi
	done
}

# Main
if [ -z "${1}" -o -z "${2}" -o -z "${3}" ]; then
	echo "Usage : $0 {ROUTER_ID} {ADDRESS} {PORT}"
	exit
else
	echo "Starting wrt-link..."

	if [ -f /tmp/wrt-link.pid ]; then
		if [ ! -d /proc/$(cat /tmp/wrt-link.pid) ]; then
			echo "NOTICE : wrt-link.pid detected but process $(cat /tmp/wrt-link.pid) does not exist."
			rm -f /tmp/wrt-link.pid
		else
			echo "ERROR: wrt-link already running."
			exit
		fi
	fi

	echo $$ > /tmp/wrt-link.pid

	mkdir /tmp/wrt-link/

	setupIPLogger

	LASTUPDATE=$(date +%s)

	while true
	do
		if [ $(date +%s) -gt $((${LASTUPDATE} + 59)) ]; then # Every 60 seconds
			LASTUPDATE=$(date +%s)
			echo "DEBUG: Update ${LASTUPDATE}"
			readIPLogger ${LASTUPDATE}
			setupIPLogger
			readConntrack ${LASTUPDATE}
			sendFiles "${1}" "${2}" "${3}"
		fi

		sleep 1
	done
fi
