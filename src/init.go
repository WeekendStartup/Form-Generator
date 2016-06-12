package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var db *dynamodb.DynamoDB

func init() {
	if os.Getenv("AWS_REGION") == "" {
		log.Fatalln("AWS_REGION not set")
	}
	db = dynamodb.New(session.New())
	if db == nil {
		log.Fatalln("could not initialize dynamodb client")
	}
}
