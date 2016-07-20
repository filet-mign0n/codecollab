# CodeCollab

Online collaborative code editor. The server is written in Go to leverage the grace of concurrency patterns.
###Requirement
Go 1.5 or above
###Install
```sh
> cd $GOPATH/src && git clone https://github.com/filet-mign0n/codecollab \
> && cd codecollab && go get ./... && go test && go install
```
###Launch
```sh
> $GOPATH/bin/codecollab
```
###Chose your host and port
```sh
> $GOPATH/bin/codecollab -h 10.0.0.1 -p 8080	# default is localhost:8000
```
###Two levels of logging
```sh
> $GOPATH/bin/codecollab -v	-d	# v for verbosity (client and server activity), d for debug
```
