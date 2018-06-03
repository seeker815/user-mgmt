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
	github "github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

var (
	jiraClient *jira.Client
	hashSHA    string
	pubKey     []string
	gitUser    string
	userJSON   []byte
	gitRepo    string
	gitPath    string
	mb         User
	xc         *github.RepositoryContent
)

//User struct to manage user credentials
type User struct {
	Admin          string   `yaml:"admin"`
	HashedPassword string   `yaml:"hashed_password"`
	Pubkeys        []string `yaml:"pubkeys"`
	Shell          string   `yaml:"shell"`
	State          string   `yaml:"state"`
	Username       string   `yaml:"username"`
}

// createUserCred parses jira comments for Pub key and Password hash
func createUserCred(cmntList []*jira.Comment) ([]string, string) {
	for _, cmnt := range cmntList {
		matchedPubkey, _ := regexp.MatchString("^ssh-rsa\\s\\S*\\s[a-zA-Z0-9-@]*", cmnt.Body)
		if matchedPubkey == true {
			pubKey = []string{strings.TrimSpace(cmnt.Body)}
			fmt.Println(pubKey)
			continue
		}

		matchedHash, _ := regexp.MatchString("^[0-9a-zA-Z\\S]+$", cmnt.Body)
		if matchedHash == true {
			hashSHA = cmnt.Body
			fmt.Println(hashSHA)
			continue
		}

	}
	return pubKey, hashSHA
}

// gitFetch fetches git content/file for the repo path
func gitFetch(repo string, filePath string) (*github.RepositoryContent, error) {
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
	log.Info("\n%v\n", github.Stringify(user.Login))
	log.Info("\n%v\n", github.Stringify(user.OrganizationsURL))
	s := github.Stringify(user.Login)
	gitUser := s[1 : len(s)-1]
	rcg := &github.RepositoryContentGetOptions{Ref: "master"}
	fc, _, _, err := gitClient.Repositories.GetContents(ctx, gitUser, repo, filePath, rcg)
	return fc, err
}

func gitPush(repo string, filePath string, jiraID string, fc *github.RepositoryContent, y []byte) (*github.RepositoryContentResponse, error) {
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
	log.Info("\n%v\n", github.Stringify(user.Login))
	log.Info("\n%v\n", github.Stringify(user.OrganizationsURL))
	s := github.Stringify(user.Login)
	gitUser := s[1 : len(s)-1]

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

	response, _, err := gitClient.Repositories.UpdateFile(ctx, gitUser, repo, filePath, rcp)
	if err != nil {
		panic(err)
	}
	fmt.Println("Commit status", response)
	return response, err
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)

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
	log.WithFields(log.Fields{
		"Name":     issue.Fields.Type.Name,
		"Summary":  issue.Fields.Summary,
		"key":      issue.Fields.Assignee.Name,
		"Priority": issue.Fields.Priority.Name,
	}).Info("Fetched JIRA attributes")

	if len(issue.Fields.Comments.Comments) == 0 {
		fmt.Println("No Pubkey/password has been entered, nil!")
		return
	}

	pubKey, hashSHA = createUserCred(issue.Fields.Comments.Comments)
	mb = User{"true", hashSHA, pubKey, "bash", "present", issue.Fields.Assignee.Name}
	userJSON, err = yaml.Marshal(mb)
	if err != nil {
		log.Error("Unable to connect to local syslog daemon:", err)
		return
	}

	xc, err = gitFetch(gitRepo, gitPath)
	if err != nil {
		log.Error("Github pull failed", err)
	}

	sDec, err := b64.StdEncoding.DecodeString(*xc.Content)
	originalYaml := struct {
		UserList []User `yaml:"user_list"`
	}{}

	log.Info("content", string(sDec))

	err = yaml.Unmarshal(sDec, &originalYaml)
	if err != nil {
		log.Error("Error unmarshalling yaml: ", err)
	}

	originalYaml.UserList = append(originalYaml.UserList, mb)
	b, _ := yaml.Marshal(originalYaml)

	fmt.Printf("\n content %v\n", string(sDec))
	fmt.Printf("\n new content %v\n", string(b))

	r, err := gitPush(gitRepo, gitPath, jiraID, xc, b)
	if err != nil {
		log.Error("Error in Git Push : ", err)
	}

	log.Info("Git push response", r)
}
