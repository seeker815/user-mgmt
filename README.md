# User Management 


## Introduction

A tool that automates the user management from collecting the user's credentials 
to passing those to be managed ansible/Chef

 - Tool is given an input of jira ticket ID with user's credetials and the github repo path

 - The user's public key and hashed password is parsed from Jira ticket and user json 

 - A user json structure is constructed using these details

 - File content is fetched from git repo and user json is appeneded 

 - Tool also adds a comment on jira ticket 


## JIRA/Git credentials 

- Set the JIRA credentials in environment

	> JIRA_USER=xxxxx JIRA_PASS=xxxx JIRA_URL=xxx

- Set the GIT credentials in environment

	> GITHUB_USER=xxxx GITHUB_TOKEN=xxxxx

## Examples

- go get tool and invoke
	> ./user-mgmgt

- Inputs for the tool
	> Enter JIRA ID for user credentials:
  	  IXE-9323

- Enter the git repository and path separated with a comma
  	> repo-name, path/main.yml

- Sample User json
  ```
  - admin: 'true'
  hashed_password: $6$4NEzE8otYL18/6sdfzjvl.MPAFK.GxfDv4D92YYwbrQOX6Vor8ITUhnBgCkd2sQiLx8ADyfN81
  pubkeys: [ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtPkgNOIMJnQjKNO+xhH2eVUNb7G7IUSv01scIekvWj14BHkxMK9SWcCcMUwyAHVD9yAtCscSpweGs50UsvvPZ3Y4IbIEQ90DEnPU2gItIUc7/JGjIVdWdoHrtJBsMiJNtpkVZSsnXxfW2KfUKSxU3XIH4lzTgMJQTNTHiWini3mbzEbK2Hmgf8kNf62rP45m5NTR
      sample@gmail.com]
  shell: bash
  state: present
  username: sample
  ```