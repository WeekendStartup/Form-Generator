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

type Input struct {
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Value       string `json:"value,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Required    bool   `json:"required,omitempty"`

	Options []Option `json:"options,omitempty"`
}

type Option struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

func modelHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		createModel(w, r)
	case "GET":
		getModel(w, r)
	case "DELETE":
		deleteModel(w, r)
	default:
		http.Error(w, "", http.StatusNotImplemented)
		return
	}
}

func createModel(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	model := strings.TrimSpace(r.PostForm.Get("model"))
	if model == "" {
		http.Error(w, "model is required", http.StatusBadRequest)
		return
	}
	var inputs []Input
	err = json.Unmarshal([]byte(model), &inputs)
	if err != nil {
		http.Error(w, "model is not valid json", http.StatusBadRequest)
		return
	}
	modelId, err := generateUUID()
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	gzippedForm, err := gzipBytes([]byte(model))
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{}
	expressionAttributeValues[":modelId"] = &dynamodb.AttributeValue{S: aws.String(modelId)}
	item := map[string]*dynamodb.AttributeValue{}
	item["modelId"] = &dynamodb.AttributeValue{S: aws.String(modelId)}
	item["model"] = &dynamodb.AttributeValue{B: gzippedForm}
	item["createdAt"] = &dynamodb.AttributeValue{S: aws.String(time.Now().UTC().Format(time.RFC3339Nano))}
	req := &dynamodb.PutItemInput{
		TableName:                 aws.String("form-model"),
		ConditionExpression:       aws.String("modelId <> :modelId"),
		ExpressionAttributeValues: expressionAttributeValues,
		Item: item,
	}
	_, err = db.PutItem(req)
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

func getModel(w http.ResponseWriter, r *http.Request) {
	modelId := strings.TrimSpace(r.FormValue("modelId"))
	if modelId == "" {
		http.Error(w, "modelId is required", http.StatusBadRequest)
		return
	}
	params := &dynamodb.GetItemInput{
		TableName: aws.String("form-model"),
		Key: map[string]*dynamodb.AttributeValue{
			"modelId": {S: aws.String(modelId)},
		},
	}
	resp, err := db.GetItem(params)
	if err != nil {
		http.Error(w, "", http.StatusServiceUnavailable)
		return
	}
	if resp.Item["model"] == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	gzippedModel := resp.Item["model"].B
	ungzippedModel, err := ungzipBytes(gzippedModel)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(ungzippedModel)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func deleteModel(w http.ResponseWriter, r *http.Request) {
	modelId := strings.TrimSpace(r.FormValue("modelId"))
	if modelId == "" {
		http.Error(w, "modelId is required", http.StatusBadRequest)
		return
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s://%s/%s=%s", "http", r.Host, "model?modelId", modelId), nil)
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
		TableName: aws.String("form-model"),
		Key: map[string]*dynamodb.AttributeValue{
			"modelId": {S: aws.String(modelId)},
		},
	}
	_, err = db.DeleteItem(params)
	if err != nil {
		http.Error(w, "", http.StatusServiceUnavailable)
		return
	}
}
