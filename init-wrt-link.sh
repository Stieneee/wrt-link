#!/bin/sh

echo 'init-wrt-link.sh version VERSIONPLACEHOLDER' >> /tmp/wrt-link.log

_stop() {
  test -f /tmp/stop-wrt-link && echo "Stopping!" && rm /tmp/stop-wrt-link && exit 0 || return 0
}

downloadAndRun() {
  echo $CPU_TYPE CPU Detected >> /tmp/wrt-link.log
  until scp -P 222 public@get.logmy.io:latest/wrt-link_linux_$CPU_TYPE /tmp/wrt-link; do 
    echo "Failed to download wrt-link_linux_$CPU_TYPE. trying again 5 second" >> wrt-link.log
    sleep 5 
  done

  chmod u+x /tmp/wrt-link

  while true
  do
    _stop
    /tmp/wrt-link $1 $2 $3 2>> /tmp/wrt-link.log
    echo 'wrt-link exited restarting in 30 seconds' >> /tmp/wrt-link.log
    sleep 30
  done
}

# ash supported substring search using grep
if cat /proc/cpuinfo | grep mips64le - > /dev/null; then
  CPU_TYPE=mips64le
  downloadAndRun $1 $2 $3
elif cat /proc/cpuinfo | grep mips64 - > /dev/null; then
  CPU_TYPE=mips64
  downloadAndRun $1 $2 $3
elif cat /proc/cpuinfo | grep mips - > /dev/null; then
  CPU_TYPE=mips
  downloadAndRun $1 $2 $3
elif cat /proc/cpuinfo | grep mipsle - > /dev/null; then
  CPU_TYPE=mipsle
  downloadAndRun $1 $2 $3
elif cat /proc/cpuinfo | grep arm64 - > /dev/null; then
  CPU_TYPE=arm64
  downloadAndRun $1 $2 $3
elif cat /proc/cpuinfo | grep arm - > /dev/null; then
  CPU_TYPE=arm
  downloadAndRun $1 $2 $3
elif cat /proc/cpuinfo | grep ppc64le - > /dev/null; then
  CPU_TYPE=ppc64le
  downloadAndRun $1 $2 $3
elif cat /proc/cpuinfo | grep ppc64 - > /dev/null; then
  CPU_TYPE=ppc64
  downloadAndRun $1 $2 $3
elif cat /proc/cpuinfo | grep amd64 - > /dev/null; then
  CPU_TYPE=amd64
  downloadAndRun $1 $2 $3
elif cat /proc/cpuinfo | grep 386 - > /dev/null; then
  CPU_TYPE=386
  downloadAndRun $1 $2 $3
elif cat /proc/cpuinfo | grep s390x - > /dev/null; then
  CPU_TYPE=s390x
  downloadAndRun $1 $2 $3
else
  echo 'ERROR: CPU type not detected!' >> /tmp/wrt-link.log
  exit 1
fi

