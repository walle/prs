// package main implements the functionality of the application `prs`.
// prs uses a github access token and a username to list all open pull requests
// for the given user.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Context and client used to communicate with github's API.
var (
	ctx    context.Context
	client *github.Client
)

func main() {
	token, username, err := validateInput()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx = context.Background()
	client = setupClient(ctx, token)

	issues, err := retrieveIssuesFor(username)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not fetch issues:", err)
		os.Exit(1)
	}

	for _, i := range issues.Issues {
		user, repo := extractUserAndRepo(&i)
		pr, err := fetchPR(user, repo, &i)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not fetch pull request for issue:", err)
		}

		printInfo(os.Stdout, user, repo, &i, pr)
	}
}

// validateInput reads the access token from the environment and the
// username from environment and commandline, commandline takes precedens.
// If either token or username is missing an error suitable for showing
// the user is returned.
func validateInput() (string, string, error) {
	token := os.Getenv("PRS_GITHUB_ACCESS_TOKEN")
	if token == "" {
		return "", "", errors.New("No access token in env | PRS_GITHUB_ACCESS_TOKEN")
	}

	username := os.Getenv("PRS_USERNAME")

	if username == "" {
		if len(os.Args) != 2 {
			return "", "", errors.New("usage: prs [PRS_USERNAME]")
		}
		username = os.Args[1]
	} else {
		if len(os.Args) == 2 {
			username = os.Args[1]
		}
	}

	return token, username, nil
}

// setupClient creates a new client for the github API using the token as
// authorization mechanism.
func setupClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// retrieveIssuesFor returns issue data for all open pull requests that
// relates to `username` in any way.
// Results are ordered by created date, oldest first.
func retrieveIssuesFor(username string) (*github.IssuesSearchResult, error) {
	issues, _, err := client.Search.Issues(ctx,
		fmt.Sprintf("type:pr involves:%s state:open", username),
		&github.SearchOptions{
			Sort:  "created",
			Order: "asc",
		})
	return issues, err
}

// extractUserAndRepo returns the username/org of the repo and it's name.
func extractUserAndRepo(i *github.Issue) (string, string) {
	parts := strings.Split(i.GetURL(), "/")
	return parts[4], parts[5]
}

// fetchPR returns the pull request for issue `i`.
func fetchPR(user, repo string, i *github.Issue) (*github.PullRequest, error) {
	pr, _, err := client.PullRequests.Get(ctx, user, repo, i.GetNumber())
	return pr, err
}

// printInfo collects all the information about the PR and prints it to `w`.
func printInfo(w io.Writer, u, r string, i *github.Issue, pr *github.PullRequest) {
	repoPath := u + "/" + r
	repoInfo := repoPath + " | " + i.GetTitle() + " |"
	prStats := fmt.Sprintf("* %d + %d - %d |", pr.GetChangedFiles(),
		pr.GetAdditions(), pr.GetDeletions())
	prComments := fmt.Sprintf("C %d |", pr.GetComments())
	prCreatedAt := i.GetCreatedAt().Format("2006-01-02") + " |"
	prURL := i.GetURL()

	fmt.Fprintln(w, repoInfo, prStats, prComments, prCreatedAt, prURL)
}
