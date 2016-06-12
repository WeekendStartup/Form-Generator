package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func contentHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		createContent(w, r)
	case "GET":
		getContent(w, r)
	case "DELETE":
		deleteContent(w, r)
	default:
		http.Error(w, "", http.StatusNotImplemented)
		return
	}
}

func createContent(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	contentStr := strings.TrimSpace(r.PostForm.Get("content"))
	if contentStr == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}
	var content map[string]string
	err = json.Unmarshal([]byte(contentStr), &content)
	if err != nil {
		http.Error(w, "contnet is not valid json", http.StatusInternalServerError)
		return
	}
	if content["modelId"] == "" {
		http.Error(w, "modelId is not provided in the json", http.StatusBadRequest)
		return
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s://%s/%s=%s", "http", r.Host, "model?modelId", content["modelId"]), nil)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	if resp.StatusCode == http.StatusNotFound {
		http.Error(w, "model does not exist", http.StatusBadRequest)
		return
	}
	contentID, err := generateUUID()
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	gzippedContent, err := gzipBytes([]byte(contentStr))
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{}
	expressionAttributeValues[":contentId"] = &dynamodb.AttributeValue{S: aws.String(contentID)}
	item := map[string]*dynamodb.AttributeValue{}
	item["contentId"] = &dynamodb.AttributeValue{S: aws.String(contentID)}
	item["modelId"] = &dynamodb.AttributeValue{S: aws.String(content["modelId"])}
	item["content"] = &dynamodb.AttributeValue{B: gzippedContent}
	item["createdAt"] = &dynamodb.AttributeValue{S: aws.String(time.Now().UTC().Format(time.RFC3339Nano))}
	putItemInput := &dynamodb.PutItemInput{
		TableName:                 aws.String("form-content"),
		ConditionExpression:       aws.String("contentId <> :contentId"),
		ExpressionAttributeValues: expressionAttributeValues,
		Item: item,
	}
	_, err = db.PutItem(putItemInput)
	if err != nil {
		apierr, isapierror := err.(awserr.Error)
		if isapierror && apierr.Code() == "ConditionalCheckFailedException" {
			http.Error(w, "", http.StatusPreconditionFailed)
			return
		}
		http.Error(w, "", http.StatusServiceUnavailable)
		return
	}
}

func getContent(w http.ResponseWriter, r *http.Request) {
	contentID := strings.TrimSpace(r.FormValue("contentId"))
	if contentID == "" {
		http.Error(w, "contentId is required", http.StatusBadRequest)
		return
	}
	params := &dynamodb.GetItemInput{
		TableName: aws.String("form-content"),
		Key: map[string]*dynamodb.AttributeValue{
			"contentId": {S: aws.String(contentID)},
		},
	}
	resp, err := db.GetItem(params)
	if err != nil {
		http.Error(w, "", http.StatusServiceUnavailable)
		return
	}
	if resp.Item["content"] == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	gzippedContent := resp.Item["content"].B
	ungzippedContent, err := ungzipBytes(gzippedContent)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(ungzippedContent)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func deleteContent(w http.ResponseWriter, r *http.Request) {
	contentID := strings.TrimSpace(r.FormValue("contentId"))
	if contentID == "" {
		http.Error(w, "contentId is required", http.StatusBadRequest)
		return
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s://%s/%s=%s", "http", r.Host, "content?contentId", contentID), nil)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	if resp.StatusCode == http.StatusNotFound {
		http.Error(w, "", http.StatusConflict)
		return
	}
	params := &dynamodb.DeleteItemInput{
		TableName: aws.String("form-content"),
		Key: map[string]*dynamodb.AttributeValue{
			"contentId": {S: aws.String(contentID)},
		},
	}
	_, err = db.DeleteItem(params)
	if err != nil {
		http.Error(w, "", http.StatusServiceUnavailable)
		return
	}
}
