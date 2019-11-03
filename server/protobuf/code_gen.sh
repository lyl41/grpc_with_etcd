#!/bin/bash

protoc  -I. -I$GOPATH/src --go_out=plugins=grpc:. ./*.proto