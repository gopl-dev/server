package handler

import (
	"log"
	"net/http"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/content"
	"github.com/gopl-dev/server/github"
)

const (
	GitHubPingEvent              = "ping"
	GitHubPushEvent              = "push"
	GitHubGitHubPullRequestEvent = "pull_request"
)

var (
	ErrInvalidEvent = app.NewError(http.StatusBadRequest, "invalid event")
)

type GitHubWebHookHeaders struct {
	// Name of the event that triggered the delivery
	EventName string `header:"X-GitHub-Event"`
	// A GUID to identify the delivery.
	EventID string `header:"X-GitHub-Delivery"`

	// The HMAC hex digest of the response body.
	// This header will be sent if the webhook is configured with a secret.
	// The HMAC hex digest is generated using the sha1 hash function and the secret as the HMAC key.
	EventSig string `header:"X-Hub-Signature"`
}

type GitHubRepoPushed struct {
	Repo GitHubRepo `json:"repository"`
}

type GitHubRepo struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	CloneURL      string `json:"clone_url"`
	Path          string `json:"full_name"`
	Private       bool   `json:"private"`
	DefaultBranch string `json:"default_branch"`
	HTMLURL       string `json:"html_url"`
}

// ImportContentFromGitHubRepo is a webhook's handler that is triggered by GitHub
// Once commits pushed to a branch, GitHub will send request to a route which will call this method
// (Trigger must set manually on GitHub)
// Here we simply check for valid  signature, then clone repo to have data locally and then import it.
// Note: Payloads are capped at 25 MB. If your event generates a larger payload, a webhook will not be fired. This may happen, for example, on a create event if many branches or tags are pushed at once. We suggest monitoring your payload size to ensure delivery.
// Note: You will not receive a webhook for this event when you push more than three tags at once.
// https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#push
func ImportContentFromGitHubRepo(w http.ResponseWriter, r *http.Request) {
	h := NewHandler(r, w)

	// copy request body now, it'll need soon to check signature
	body, err := copyRequestBody(r)
	if err != nil {
		h.Abort(err)
		return
	}

	var req GitHubRepoPushed
	var headers GitHubWebHookHeaders
	h.bindJSON(req)
	if h.Aborted() {
		return
	}
	h.MapHeaders(headers)

	// event must be a ping or push
	if headers.EventName != GitHubPushEvent && headers.EventName != GitHubPingEvent {
		h.Abort(ErrInvalidEvent)
		return
	}

	repo, err := service.FindGitHubRepoByPath(req.Repo.Path)
	if err != nil {
		h.Abort(err)
		return
	}

	validSig, err := github.IsValidHMAC(body, headers.EventSig, repo.Secret)
	if err != nil {
		h.Abort(app.ErrBadRequest("failed to validate signature"))
		return
	}
	if !validSig {
		h.Abort(app.ErrBadRequest("invalid signature"))
		return
	}

	// import data on goroutine, because repo is authorized, but might take some time to import
	// TODO: make selective import using diff ?
	// TODO import using queue
	go func() {
		repo.CloneURL = req.Repo.CloneURL
		err = content.ImportFromGitHub(repo)
		if err != nil {
			log.Printf("[ERROR] import from GH: " + err.Error())
		}
	}()

	h.jsonOK("ok")
}
