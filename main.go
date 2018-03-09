package main

import (
	"bufio"
	"context"
	b64 "encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strings"

	jira "github.com/andygrunwald/go-jira"
	yaml2 "github.com/ghodss/yaml"
	github "github.com/google/go-github/github"
	yaml "gopkg.in/yaml.v2"
)

var (
	jiraClient *jira.Client
	hashSHA    string
	pubKey     string
	gitUser    string
	userJSON   []byte
	gitRepo    string
	gitPath    string
)

//User struct to manage user credentials
type User struct {
	Admin          string `yaml:"admin"`
	HashedPassword string `yaml:"hashed_password"`
	Pubkeys        string `yaml:"pubkeys"`
	Shell          string `yaml:"shell"`
	State          string `yaml:"state"`
	Username       string `yaml:"username"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter JIRA ID for user credentials: ")
	ji, _ := reader.ReadString('\n')
	jiraID := strings.TrimSpace(ji)

	readerGit := bufio.NewReader(os.Stdin)
	fmt.Println("Enter the git repository and path separated with a comma")
	gr, _ := readerGit.ReadString('\n')
	st := strings.Split(gr, ",")
	if len(st) == 2 {
		gitRepo = strings.TrimSpace(st[0])
		gitPath = strings.TrimSpace(st[1])
	}

	issue, _, err := jiraClient.Issue.Get(jiraID, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s: %+v\n", issue.Key, issue.Fields.Summary)
	fmt.Printf("Type: %s\n", issue.Fields.Type.Name)
	fmt.Printf("Priority: %s\n", issue.Fields.Priority.Name)
	fmt.Printf("Assignee: %s\n", issue.Fields.Assignee.Name)

	cmntList := issue.Fields.Comments.Comments
	if len(cmntList) == 0 {
		fmt.Println("No Pubkey/password has been entered, nil!")
	} else {
		for _, cmnt := range cmntList {
			fmt.Println(cmnt.Body)

			matchedPubkey, _ := regexp.MatchString("^ssh-rsa\\s\\S*\\s[a-zA-Z0-9-@]*", cmnt.Body)
			if matchedPubkey == true {
				pubKey = "[" + strings.TrimSpace(cmnt.Body) + "]"
				fmt.Println(pubKey)
				continue
			}

			matchedHash, _ := regexp.MatchString("^[0-9a-zA-Z\\S]+$", cmnt.Body)
			if matchedHash == true {
				hashSHA = cmnt.Body
				continue
			}

		}
		mb := User{"true", hashSHA, pubKey, "bash", "present", issue.Fields.Assignee.Name}
		userJSON, err = yaml.Marshal(mb)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return
		}
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

	fc, _, _, err := gitClient.Repositories.GetContents(ctx, gitUser, gitRepo, gitPath, rcg)
	if err != nil {
		fmt.Printf("\nGithub pull failed: %v\n", err)
	}

	sDec, err := b64.StdEncoding.DecodeString(*fc.Content)
	fmt.Println("content", string(sDec))

	sDecs := append(sDec, userJSON...)
	fmt.Println("appended", string(sDecs))

	y, err := yaml2.JSONToYAML(sDec)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	fmt.Println("content in yaml", string(y))
	ya := append(y, userJSON...)
	fmt.Println("appended", string(ya))

	author := &github.CommitAuthor{
		Name:  user.Name,
		Email: user.Email,
		Login: user.Login,
	}

	commitMsg := fmt.Sprintf("Updating user credentials for %s", jiraID)

	rcp := &github.RepositoryContentFileOptions{
		Message:   &commitMsg,
		Content:   y,
		SHA:       fc.SHA,
		Committer: author,
	}

	response, _, err := gitClient.Repositories.UpdateFile(ctx, gitUser, gitRepo, gitPath, rcp)
	if err != nil {
		panic(err)
	}
	fmt.Println("Commit status", response)

}
