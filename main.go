/*
Copyright 2022 Ferhad Mehdizade

Usage:
  ./bb-workspace-cleaner -user <user> -organization <organization> -months <months>
Where,
  -user
    Username of BitBucket account
  -organization
    Name of organization user wants to interact
  -months
    Amount of months repository last updated (default 3)

App password (https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/) should be written to pass.txt file for authentication
*/

package main

import (
	"fmt"
	"bufio"
	"os"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
	"log"
	"flag"
)

const (
	PASSFILE = "pass.txt"
	BASE_BITBUCKET_API = "https://api.bitbucket.org/2.0"
	BITBUCKET_REPOSITORIES_API = "/repositories/%s"
	BITBUCKET_REPOSITORY_BRANCH_API = "/%s/refs/branches"
)

var (
	user = flag.String("user", "", "Username of BitBucket account")
	organization = flag.String("organization", "", "Name of organization user wants to interact")
	months = flag.Int("months", 3, "Amount of months repository last updated")
	pass string
)

type RepositoryResponse struct {
	Size int `json:"size"`
	Values []struct {
		Name string `json:"name"`
	} `json:"values"`
}

type BranchResponse struct {
	Size int `json:"size"`
	Values []struct {
		Name string `json:"name"`
		Target struct {
			Date string `json:"date"`
		} `json:"target"`
	} `json:"values"`
}

func init() {
	pass = initPass()
}

func main() {
	flag.Parse()

	repoNames := getRepositories()

	for _, repo := range repoNames {
		log.Println("Fetching all branches of ", repo)
		branchResponse := getBranchesOfRepository(repo)
		log.Println(branchResponse.Size)
		deleteBranchesUpdatedXMonthsAgo(repo, branchResponse)
		log.Println("******")
	}
}

func deleteBranchesUpdatedXMonthsAgo(repo string, branchResponse BranchResponse) {
	for _, branch := range branchResponse.Values {
		dateStr := branch.Target.Date
		lastUpdateTime, _ := time.Parse(time.RFC3339, dateStr)

		now := time.Now()
		monthsDiff := int(now.Sub(lastUpdateTime).Hours() / 24 / 30)

		if monthsDiff >= *months {
			if branch.Name != "master" || branch.Name != "staging" {
				log.Println("Branch is old ", branch, " and should be deleted")
				deleteBranch(repo, branch.Name)
			}
		}
	}
}

func deleteBranch(repo, branch string) {
	client := &http.Client{}

	req, _ := http.NewRequest("DELETE",
		BASE_BITBUCKET_API +
			fmt.Sprintf(BITBUCKET_REPOSITORIES_API, *organization) +
			fmt.Sprintf(BITBUCKET_REPOSITORY_BRANCH_API, repo) + "/" + branch,
		nil)

	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(*user, pass)
	q := req.URL.Query()
	q.Add("role", "contributor")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error happened while making request")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		log.Println("Branch ", branch, " is successfully deleted")
	} else {
		log.Fatal("Branch deletion failed for branch ", branch)
	}
	log.Println("******")
}

func getBranchesOfRepository(repo string) BranchResponse {
	client := &http.Client{}

	req, _ := http.NewRequest("GET",
		BASE_BITBUCKET_API +
			fmt.Sprintf(BITBUCKET_REPOSITORIES_API, *organization) +
			fmt.Sprintf(BITBUCKET_REPOSITORY_BRANCH_API, repo),
		nil)

	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(*user, pass)
	q := req.URL.Query()
	q.Add("role", "contributor")
	q.Add("pagelen", "100")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error happened while making request")
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var result BranchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Fatal("Can not unmarshall JSON")
		os.Exit(1)
	}

	log.Println("All branches of ", repo, " have been fetched")
	return result
}

func getRepositories() []string {
	client := &http.Client{}

	req, _ := http.NewRequest("GET",
		BASE_BITBUCKET_API +
			fmt.Sprintf(BITBUCKET_REPOSITORIES_API, *organization),
		nil)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(*user, pass)

	q := req.URL.Query()
	q.Add("role", "contributor")
	q.Add("pagelen", "50")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error happened while making request")
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var result RepositoryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Fatal("Can not unmarshall JSON")
		os.Exit(1)
	}

	repoNames := make([]string, result.Size)
	for idx, repo := range result.Values {
		repoNames[idx] = repo.Name
	}
	log.Println("All repositories of organization", *organization, "that user has access fetched")
	return repoNames
}

func initPass() string {
	passFile, err := os.Open(PASSFILE)
	if err != nil {
		fmt.Println(err)
	}
	defer passFile.Close()
	fileScanner := bufio.NewScanner(passFile)
	fileScanner.Split(bufio.ScanLines)
	fileScanner.Scan()
	pass := fileScanner.Text()
	log.Println("Pass is initialized")
	return pass
}
