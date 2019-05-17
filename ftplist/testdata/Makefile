.DELETE_ON_ERROR:

PUREFTPD	:= pure-ftpd-1.0.49
PUREFTPD_TAR	:= $(PUREFTPD).tar.bz2
PUREFTPD_256	:= 8a727dfef810f275fba3eb6099760d4f8a0bdeae2c1197d0d5bfeb8c1b2f61b6
PUREFTPD_URL	:= https://download.pureftpd.org/pub/pure-ftpd/releases/$(PUREFTPD_TAR)
PUREFTPD_PEM	:= $(CURDIR)/ftpd.pem

all: $(PUREFTPD)/src/pure-ftpd ftpd.sh

ftpd.sh: Makefile
	echo "#!/bin/bash -e" > $@
	echo "export FTP_ANON_DIR=\"$(CURDIR)\"" >> $@
	echo "exec \"$(CURDIR)/$(PUREFTPD)/src/pure-ftpd\" \"\$$@\"" >> $@
	chmod +x $@

$(PUREFTPD)/src/pure-ftpd: $(PUREFTPD)/Makefile
	$(MAKE) -C $(PUREFTPD)

$(PUREFTPD)/Makefile: $(PUREFTPD)/configure
	cd $(PUREFTPD) && ./configure --with-nonroot --with-puredb --with-tls --with-certfile="$(PUREFTPD_PEM)"

$(PUREFTPD)/configure: $(PUREFTPD_TAR)
	echo "$(PUREFTPD_256) $(PUREFTPD_TAR)" | sha256sum -c -
	tar -xjf $(PUREFTPD_TAR)
	touch $(PUREFTPD)/configure

$(PUREFTPD_TAR):
	wget $(PUREFTPD_URL)

clean:
	rm -rf $(PUREFTPD) $(PUREFTPD_TAR)
