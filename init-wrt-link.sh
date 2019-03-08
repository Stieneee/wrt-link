#!/bin/sh

echo 'init-wrt-link.sh version 0.2.3' >> /tmp/wrt-link.log

# ash supported substring search using grep
if cat /proc/cpuinfo | grep mips - > /dev/null; then
  echo 'MIPS CPU Detected' >> /tmp/wrt-link.log
  until scp -P 222 public@get.logmy.io:latest/wrt-link-mips /tmp/wrt-link; do 
    echo "Failed to download wrt-link-mips. trying again in 5 second" >> wrt-link.log
    sleep 5 
  done
else
  echo 'ERROR: CPU type not detected!' >> /tmp/wrt-link.log
  exit 1
fi

chmod u+x /tmp/wrt-link

_stop() {
  test -f /tmp/stop-wrt-link && echo "Stopping!" && rm /tmp/stop-wrt-link && exit 0 || return 0
}

while true
do
  _stop
  /tmp/wrt-link $1 $2 $3 2>> /tmp/wrt-link.log
  echo 'wrt-link exited restarting in 30 seconds' >> /tmp/wrt-link.log
  sleep 30
done