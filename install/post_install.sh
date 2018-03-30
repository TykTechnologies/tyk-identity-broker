#!/bin/bash
UNIT_NAME="tyk-identity-broker"

echo "Setting permissions"
# Config file must not be world-readable due to sensitive data
chown -R tyk:tyk /opt/$UNIT_NAME
chmod 660 /opt/$UNIT_NAME/tib.conf

echo "Installing init scripts..."

SYSTEMD="/lib/systemd/system"
UPSTART="/etc/init"
SYSV1="/etc/init.d"
SYSV2="/etc/rc.d/init.d/"
DIR="/opt/$UNIT_NAME/install"

if [ -d "$SYSTEMD" ] && systemctl status > /dev/null 2> /dev/null; then
	echo "Found Systemd"
	[ -f /etc/default/$UNIT_NAME ] || cp $DIR/inits/systemd/default/$UNIT_NAME /etc/default/
	cp $DIR/inits/systemd/system/$UNIT_NAME.service /lib/systemd/system/
	systemctl --system daemon-reload
	exit
fi

if [ -d "$UPSTART" ]; then
	[ -f /etc/default/$UNIT_NAME ] || cp $DIR/inits/upstart/default/$UNIT_NAME /etc/default/
	if [[ "$(initctl version)" =~ .*upstart[[:space:]]1\..* ]]; then
		echo "Found upstart 1.x+"
		cp $DIR/inits/upstart/init/1.x/$UNIT_NAME.conf /etc/init/
	else
		echo "Found upstart 0.x"
		cp $DIR/inits/upstart/init/0.x/$UNIT_NAME.conf /etc/init/
	fi
	exit
fi

if [ -d "$SYSV1" ]; then
	echo "Found SysV1"
	[ -f /etc/default/$UNIT_NAME ] || cp $DIR/inits/sysv/default/$UNIT_NAME /etc/default/
	cp $DIR/inits/sysv/init.d/$UNIT_NAME /etc/init.d/
	exit
fi

if [ -d "$SYSV2" ]; then
	echo "Found Sysv2"
	[ -f /etc/default/$UNIT_NAME ] || cp $DIR/inits/sysv/default/$UNIT_NAME /etc/default/
	cp $DIR/inits/sysv/init.d/$UNIT_NAME /etc/rc.d/init.d/
	exit
fi
