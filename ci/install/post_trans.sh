#!/bin/sh



# Generated by: tyk-ci/wf-gen
# Generated on: Tue Mar  8 20:25:55 UTC 2022

# Generation commands:
# ./pr.zsh -p -base TD-879/Sync-releng-templates-master -branch TD-879/Sync-releng-templates-master -title internal: Sync releng templates -repos tyk-identity-broker
# m4 -E -DxREPO=tyk-identity-broker


if command -V systemctl >/dev/null 2>&1; then
    if [ ! -f /lib/systemd/system/tyk-identity-broker.service ]; then
        cp /opt/tyk-identity-broker/install/inits/systemd/system/tyk-identity-broker.service /lib/systemd/system/tyk-identity-broker.service
    fi
else
    if [ ! -f /etc/init.d/tyk-identity-broker ]; then
        cp /opt/tyk-identity-broker/install/inits/sysv/init.d/tyk-identity-broker /etc/init.d/tyk-identity-broker
    fi
fi
