#!/bin/sh

perl arithd.pl > ../arithd.go
gofmt -s -w ../arithd.go
