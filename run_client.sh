#!/bin/sh
set -e
USAGE="./run_client.sh OPERATION [KEY] [VALUE]"
[[ -z "$1" ]] && (echo $USAGE; exit 1)
[ -n "$AWS_SECRET_ACCESS_KEY" ] || ( echo "AWS_* env vars are not set"; exit 1 )
QUEUE=$(terraform output -raw sqs_url)
[[ ! -z "$2" ]] && KEY="-key $2"
[[ ! -z "$3" ]] && VALUE="-val $3"
go run cmd/client/main.go -queue $QUEUE -op $1 $KEY $VALUE
