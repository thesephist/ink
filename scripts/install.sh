#!/bin/sh

go install -ldflags="-s -w"
ls -l `which ink`
