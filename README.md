# CodeCollab
===========
Online collaborative code editor. The server is written in Go to leverage the grace of concurrency patterns.

###Install
```sh
> git clone https://github.com/filet-mign0n/CodeCollab && go test && go install
```
###Launch
```sh
> $GOPATH/CodeCollab
```
###Chose your host and port
```sh
> $GOPATH/CodeCollab -h 10.0.0.1 -p 8080	#default is localhost:8000
```
###Two levels of logging
```sh
> $GOPATH/CodeCollab -v		#for client and server activity
> $GOPATH/CodeCollab -d		#for debugging
```