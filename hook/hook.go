// this is gutted/extended from https://github.com/phayes/hookserve/
// TODO(jesse): could clean this up alot
package hook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bmatsuo/go-jsontree"
	"github.com/julienschmidt/httprouter"
)

var ErrInvalidEventFormat = errors.New("Unable to parse event string. Invalid Format.")

type Event struct {
	Owner      string // The username of the owner of the repository
	Repo       string // The name of the repository
	Ref        string // The ref on the event
	Branch     string // The branch the event took place on
	Commit     string // The head commit hash attached to the event
	Type       string // Can be either "pull_request", "push", "create_tag", "create_branch"
	BaseOwner  string // For Pull Requests, contains the base owner
	BaseRepo   string // For Pull Requests, contains the base repo
	BaseBranch string // For Pull Requests, contains the base branch
}

func (e *Event) String() (output string) {
	output += "type:   " + e.Type + "\n"
	output += "owner:  " + e.Owner + "\n"
	output += "repo:   " + e.Repo + "\n"
	output += "branch: " + e.Branch + "\n"
	output += "ref:    " + e.Ref + "\n"
	output += "commit: " + e.Commit + "\n"

	if e.Type == "pull_request" {
		output += "bowner: " + e.BaseOwner + "\n"
		output += "brepo:  " + e.BaseRepo + "\n"
		output += "bbranch:" + e.BaseBranch + "\n"
	}

	return
}

type Server struct {
	Port   int        // Port to listen on. Defaults to 80
	Path   string     // Path to receive on. Defaults to "/postreceive"
	Secret string     // Option secret key for authenticating via HMAC
	Events chan Event // Channel of events. Read from this channel to get push events as they happen.
}

// Create a new server with sensible defaults.
// By default the Port is set to 80 and the Path is set to `/postreceive`
func NewServer() *Server {
	return &Server{
		Port:   80,
		Path:   "/postreceive",
		Events: make(chan Event, 10), // buffered to 10 items
	}
}

// Satisfies the extended httpRouter handler interface which has a paramater
func (s *Server) ServeHTTPRouter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s.ServeHTTP(w, r)
}

// Satisfies the http.Handler interface.
// Instead of calling Server.ListenAndServe you can integrate hookserve.Server inside your own http server.
// If you are using hookserve.Server in his way Server.Path should be set to match your mux pattern and Server.Port will be ignored.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != "POST" {
		http.Error(w, "405 Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if req.URL.Path != s.Path {
		http.Error(w, "404 Not found", http.StatusNotFound)
		return
	}

	eventType := req.Header.Get("X-GitHub-Event")
	if eventType == "" {
		http.Error(w, "400 Bad Request - Missing X-GitHub-Event Header", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If we have a Secret set, we should check the MAC
	if s.Secret != "" {
		sig := req.Header.Get("X-Hub-Signature")

		if sig == "" {
			http.Error(w, "403 Forbidden - Missing X-Hub-Signature required for HMAC verification", http.StatusForbidden)
			return
		}

		mac := hmac.New(sha1.New, []byte(s.Secret))
		mac.Write(body)
		expectedMAC := mac.Sum(nil)
		expectedSig := "sha1=" + hex.EncodeToString(expectedMAC)
		if !hmac.Equal([]byte(expectedSig), []byte(sig)) {
			http.Error(w, "403 Forbidden - HMAC verification failed", http.StatusForbidden)
			return
		}
	}

	request := jsontree.New()
	err = request.UnmarshalJSON(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse the request and build the Event
	event := Event{}
	event.Type = eventType
	switch eventType {
	case "push":
		err = pushEvent(request, &event)

	case "pull_request":
		err = prEvent(request, &event)

	case "ping":
		err = pingEvent(request, &event)

	case "create":
		err = createEvent(request, &event)

	default:
		http.Error(w, "Unknown Event Type "+eventType, http.StatusInternalServerError)
		return
	}

	// if we had any error bail
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// We've built our Event - put it into the channel and we're done
	go func() {
		s.Events <- event
	}()

	w.Write([]byte(event.String()))
}

func pingEvent(request *jsontree.JsonTree, event *Event) (err error) {
	return nil
}

// create payload support https://developer.github.com/v3/activity/events/types/#createevent
func createEvent(request *jsontree.JsonTree, event *Event) (err error) {
	t, err := request.Get("ref_type").String()
	if err != nil {
		return fmt.Errorf("unable to subtype of create event: %s", err.Error())
	}
	event.Type = fmt.Sprintf("create_%s", t)
	log.Println("Processing Create event, ", event.Type)

	if event.Ref, err = request.Get("ref").String(); err != nil {
		return fmt.Errorf("unable to find ref for create action: %s", err.Error())
	}

	if event.Repo, err = request.Get("repository").Get("name").String(); err != nil {
		return fmt.Errorf("unable to fetch repo name: %s", err.Error())
	}

	if event.Owner, err = request.Get("repository").Get("owner").Get("name").String(); err != nil {
		if event.Owner, err = request.Get("repository").Get("owner").Get("login").String(); err != nil {
			return fmt.Errorf("unable determine rpo owner: %s", err.Error())
		}
	}

	return
}

func pushEvent(request *jsontree.JsonTree, event *Event) (err error) {
	rawRef, err := request.Get("ref").String()
	if err != nil {
		return
	}
	// If the ref is not a branch, we don't care about it
	if rawRef[:11] != "refs/heads/" || request.Get("head_commit").IsNull() {
		return
	}
	event.Ref = rawRef

	// Fill in values
	event.Branch = rawRef[11:]
	event.Repo, err = request.Get("repository").Get("name").String()
	if err != nil {
		return
	}
	event.Commit, err = request.Get("head_commit").Get("id").String()
	if err != nil {
		return
	}
	event.Owner, err = request.Get("repository").Get("owner").Get("name").String()
	if err != nil {
		return
	}
	return
}

func prEvent(request *jsontree.JsonTree, event *Event) (err error) {
	action, err := request.Get("action").String()
	if err != nil {
		return
	}
	// If the action is not to open or to synchronize we don't care about it
	if action != "synchronize" && action != "opened" {
		return
	}

	event.Owner, err = request.Get("pull_request").Get("head").Get("repo").Get("owner").Get("login").String()
	if err != nil {
		return
	}

	event.Repo, err = request.Get("pull_request").Get("head").Get("repo").Get("name").String()
	if err != nil {
		return
	}
	event.Branch, err = request.Get("pull_request").Get("head").Get("ref").String()
	if err != nil {
		return
	}
	event.Commit, err = request.Get("pull_request").Get("head").Get("sha").String()
	if err != nil {
		return
	}
	event.BaseOwner, err = request.Get("pull_request").Get("base").Get("repo").Get("owner").Get("login").String()
	if err != nil {
		return
	}
	event.BaseRepo, err = request.Get("pull_request").Get("base").Get("repo").Get("name").String()
	if err != nil {
		return
	}
	event.BaseBranch, err = request.Get("pull_request").Get("base").Get("ref").String()
	if err != nil {
		return
	}
	return
}
