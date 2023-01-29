#!/bin/sh

goal arithd.goal > ../arithd.go
gofmt -s -w ../arithd.go
