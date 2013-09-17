all:
	go build

install: all
	install -d -m 0755 /usr/share/bor
	install -m 0644 share/Makefile.tpl /usr/share/bor/Makefile.tpl
	install -m 0644 share/TAPListener.cpp /usr/share/bor/TAPListener.cpp
	install -m 0644 bor.conf /etc/bor.conf
	install -m 0644 bor /usr/bin/bor
