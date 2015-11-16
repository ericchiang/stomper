# stomper - A Docker image squasher

## Installation

```
go get github.com/ericchiang/stomper
```

## Usage

```
$ sudo docker save golang:1.5  | stomper -i -t squashed | sudo docker load
$ docker history golang:squashed
IMAGE               CREATED             CREATED BY                                      SIZE                COMMENT
f43a05b0404b        27 seconds ago      /bin/sh -c #(nop) COPY file:7e87b0ea22c04c4eb   705.6 MB            
$ docker run -it --rm golang:squashed
root@ae60e353be94:/go# go version
go version go1.5.1 linux/amd64
root@ae60e353be94:/go# 
```
