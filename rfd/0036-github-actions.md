
---
author: jane quintero (jane@goteleport.com)   
state: implemented
---


# Bot 

## What 

This RFD proposes the implementation of using Github Actions to better manage the Teleport repository's pull requests. The first iteration will include:  
- Auto assigning reviewers to pull requests. 
- Checking approvals for pull requests. 
- Dismissing stale workflow runs. 

## Why 

To improve the speed and quality of the current development workflow process.

## Implementation 

### Getting Data 

Pull request metadata will be obtained via the a event payload file that is already on the Github actions runner. Github actions sets an environment variable, `GITHUB_EVENT_PATH`, which is the path to that event payload in JSON format. With this, the pull request metadata will be unmarshaled into a `PullRequestMetadata` struct and will be used to make the necessary API calls. 

```go
  // Example PullRequestMetadata struct
type PullRequestMetadata struct {
	Author     string
	RepoName   string
	RepoOwner  string
	Number     int
	HeadSHA    string
	BaseSHA    string
	Reviewer   string
	BranchName string
}
```


### Workflows
### Assigning Reviewers 

Reviewers will be assigned when a pull request is opened, marked ready for review, or reopened. 

```yaml
# Example workflow configuration 
name: Assign
on: 
  # Types of events the workflows will trigger on
  pull_request_target:
    types: [assigned, opened, reopened, ready_for_review]

permissions:  
    pull-requests: write
    actions: none
    checks: none
    contents: none
    deployments: none
    issues: none
    packages: none
    repository-projects: none
    security-events: none
    statuses: none

jobs:
  auto-request-review:
    name: Auto Request Review
    runs-on: ubuntu-latest
    steps:
      # Check out the master branch of the Teleport repository. This is to prevent an
      # attacker from submitting their review assignment logic.
      - name: Checkout master branch
        uses: actions/checkout@master        
      - name: Installing the latest version of Go.
        uses: actions/setup-go@v2
      # Run "assign-reviewers" subcommand on bot.
      - name: Assigning reviewers 
        run: go run cmd/bot.go --token=${{ secrets.GITHUB_TOKEN }}  --reviewers=${{ secrets.reviewers }} assign-reviewers

```

### Checking Reviews 

Every time pull request and pull request review events occur, the bot will check if all the required reviewers have approved. 

```yaml
# Example Check workflow
name: Check
on: 
  # Types of events the workflows will trigger on
  pull_request_review:
    type: [submitted, edited, dismissed]
  pull_request_target: 
    types: [assigned, opened, reopened, ready_for_review, synchronize]

permissions:  
    actions: write
    pull-requests: write
    checks: none
    contents: none
    deployments: none
    issues: none
    packages: none
    repository-projects: none
    security-events: none
    statuses: none

jobs: 
  check-reviews:
    name: Checking reviewers 
    runs-on: ubuntu-latest
    steps:
      # Check out the master branch of the Teleport repository. This is to prevent an
      # attacker from submitting their review assignment logic. 
      - name: Check out the master branch 
        uses: actions/checkout@master
      - name: Installing the latest version of Go.
        uses: actions/setup-go@v2
        # Run "check-reviewers" subcommand on bot.
      - name: Checking reviewers
        run: go run cmd/bot.go --token=${{ secrets.GITHUB_TOKEN }}  --reviewers=${{ secrets.reviewers }} check-reviewers
```

#### Secrets 

To know which reviewers to assign and check for, a hardcoded JSON object will be used as a Github secret. Usernames will be the name of the key and the value will be a list of required reviewers' usernames. 
A wildcard key will be required to assign reviewers to external contributor's PRs. 

```json
    // Example `reviewers` secret
    {
        "author1": ["reviewer0", "reviewer1"],
        "author2": ["reviewer2", "reviewer3", "reviewer4"],
        "*": ["reviewer5", "reviewer6"]
    }
```

### Dismissing Stale Runs 

This workflow dismisses stale workflow runs every 30 minutes for every open pull request in the Teleport repository. There is a separate workflow for this because when a review event occurs on an external contributor's PR, the token in that context does not have the correct permissions. 

```yaml
  #  Example dismissing stale runs workflow 
  name: Dismiss Stale Workflows Runs
on:
  schedule:
    - cron:  '0,30 * * * *'
     
permissions: 
  actions: write 
  pull-requests: read
  checks: none
  contents: none
  deployments: none
  issues: none
  packages: none
  repository-projects: none
  security-events: none
  statuses: none

jobs: 
  dismiss-stale-runs:
    name: Dismiss Stale Workflow Runs
    runs-on: ubuntu-latest
    steps:
      - name: Check out the master branch 
        uses: actions/checkout@master
      - name: Installing the latest version of Go.
        uses: actions/setup-go@v2
        # Run "dismiss-runs" subcommand on bot.
      - name: Dismiss
        run: cd .github/workflows/teleport-ci && go run cmd/bot.go --token=${{ secrets.GITHUB_TOKEN }} dismiss-runs

```
### Authentication & Permissions

For authentication, Github Actions provides a token to use in workflow, saved as `GITHUB_TOKEN` in the `secrets` context, to authenticate on behalf of Github actions. The token expires when the job is finished. 

### Bot Edits and Failures 

The [CODEOWNERS](https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/creating-a-repository-on-github/about-code-owners) feature will be used to assign reviewers who can approve edits to the `.github/workflows` directory if there is a change to `.github/workflows`.

__CODEOWNERS need to approve these changes before the edits get merged.__ 

In the event that the bot has a bug or fails, repository admins can override the failed checks and force merge any changes. 
### Security 

To prevent edits to the contents of the workflow directory after CODEOWNERS have approved for external contributors, we need to invalidate approvals for the following commits. This can be done via hitting the [dismiss a review](https://docs.github.com/en/rest/reference/pulls#dismiss-a-review-for-a-pull-request) endpoint for all reviews in an approved state. This will occur on the `synchronize` (commit push to a pull request) event type. 



### Vulnerabilities 

#### [Bypassing a Required Reviewer](https://medium.com/cider-sec/bypassing-required-reviews-using-github-actions-6e1b29135cc7)

The Github actions token is granted write permissions to pull requests and with this, an attacker can run a command that approves the pull request with malicious code. An attacker could only merge this pull request if the repository settings didn't require 2+ reviewers or if `CODEOWNERS` wasn't exhaustively set. The repo this bot manages has both of these attributes. `CODEOWNERS` cover all paths of the repository and each pull request requires 3 reviewers before it can be merged. 

#### User Obtains the Github Token

The maximum permissions on any given workflow are `pull-requests:write` and `actions:write`. If an attacker were to somehow obtain the Github token, the following is what they would have access to: 

- [Permissions](https://docs.github.com/en/rest/reference/permissions-required-for-github-apps#permission-on-actions) on Actions
- [Permissions](https://docs.github.com/en/rest/reference/permissions-required-for-github-apps#permission-on-pull-requests) on Pull Requests

Concerning actions an attacker can perform:  

| Action     | Scenario/Mitigation  |
| ----------- | ----------- |
| Cancels a run     |   An attacker could get all approvals on a pull request, trigger a `pull_request_target` event (such as `synchronize` which is a pushed commit to PR), and cancel a run in that commit with malicious code. If only internal contributors/CODEOWNERS have the ability to merge a PR from a fork, the security in place should be enough.     |
| Delete/edit comments on issues or pull requests.   | An attacker could edit or delete important comments. There doesn't seem to be a way to get a malicious commit in master this way.  |
| Delete logs of a workflow. | An attacker could delete a logs of a workflow that would otherwise fail in their favor. This alone wouldn't be enough to get a malicious commit in for external contributors if the internal contributor that merges their pull request in checks that `Assign` and `Check` workflows pass. Because we do not invalidate reviews for internal contributors when a new commit is pushed and they have the ability to merge their own pull request in, it could be possible to get a malicious commit in. |
| Re-run a workflow.  | An attacker could re-run a workflow though there wouldn't be any benefit to them even if code was changes. Workflow would just run against the new code and pass/fail accordingly. | 
| Update an issue or a pull request. | An attacker could update the contents of a pull request or issue.  There doesn't seem to be a way to get a malicious commit in master this way.| 



#### Github Contexts 

The bot utilizes the encrypted secret context ([more on contexts](https://docs.github.com/en/actions/security-guides/encrypted-secrets)) and the secret passed to the bot contains a JSON encoded string that maps authors to their required reviewers. While this secret can't be used to access any resources, it can be exposed. If an attacker were to somehow change this secret to change the authors or map themselves to an empty list ([see secrets](#secrets)), again `CODEOWNERS` is exhaustive and will require an approval in the event this happens, therefore a malicious commit/pull request could not be merged. 

Note: The docs for security hardening for Github actions strongly recommends structured data is not used as a secret (including JSON). The bot isn't using any actual secrets and it seems ok this case. 

### Dependencies
This bot will use the [go-github](https://github.com/google/go-github) client library to access the Github API to assign, check reviewers, and dismiss stale workflow runs. 


### Scenarios Listed

Internal contributors:

- Reviewers will be assigned when a pull request event triggers.
- PR can't be merged until it has all required approvals.
- Each review event will trigger the check workflow and the PR will be checked for approvals.
- Each time the `Check` workflow triggers, stale workflow runs will be dismissed.
- New commits will not invalidate approvals.

External contributors:

- Reviewers will be assigned when a pull request event triggers.
- PR can't be merged until it has all required approvals.
- Each review event will trigger the check workflow and the PR will be checked for approvals.
- A cron job will invalidate stale `Check` workflow runs every 30 minutes.
- New commits *will* invalidate approvals and reviewers will be tagged in a comment on the pull request to re-review upon invalidation.

