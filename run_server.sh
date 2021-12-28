#!/bin/bash
set -e
QUEUE=$(terraform output -raw sqs_url)
go run cmd/server/server.go -queue $QUEUE -out /tmp/foo.log
