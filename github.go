package main

import (
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var GithubClient *github.Client

// create struct for the token source
type tokenSource struct {
	token *oauth2.Token
}

// add Token() method to satisfy oauth2.TokenSource interface
func (t *tokenSource) Token() (*oauth2.Token, error) {
	return t.token, nil
}

func init() {
	ts := &tokenSource{
		&oauth2.Token{AccessToken: Config.GithubToken},
	}

	tc := oauth2.NewClient(oauth2.NoContext, ts)
	GithubClient = github.NewClient(tc)
}
