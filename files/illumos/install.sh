#!/bin/sh
name="dir2opds"

# bin
mkdir -p /usr/local/bin
install -f /usr/local/bin -m 755 bin/${name}

# SMF manifest
install -f /lib/svc/manifest/network -m 644 files/manifest/${name}.xml
svccfg validate /lib/svc/manifest/network/${name}.xml
svccfg import /lib/svc/manifest/network/${name}.xml
