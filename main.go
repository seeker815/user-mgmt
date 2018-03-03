package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	jira "github.com/andygrunwald/go-jira"
	github "github.com/google/go-github/github"
)

var (
	jiraClient *jira.Client
	hashSHA    string
	pubKey     string
	gitUser    string
)

//User struct to manage user credentials
type User struct {
	Admin    string
	Password string
	Pubkey   string
	Shell    string
	State    string
	Username string
}

func main() {
	issue, _, err := jiraClient.Issue.Get("IXE-9575", nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s: %+v\n", issue.Key, issue.Fields.Summary)
	fmt.Printf("Type: %s\n", issue.Fields.Type.Name)
	fmt.Printf("Priority: %s\n", issue.Fields.Priority.Name)
	fmt.Printf("Assignee: %s\n", issue.Fields.Assignee.Name)

	// get jira comments (slice of struct pointers)
	cmntList := issue.Fields.Comments.Comments

	if len(cmntList) == 0 {
		fmt.Println("No Pubkey/password has been entered, nil!")
	} else {
		for _, cmnt := range cmntList {
			fmt.Println(cmnt.Body)

			matchedPubkey, _ := regexp.MatchString("^ssh-rsa\\s\\S*\\s[a-zA-Z0-9-@]*", cmnt.Body)
			if matchedPubkey == true {
				pubKey = cmnt.Body
				continue
			}

			matchedHash, _ := regexp.MatchString("^[0-9a-zA-Z\\S]+$", cmnt.Body)
			if matchedHash == true {
				hashSHA = cmnt.Body
				continue
			}

		}
		mb := User{"true", hashSHA, pubKey, "bash", "present", issue.Fields.Assignee.Name}
		fmt.Println(mb)
		userJSON, _ := json.Marshal(mb)
		fmt.Println(string(userJSON))
	}

	username, token := os.Getenv("GITHUB_USER"), os.Getenv("GITHUB_TOKEN")
	tp := github.BasicAuthTransport{
		Username: username,
		Password: token,
	}

	gitClient := github.NewClient(tp.Client())
	ctx := context.Background()
	user, _, err := gitClient.Users.Get(ctx, "")

	if err != nil {
		panic(err)
	}

	fmt.Printf("\n%v\n", github.Stringify(user.Login))
	fmt.Printf("\n%v\n", github.Stringify(user.OrganizationsURL))
	s := github.Stringify(user.Login)
	gitUser := s[1 : len(s)-1]

	rcg := &github.RepositoryContentGetOptions{Ref: "master"}

	// git pull
	fc, rc, resp, err := gitClient.Repositories.GetContents(ctx, gitUser, "s3-registry", "hooks/build", rcg)
	if err != nil {
		fmt.Printf("\nGithub pull failed: %v\n", err)
	}

	fmt.Println("Git response", fc, rc, resp)

	// git PUT

}
