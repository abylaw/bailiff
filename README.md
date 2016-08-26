# Bailiff

Brings code into the courtroom for judgement.

0. Get go ;). https://golang.org/doc/install#osx might be the easiest way since brew doesn't make paths easy. 
0. `go get -u github.com/benchlabs/bailiff` - clones repo and builds binary
1. If previous command doesn't work because it doesn't read SSH properly:
     
     ```
     cd $GOPATH/src/github.com/BenchLabs
     git clone git@github.com:BenchLabs/bailiff
     cd bailiff
     go get -u
     go build
     ```
1. Get email from applicant with repo url. Find the cloneable url like `git@github.com:BenchLabs/bailiff.git`
2. `bailiff open <repo url> <name of applicant>`
3. Go to your repository (defined in your config) and see the fresh pull request. 
4. Make insightful and polite comments on the quality of a stranger's code.
5. Delete the insightful and polite comments once they get hired: `bailiff hired <name of applicant>`

The name of the applicant will be the branch name in the target repository, which is defined in the config. 

If you need help:

```
bailiff help
```

### Config
You need a ~/.bailiff.conf.json

```json
{
"owner": "BenchLabs",
"repo": "courtroom",
"githubToken": "ask an admin for this",
"slackChannel": "courtroom",
"slackToken": "create your bot or use @judge's key"
}
```
