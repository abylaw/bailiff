package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os/user"
	"path"
)

var Config *config

type config struct {
	GithubToken 	string `json "githubToken"`
	Owner       	string `json "owner"`
	Repo        	string `json "repo"`
	SlackChannel	string `json "slackChannel`
	SlackToken	string `json "slackToken"`
}

func init() {
	// get config
	user, err := user.Current()
	if err != nil {
		log.Fatal("Could not get current user")
	}

	configPath := path.Join(user.HomeDir, ".bailiff.conf.json")
	configString, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal("Could not find config file at ", configPath, err)
	}
	Config = &config{}
	if err := json.Unmarshal(configString, &Config); err != nil {
		log.Fatal("Error parsing config: ", err)
	}
}
