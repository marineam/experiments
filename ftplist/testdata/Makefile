PUREFTPD	:= pure-ftpd-1.0.49
PUREFTPD_TAR	:= $(PUREFTPD).tar.bz2
PUREFTPD_256	:= 8a727dfef810f275fba3eb6099760d4f8a0bdeae2c1197d0d5bfeb8c1b2f61b6
PUREFTPD_URL	:= https://download.pureftpd.org/pub/pure-ftpd/releases/$(PUREFTPD_TAR)

all: pure-ftpd

pure-ftpd: $(PUREFTPD)/src/pure-ftpd
	cp $< $@

$(PUREFTPD)/src/pure-ftpd: $(PUREFTPD)/Makefile
	$(MAKE) -C $(PUREFTPD)

$(PUREFTPD)/Makefile: Makefile $(PUREFTPD)/configure
	cd $(PUREFTPD) && ./configure --with-nonroot --with-puredb --with-tls

$(PUREFTPD)/configure: $(PUREFTPD_TAR)
	echo "$(PUREFTPD_256) $(PUREFTPD_TAR)" | sha256sum -c -
	tar -xjf $(PUREFTPD_TAR)
	touch $(PUREFTPD)/configure

$(PUREFTPD_TAR):
	wget $(PUREFTPD_URL)

clean:
	rm -rf $(PUREFTPD) $(PUREFTPD_TAR)
