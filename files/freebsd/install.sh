#!/bin/sh
name="dir2opds"

# bin
install -m755 bin/${name} /usr/local/bin/${name}

# rc
mkdir -p /usr/local/etc/rc.d/
install -m755 files/rc.d/${name} /usr/local/etc/rc.d/

# log rotation
mkdir -p /usr/local/etc/newsyslog.conf.d/
install -m644 files/newsyslog.conf.d/${name}.conf /usr/local/etc/newsyslog.conf.d/
