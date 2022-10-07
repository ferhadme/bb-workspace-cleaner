package main

import (
	"fmt"
	"bufio"
	"os"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
)

const (
	PASSFILE = "pass.txt"
	BASE_BITBUCKET_API = "https://api.bitbucket.org/2.0"
	BITBUCKET_REPOSITORIES_API = "/repositories/bettechbackend"
	BITBUCKET_REPOSITORY_BRANCH_API = "/%s/refs/branches"
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

func main() {
	pass := initPass()
	repoNames := getRepositories(pass)
	for _, repo := range repoNames {
		branchResponse := getBranchesOfRepository(pass, repo)
		deleteBranchesUpdatedXMonthsAgo(pass, 3, repo, branchResponse)
		break
	}
}

func deleteBranchesUpdatedXMonthsAgo(pass string, x uint8, repo string, branchResponse BranchResponse) {
	for _, branch := range branchResponse.Values {
		dateStr := branch.Target.Date
		lastUpdateTime, err := time.Parse(time.RFC3339, dateStr)

		if err != nil {
			fmt.Println(err)
			return
		}
		now := time.Now()
		fmt.Println(lastUpdateTime)
		fmt.Println(now)
		hoursDiff := now.Sub(lastUpdateTime).Hours()
		fmt.Println(hoursDiff)
		if hoursDiff >= 2190 {
			fmt.Println("Branch is old")
			// TODO: Delete this branch
		}
	}
}

func getBranchesOfRepository(pass string, repo string) BranchResponse {
	client := &http.Client{}

	req, _ := http.NewRequest("GET",
		BASE_BITBUCKET_API + BITBUCKET_REPOSITORIES_API +
			fmt.Sprintf(BITBUCKET_REPOSITORY_BRANCH_API, repo), nil)

	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth("ferhadme", pass)
	q := req.URL.Query()
	q.Add("role", "contributor")
	q.Add("pagelen", "100")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error happened")
		return BranchResponse{}
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var result BranchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Println("Can not unmarshall JSON")
		return BranchResponse{}
	}

	return result
}

func getRepositories(pass string) []string {
	client := &http.Client{}

	req, _ := http.NewRequest("GET",
		BASE_BITBUCKET_API + BITBUCKET_REPOSITORIES_API, nil)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth("ferhadme", pass)

	q := req.URL.Query()
	q.Add("role", "contributor")
	q.Add("pagelen", "50")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error happened")
		return []string {}
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var result RepositoryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Println("Can not unmarshall JSON")
		return []string {}
	}

	repoNames := make([]string, result.Size)
	for idx, repo := range result.Values {
		repoNames[idx] = repo.Name
	}
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
	return fileScanner.Text()
}
