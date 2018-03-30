#!/bin/bash
UNIT_NAME="tyk-identity-broker"

echo "Removing init scripts..."

SYSTEMD="/lib/systemd/system"
UPSTART="/etc/init"
SYSV1="/etc/init.d"
SYSV2="/etc/rc.d/init.d/"
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -f "/lib/systemd/system/$UNIT_NAME.service" ]; then
	echo "Found Systemd"
	echo "Stopping the service"
	systemctl stop $UNIT_NAME.service
	echo "Removing the service"
	rm /lib/systemd/system/$UNIT_NAME.service
	systemctl --system daemon-reload
fi

if [ -f "/etc/init/$UNIT_NAME.conf" ]; then
	echo "Found upstart"
	echo "Stopping the service"
	stop $UNIT_NAME
	echo "Removing the service"
	rm /etc/init/$UNIT_NAME.conf
fi

if [ -f "/etc/init.d/$UNIT_NAME" ]; then
	echo "Found Sysv1"
	/etc/init.d/$UNIT_NAME stop
	rm /etc/init.d/$UNIT_NAME
fi

if [ -f "/etc/rc.d/init.d/$UNIT_NAME" ]; then
	echo "Found Sysv2"
	echo "Stopping the service"
	/etc/rc.d/init.d/$UNIT_NAME stop
	echo "Removing the service"
	rm /etc/rc.d/init.d/$UNIT_NAME
fi
