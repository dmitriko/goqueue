package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog"

	"goqueue/pkg/server"
)

//Appends text received via channel to file
func writeOut(ctx context.Context, filename string, inCh chan string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for {
		select {
		case text := <-inCh:
			if _, err = f.WriteString(text); err != nil {
				panic(err)
			}
			if _, err = f.WriteString("\n"); err != nil {
				panic(err)
			}
		case <-ctx.Done():
			return
		}

	}
}

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	queue := flag.String("queue", "", "The name of the queue.")
	out := flag.String("out", "", "File path to put output.")
	flag.Parse()
	if *queue == "" {
		log.Fatal("You must specify url of the queue with -queue")
	}
	if *out == "" {
		log.Fatal("You must specify file path to output wiht -out")
	}
	storage := server.NewStorage()
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
	outCh := make(chan string)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go writeOut(ctx, *out, outCh)
	storage.Listen(ctx, svc, *queue, outCh)
}
