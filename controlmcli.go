package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

func main() {

	// build an http transport
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}

	// let's store our credentials in JSON format
	loginCredentials := strings.NewReader(`{"username": "$USERNAME", "password": "$PASSWORD"}`)

	// the POST request, controlm api login url & credentials are set to req variable.
	req, err := http.NewRequest("POST", "https://$HOST:$PORT/automation-api/session/login", loginCredentials)
	if err != nil {
		log.Fatal(err)
	}
	req.Header = http.Header{
		"Content-Type":    []string{"application/json"},
		"X-Custom-Header": []string{"myvalue"},
	}

	// POST to the controlm api loging URL with all values set
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// wrap the POST ( JSON response ) into a new variable
	post_response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// post_response returns 3 json fields - below struct to capture the values
	type posthold struct {
		Username string
		Token    string
		Version  string
	}

	// this will hold the json struct values
	var structValues posthold

	// unmarhsal structValues values
	err = json.Unmarshal([]byte(post_response), &structValues)
	if err != nil {
		log.Fatal(err)
	}

	// check if jobname given
	if len(os.Args) == 1 {
		color.Red("Please enter a jobname!")
		os.Exit(1)
	}

	color.Yellow("Getting ControlM log : ")
	url1 := "https://$HOST:$PORT/automation-api/run/jobs/status?jobname=" + os.Args[1]

	// GET url1 with token
	response2, err := http.NewRequest("GET", url1, nil)
	if err != nil {
		log.Fatal(err)
	}
	response2.Header = http.Header{
		"Content-Type":    []string{"application/json"},
		"X-Custom-Header": []string{"myvalue"},
		"Authorization":   []string{"Bearer " + structValues.Token},
	}

	// GET response2
	controlmOutput1, err := client.Do(response2)
	if err != nil {
		log.Fatal(err)
	}
	defer controlmOutput1.Body.Close()

	// place controlmOutput1 output and place it into the jobstate
	jobstate, err := ioutil.ReadAll(controlmOutput1.Body)
	if err != nil {
		log.Fatal(err)
	}
	// struct to hold all the JSON values for jobstate. We may need them later
	type JobStruct struct {
		Statuses []struct {
			Jobid          string `json:"jobId"`
			Folderid       string `json:"folderId"`
			Numberofruns   int    `json:"numberOfRuns"`
			Name           string `json:"name"`
			Folder         string `json:"folder"`
			Type           string `json:"type"`
			Status         string `json:"status"`
			Held           bool   `json:"held"`
			Deleted        bool   `json:"deleted"`
			Starttime      string `json:"startTime"`
			Endtime        string `json:"endTime"`
			Orderdate      string `json:"orderDate"`
			Ctm            string `json:"ctm"`
			Description    string `json:"description"`
			Host           string `json:"host"`
			Application    string `json:"application"`
			Subapplication string `json:"subApplication"`
			OutputURI      string `json:"outputURI"`
			LogURI         string `json:"logURI"`
		}
		Returned int `json:"returned"`
		Total    int `json:"total"`
	}

	var p JobStruct
	err = json.Unmarshal([]byte(jobstate), &p)
	if err != nil {
		log.Fatal(err)
	}

	// joblogdata contains the controlm Jobid needed to search the controlmlogs
	joblogdata := p.Statuses[0].Jobid

	// what if we want the output
	url3 := "https://$HOST:$PORT/automation-api/run/job/" + joblogdata + "/output"
	// new GET request using the old token but new url for the logs
	response4, err := http.NewRequest("GET", url3, nil)
	if err != nil {
		log.Fatal(err)
	}
	response4.Header = http.Header{
		"Content-Type":           []string{"application/json"},
		"Annotation-Subject":     []string{"per_dev"},
		"Annotation-Description": []string{"per_dev"},
		"Authorization":          []string{"Bearer " + structValues.Token},
	}

	//
	controlmOutput13, err := client.Do(response4)
	if err != nil {
		log.Fatal(err)
	}
	defer controlmOutput13.Body.Close()

	theLogOutputFromControlm, err := ioutil.ReadAll(controlmOutput13.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", theLogOutputFromControlm)
	// log out of controlm
	logoutPost, err := http.NewRequest("POST", "https://$HOST:$PORT/automation-api/session/logout", nil)
	if err != nil {
		log.Fatal(err)
	}
	logoutPost.Header.Set("Authorization", "Bearer"+structValues.Token)
	logoutStatus, err := client.Do(logoutPost)
	if err != nil {
		log.Fatal(err)
	}
	defer logoutStatus.Body.Close()
}
