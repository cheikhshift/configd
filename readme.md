# configd-community edition

Config'd offers configuration management as a service. Read the guide below to setup an instance.

# Builds

[Linux Binary](#)


# Requirements (to build)

- Go 1.8+
- Environment variable Path has `$GOPATH/bin` in it.
- MongoDB v3.2+

# Get source

	go get github.com/cheikhshift/configd

# Run command

## Add SSL files
Make sure you have a server key file with name `server.key` in your working directory. 

Make sure you have a server certificate file with name `server.crt` in your working directory. 

	configd


### Packages used :
- github.com/cheikhshift/gos
- Twitter Bootstrap. 
- JSONEditor by Jos de Jong.

### Contribution

PRs are encouraged and appreciated. 
