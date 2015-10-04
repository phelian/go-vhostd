#!/bin/bash
export GOPATH=$HOME
golint .
go build 