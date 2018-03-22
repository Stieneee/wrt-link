#!/bin/bash

sed -e 's/\[UNREPLIED\]//' $1 | awk '
/tcp/ { printf "%s %s %s %s %s \n", $1, $5, $6, $10, $16 }
/udp/ { printf "%s %s %s %s %s \n", $1, $4, $5, $9, $15 }
' | sed -e 's/src=//' -e 's/dst=//' -e 's/bytes=//g' > .samples/output.txt

# sed -e 's/\[UNREPLIED\]//' $1
