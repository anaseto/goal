#!/bin/sh

perl genop.pl > dyadops.go
gofmt -s -w dyadops.go
