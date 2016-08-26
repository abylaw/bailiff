package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/google/go-github/github"
	"github.com/nlopes/slack"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func gitRun(args ...string) error {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		log.Println(string(output))
	}
	return err
}

func postReviewMessageOnSlack(caseNumber string, defendantName string) (string, string, error) {
	api := slack.New(Config.SlackToken)
	params := slack.PostMessageParameters{}
	courtroomRepo := "https://github.com/" + Config.Owner + "/" + Config.Repo + "/pull/" + caseNumber

	attachment := slack.Attachment{
		Pretext: "Case " + caseNumber,
		Text: fmt.Sprintf("Defendant's Name: *%s*\nCase: %s", defendantName, courtroomRepo),
		MarkdownIn: []string{"text", "pretext"},
	}
	params.Attachments = []slack.Attachment{attachment}
	params.Markdown = true
	params.AsUser = true
	msg := fmt.Sprintf(`@channel All rise! Members of the jury, you are instructed to review %s's case.
	As jurors you are not to be swayed by sympathy. Be insightful though.`, defendantName)

	channelID, timestamp, err := api.PostMessage(Config.SlackChannel, msg, params)
	
	return channelID, timestamp, err
}

func open(c *cli.Context) {
	var err error
	courtroomRepo := "git@github.com:" + Config.Owner + "/" + Config.Repo + ".git"
	branch := c.GlobalString("branch")

	targetRepo := c.Args().Get(0)
	defendant := c.Args().Get(1)
	if targetRepo == "" || defendant == "" {
		log.Fatal("USAGE: courtroom open [repo url] [defendant]")
	}

	if strings.Contains(defendant, " ") {
		log.Fatal("The defendant's name should also be a valid branch name. Sorry.")
	}

	tempDir, err := ioutil.TempDir("/tmp", "courtroom-"+defendant)
	if err != nil {
		log.Fatal("Could not open temp directory")
	}

	log.Println("Cloning", targetRepo, "branch", branch, "into", tempDir)
	if gitRun("clone", "-b", branch, targetRepo, tempDir) != nil {
		log.Fatal("Could not clone repository into ", tempDir)
	}

	if os.Chdir(tempDir) != nil {
		log.Fatal("Could not enter cloned repository ", tempDir)
	}

	if gitRun("remote", "add", "courtroom", courtroomRepo) != nil {
		log.Fatal("Could not add remote")
	}

	if gitRun("fetch", "courtroom") != nil {
		log.Fatal("Could not fetch target repo ", courtroomRepo)
	}

	if gitRun("merge", "courtroom/master") != nil {
		log.Fatal("Could not rebase from master branch of ", courtroomRepo)
	}

	if gitRun("push", "courtroom", branch+":"+defendant) != nil {
		log.Fatal("Could not add branch ", defendant, " to ", courtroomRepo)
	}

	log.Println("Added branch ", defendant, " to ", courtroomRepo)

	canonicalRepoName := Config.Owner + "/" + Config.Repo

	title := defendant + "'s case"
	head := defendant
	base := "master"
	body := "Auto-generated pull request from benchlabs/bailiff. For code review purposes only - do not merge."
	pullRequest := github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &base,
		Body:  &body,
	}
	pr, _, err := GithubClient.PullRequests.Create(Config.Owner, Config.Repo, &pullRequest)
	if err != nil {
		log.Fatal("Could not create new pull request in ", canonicalRepoName, err)
	}
	log.Println("Created new pull request for branch", defendant, "in", canonicalRepoName)

	channelID, timestamp, err := postReviewMessageOnSlack(strconv.Itoa(*pr.Number), defendant)
	if err != nil {
		log.Fatal("Could not post review message on Slack ", canonicalRepoName, err)
	}
	log.Println("Message successfully posted to channel %s at %s", channelID, timestamp)
}

func hired(c *cli.Context) {
	defendant := c.Args().Get(0)
	if defendant == "" {
		log.Fatal("USAGE: bailiff hired [defendant]")
	}

	opt := &github.PullRequestListOptions{
		Head: Config.Owner + ":" + defendant,
	}
	pullRequests, _, err := GithubClient.PullRequests.List(Config.Owner, Config.Repo, opt)
	if len(pullRequests) == 0 || err != nil {
		log.Fatal("Could not list pull requests: ", err)
	}

	prNumber := *pullRequests[0].Number
	log.Println("Found pull request", prNumber, "for defendant", defendant)

	comments, _, err := GithubClient.Issues.ListComments(Config.Owner, Config.Repo, prNumber, nil)
	if err != nil {
		log.Fatal("Could not list comments for issue", prNumber, err)
	}
	for i := 0; i < len(comments); i++ {
		_, err = GithubClient.Issues.DeleteComment(Config.Owner, Config.Repo, *comments[i].ID)
		if err != nil {
			log.Fatal("Could not delete comment", err)
		}
	}
	log.Println("Deleted comments in issue", prNumber)

	prComments, _, err := GithubClient.PullRequests.ListComments(Config.Owner, Config.Repo, prNumber, nil)
	if err != nil {
		log.Fatal("Could not list comments for pull request", prNumber, err)
	}

	for i := 0; i < len(prComments); i++ {
		_, err = GithubClient.PullRequests.DeleteComment(Config.Owner, Config.Repo, *prComments[i].ID)
		if err != nil {
			log.Fatal("Could not delete comment", err)

		}
	}
	log.Println("Deleted comments in pull request", prNumber)

	log.Println("Deleted all comments for defendant", defendant, "in pull request", prNumber)

	closed := "closed"
	thisPr := &pullRequests[0]
	thisPr.State = &closed
	_, _, err = GithubClient.PullRequests.Edit(Config.Owner, Config.Repo, prNumber, thisPr)
	if err != nil {
		log.Fatal("Could not close pull request", prNumber, err)
	}
	log.Print("Closed pull request")

	branch := "heads/" + defendant
	_, err = GithubClient.Git.DeleteRef(Config.Owner, Config.Repo, branch)
	if err != nil {
		log.Fatal("Could not delete branch", defendant, err)
	}
	log.Print("Deleted branch ", defendant)
}

func main() {
	app := cli.NewApp()
	app.Name = "bailiff"
	app.Usage = "Send a defendant to BenchLabs/courtroom"
	app.Authors = []cli.Author{
		{Name: "jeffling", Email: "jeff@bench.co"},
		{Name: "abylaw", Email: "andrea@bench.co"},
	}
	app.Commands = []cli.Command{
		{
			Name:    "open",
			Aliases: []string{"o"},
			Usage:   "[targetRepo] [defendant] - Open a new case",
			Action:  open,
		},
		{
			Name:    "hired",
			Aliases: []string{"hi"},
			Usage:   "[defendant] - Delete all comments on defendant's case",
			Action:  hired,
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "branch, b",
			Value: "master",
			Usage: "Defendant's branch to be pulled into court",
		},
	}
	app.Run(os.Args)
}
