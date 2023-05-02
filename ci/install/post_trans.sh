#!/bin/sh

# Generated by: gromit policy
# Generated on: Fri Apr 28 10:48:32 UTC 2023

if command -V systemctl >/dev/null 2>&1; then
    if [ ! -f /lib/systemd/system/tyk-identity-broker.service ]; then
        cp /opt/tyk-identity-broker/install/inits/systemd/system/tyk-identity-broker.service /lib/systemd/system/tyk-identity-broker.service
    fi
else
    if [ ! -f /etc/init.d/tyk-identity-broker ]; then
        cp /opt/tyk-identity-broker/install/inits/sysv/init.d/tyk-identity-broker /etc/init.d/tyk-identity-broker
    fi
fi
