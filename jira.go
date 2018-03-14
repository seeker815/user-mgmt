package main

import (
	"os"

	jira "github.com/andygrunwald/go-jira"
	log "github.com/sirupsen/logrus"
)

func init() {
	var err error

	jiraClient, err = jira.NewClient(nil, os.Getenv("JIRA_URL"))
	if err != nil {
		panic(err)
	}

	// $ JIRA_USER=xxxxx JIRA_PASS=xxxx ./program

	username, password := os.Getenv("JIRA_USER"), os.Getenv("JIRA_PASS")
	jiraClient.Authentication.SetBasicAuth(username, password)
}
