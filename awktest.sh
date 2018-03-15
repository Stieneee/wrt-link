#!/bin/bash

awk '
/tcp/ { printf "%s %s %s %s %s %s %s \n", $1, $5, $6, $7, $8, $10, $16 }
/udp/ { printf "%s %s %s %s %s %s %s \n", $1, $4, $5, $6, $7, $9, $15 }
' $1 | sed -e 's/src=//' -e 's/dst=//' -e 's/sport=//' -e 's/dport=//' -e 's/bytes=//g' > output.txt
