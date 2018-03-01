package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/andygrunwald/go-jira"
	"github.com/google/go-github/github"
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
	jiraClient, err := jira.NewClient(nil, "<jira-url>")
	if err != nil {
		panic(err)
	}
	jiraClient.Authentication.SetBasicAuth("usernamem", "password")

	issue, _, err := jiraClient.Issue.Get("IXE-9580", nil)
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
		for i := range cmntList {
			cmnt := cmntList[i]
			fmt.Println(cmnt.Body)
			matchedPubkey, _ := regexp.MatchString("^ssh-rsa\\s\\S*\\s[a-zA-Z0-9-@]*", cmnt.Body)
			if matchedPubkey == true {
				pubKey := cmnt.Body
				fmt.Println("SSH Pub Key:", pubKey)
			} else {
				//shaHash, _ := ioutil.ReadFile(cmnt.Body)
				test := "VdGttKB9EsLdRxZW.bsRXow3Rx2uMhFFwUL4sHFFfgmcKMBRUJGLzXPe6pLErKuIYHhx439OIHuAr1ZuUmSSA."
				data, _ := base64.StdEncoding.DecodeString(test)

				//fmt.Printf("Sha512: %x\n\n", sha512.Sum512(shaHash))
				fmt.Println("sha checksum", data)
			}

		}
	}

	mb := User{"true", "hello", "ssh-rsa", "bash", "present", "nikil"}
	fmt.Println(mb)
	userJSON, _ := json.Marshal(mb)
	fmt.Println(string(userJSON))

	/* Git login */
	tp := github.BasicAuthTransport{
		Username: "git-user",
		Password: "git-token",
	}

	gitClient := github.NewClient(tp.Client())
	ctx := context.Background()
	user, _, err := gitClient.Users.Get(ctx, "")

	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	fmt.Printf("\n%v\n", github.Stringify(user.Login))
	fmt.Printf("\n%v\n", github.Stringify(user.OrganizationsURL))

	// PUT update the main.yml file
	s := github.Stringify(user.Login)
	s = s[1 : len(s)-1]
	//response, status, err := gitClient.Repositories.GetCommit(ctx, s, "s3-registry", "b6fc75e29cd194be301847f879cbcf09c04a7926")
	var commitMsg *string
	str1 := "Commit Msgs"
	commitMsg = &str1

	author := github.CommitAuthor{
		Name:  user.Name,
		Email: user.Email,
		Login: user.Login,
	}

	authorRef := &author
	sha := "55dc6114b38fd123be0b3221c53541de7a73bea7"

	rc := github.RepositoryContentFileOptions{
		Message:   commitMsg,
		Content:   userJSON,
		SHA:       &sha,
		Committer: authorRef,
	}

	//fmt.Println("Status of commit thru api: ", response, status, err)
	var rcp *github.RepositoryContentFileOptions
	rcp = &rc
	response, status, err := gitClient.Repositories.UpdateFile(ctx, s, "s3-registry", "hooks/build", rcp)
	fmt.Println("Commit status", response, status, err)
}

/* Post comment on Jira with commits */
