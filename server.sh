#!/bin/sh

# Compile Go program
go build

# Quick start-stop-daemon example, derived from Debian /etc/init.d/ssh
set -e

# Must be a valid filename
NAME=fuzzyguy
#This is the command to be run, give the full pathname
DAEMON=$(pwd)'/'$NAME

echo "_______________________________"

case "$1" in
  start)
        echo -n "Starting daemon: "$NAME "\n"
        $DAEMON &
	;;
  stop)
        echo -n "Stopping daemon: "$NAME "\n"
        pkill $NAME
	;;

  *)
	echo "Usage: "$1" {start|stop}"
	exit 1
esac

exit 0
