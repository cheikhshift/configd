# configd-community edition

Config'd offers configuration management as a service. Read the guide below to setup an instance.

# Builds

- [Linux Binary](https://github.com/cheikhshift/configd/raw/master/configd-server.tar.gz)
	- md5 checksum of binary : 98f2ac5c64be45f61d7c773e97fbc988

# Install binary on linux

### Download binary

Run the following command to download binary

	curl  https://github.com/cheikhshift/configd-injector/raw/master/configd-server.tar.gz \
  	--output configd-server.tar.gz

### Decompress archive

	tar -pxvzf configd-server.tar.gz

### Install command

	sudo mv configd-server /usr/sbin/


### Run command

	configd-server

The server will listen on port 8000 without SSL by default.


# Requirements (to build)

- Go 1.8+
- MongoDB v3.2+

# Get source

	go get github.com/cheikhshift/configd

### Packages used :
- github.com/cheikhshift/gos
- Twitter Bootstrap. 
- JSONEditor by Jos de Jong.

### Contribution

PRs are encouraged and appreciated. 
