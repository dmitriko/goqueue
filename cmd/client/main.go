package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func keys(m map[string]bool) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

var ops = map[string]bool{"AddItem": true, "RemoveItem": true, "GetItem": true, "GetAllItems": true}

func main() {
	queue := flag.String("queue", "", "The name of the queue.")
	op := flag.String("op", "", "The name of the operation.")
	key := flag.String("key", "", "The item's key.")
	val := flag.String("val", "", "The value for the item.")
	flag.Parse()
	if *queue == "" {
		log.Fatal("You must specify url of the queue with --queue")
	}
	if _, exists := ops[*op]; !exists {
		log.Fatal("-op must be on of ", keys(ops))
	}
	if *key == "" && *op != "GetAllItems" {
		log.Fatal("-key must be provided")
	}
	if *op == "AddItem" && *val == "" {
		log.Fatal("-val must be provided")
	}
	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = "us-east-1"
	}
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		log.Fatal(err)
	}
	svc := sqs.New(sess)
	_, err = svc.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"OP": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: op,
			},
			"Key": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: key,
			},
		},
		MessageBody: val,
		QueueUrl:    queue,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Done.")
}
