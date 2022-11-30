#!/bin/sh

perl genop.pl > arithd.go
gofmt -s -w arithd.go
