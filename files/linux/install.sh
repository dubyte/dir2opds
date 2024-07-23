#!/bin/sh
name="dir2opds"

# bin
install -m755 bin/${name} /usr/local/bin/${name}

# rc
if [ -d /lib/systemd/system/ ]; then
	install -m644 files/systemd/${name}.service /lib/systemd/system/
elif [ -d /etc/init.d/ ]; then
	install -m644 files/init.d/${name} /etc/init.d/
fi

# log rotation
install -m644 files/logrotate.d/${name} /etc/logrotate.d/
