package gh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/go-chef/metadata-parser"
	"github.com/google/go-github/github"
	"github.com/hashicorp/go-version"
	"github.com/jpillora/backoff"
	"github.com/pantheon-systems/baryon/hook"
	"github.com/pantheon-systems/baryon/universe"
)

const (
	sourceName    = "Github"
	lowerAPIlimit = 1000
	upperAPIlimit = 2000
)

type Source struct {
	client    *github.Client
	Conf      Config
	universe  *universe.Universe
	exBackoff *backoff.Backoff
}

type Config struct {
	Token        string
	MaxVersions  int
	Org          string
	SyncInterval time.Duration
}

type RepoVersion struct {
	Name    string
	Org     string
	Version version.Version
	Tag     string
	URL     string
}

func New(conf Config, universe *universe.Universe) Source {
	// add client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: conf.Token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	bo := &backoff.Backoff{
		Min:    10 * time.Second,
		Max:    10 * time.Minute,
		Factor: 2,
		Jitter: true,
	}

	gs := Source{client: github.NewClient(tc), Conf: conf, universe: universe, exBackoff: bo}

	return gs
}

func (g Source) Name() string {
	return sourceName
}

func (g Source) SyncInterval() time.Duration {
	return g.Conf.SyncInterval
}

// Sync will fetch data bout all cookbooks on the org/
func (g Source) Sync() error {
	log.Println("Syncing cookbooks from", g.Conf.Org)
	opt := &github.RepositoryListByOrgOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var repos []github.Repository
	for {
		g.maybeBackoff()

		r, resp, err := g.client.Repositories.ListByOrg(g.Conf.Org, opt)
		if err != nil {
			g.exBackoff.Duration() // increment exp backoff
			fmt.Printf("error fetching repos: %v\n\n", err)
			return err
		}
		repos = append(repos, r...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	log.Printf("Loaded %d repos from %s\n", len(repos), g.Conf.Org)

	var wg sync.WaitGroup
	wg.Add(len(repos))
	for _, r := range repos {
		go func(r github.Repository) {
			g.processRepoTags(r)
			wg.Done()
		}(r)
	}
	wg.Wait()
	return nil
}

// getTags pull all tags from a repo, and use them for a release where appropriate.
func (g Source) processRepoTags(repo github.Repository) {

	var repotags []github.RepositoryTag

	rv := RepoVersion{Name: *repo.Name, Org: *repo.Owner.Login}
	opts := &github.ListOptions{PerPage: 100}

	log.Printf("fetching tags for %s/%s\n", rv.Org, rv.Name)
	for {
		t, resp, err := g.client.Repositories.ListTags(rv.Org, rv.Name, opts)
		if err != nil {
			log.Printf("error fetching tags for repos: %s\n%s\n", rv.Name, err)
			continue
		}
		repotags = append(repotags, t...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	successCount := 0
	for _, t := range repotags {
		if successCount >= g.Conf.MaxVersions {
			break
		}

		v, err := universe.ParseVersion(*t.Name)
		if err != nil {
			log.Printf("couldn't parse version '%s' for repo '%s', %s", *t.Name, rv.Name, err)
			continue
		}
		rv.Version = *v
		rv.Tag = *t.Name

		if len(rv.Version.Segments()) < 3 {
			log.Printf("version for %s at tag %s is not valid, skipping", rv.Name, *t.Name)
			continue
		}

		g.AddRepoRelease(rv)
		successCount += 1
	}

	return
}

// Process Repo Release manages adding a release to the universe cache
func (g Source) AddRepoRelease(r RepoVersion) {
	log.Printf("Adding: repo=%s:%s version=%s tag=%s", r.Org, r.Name, r.Version.String(), r.Tag)
	g.maybeBackoff()
	opts := &github.RepositoryContentGetOptions{Ref: r.Tag}

	// berkshelf only supports pulling tarballs from github that have a releas prefixed with  a `v` this is strange, and I hope to patch berks. For now
	// we have this switch that will only accept tags berkshelf can read :|
	// TODO:(jnelson) This maybe should be in AddEntry instead of here in the implementation, its a property of the Universe not about the source Sun Sep 20 01:53:04 2015
	if g.universe.Config.BerksOnly == true {

		m, err := regexp.MatchString(`^v`, r.Tag)
		if err != nil {
			log.Printf("Failed checking compatible version tag for repo %s:%s, at %s: %s", r.Org, r.Name, r.Tag, err.Error())
			return
		}

		// no match
		if m == false {
			log.Printf("No Berks compatible version tag for repo %s:%s, at %s", r.Org, r.Name, r.Tag)
			return
		}
	}

	// match the annotated tag first
	// then see if there is metadata.json
	// then fallback to metadata.rb parsing
	// For now this is pulling metadata.rb and parsing it :| We want to do something better
	// TODO:(jnelson) move this out to a func that returns an entry to make way for other methods of resolving deps Wed Aug 26 14:28:22 2015
	log.Printf("Fetching metadata for repo=%s:%s version=%s tag=%s", r.Org, r.Name, r.Version.String(), r.Tag)
	fileContent, err := g.client.Repositories.DownloadContents(r.Org, r.Name, "metadata.rb", opts)
	if err != nil {
		log.Printf("Failed pulling metdata for repo %s:%s, at %s: %s", r.Org, r.Name, r.Tag, err)
		return
	}
	// the call above returns a nil fileContent on error. move it down here to ensure it exists
	defer fileContent.Close()

	buff := new(bytes.Buffer)
	_, err = io.Copy(buff, fileContent)
	if err != nil {
		log.Printf("Error copying response from github, %s", err)
	}

	meta, err := metadata.NewParser(buff).Parse()
	if err != nil {
		log.Printf("Error parsing metadata for %s:%s, at %s: %s", r.Org, r.Name, r.Tag, err)
		return
	}

	if meta.Name == "" {
		log.Printf("Error no name in metadata for %s:%s, at %s: %s", r.Org, r.Name, r.Tag, err)
		return
	}

	if len(meta.Version.Segments()) < 3 {
		log.Printf("Error no version in metadata for %s:%s, at %s: %s", r.Org, r.Name, r.Tag, err)
		return
	}

	entry := universe.Entry{}
	entry.Dependencies = make(map[string]string) // no nil maps mKay
	for _, dep := range meta.Depends {
		entry.Dependencies[dep.Name] = dep.Constraint.String()
	}
	entry.LocationType = "github" // github location type for berks

	// There is a bug I think in the go-github code that is forming the zip/tarball urls wrong (using api instead of web addr)
	// For now we do this by hand. potentially could get this from releases on the repo, but this is a static format
	// https://github.com/pantheon-cookbooks/pipe_tester/archive/v1.1.1.tar.gz
	// This should be t.TarballURL
	entry.DownloadUrl = fmt.Sprintf("https://github.com/%s/%s/archive/%s.tar.gz", r.Org, r.Name, r.Tag)
	entry.LocationPath = entry.DownloadUrl
	// this is used in chef-server as a path to the cookbook
	// entry.LocationPath = fmt.Sprintf("https://github.com/%s/%s/tree/%s", r.Org, r.Name, r.Tag)

	g.universe.AddEntry(universe.ResolvedEntry{sourceName, meta.Name, meta.Version, entry})
}

// eventProcessor runs the loop for watching github events and tossing them over to universe
func (g Source) EventProcessor(events <-chan hook.Event) {
	log.Printf("Starting hook event processor")
	for {
		select {
		case event := <-events:
			log.Printf("Processing event %s for repo %s ", event.Type, event.Repo)
			if event.Type == "create_tag" {
				v, err := universe.ParseVersion(event.Ref)
				if err != nil {
					log.Println("Recived 'create_tag' event with bad version: ", err)
				} else {
					rv := RepoVersion{Name: event.Repo, Org: event.Owner, Version: *v, Tag: event.Ref}
					go g.AddRepoRelease(rv)
				}
			} else {
				log.Println("event isn't a tag create, ignoring")
			}
		}
	}
	log.Printf("Uh-oh we exited the select in event processor. This shouldn't happen")
}

// rateLimit is a debug function for outputing API limits
func (g Source) RateLimit() int {
	rate, _, err := g.client.RateLimit()
	if err != nil {
		fmt.Printf("Error fetching rate limit: %+v\n", err)
		return 0
	} else {
		fmt.Println("API Rate Limit: ", rate)
	}

	return rate.Remaining
}

func (g Source) maybeBackoff() {
	for {
		rl := g.RateLimit()
		if rl <= lowerAPIlimit {
			d := g.exBackoff.Duration()
			log.Printf("Ratelimit request below threshold '%d', waiting for '%s' before retrying ", lowerAPIlimit, d)
			time.Sleep(d)
		} else if rl >= upperAPIlimit {
			log.Printf("Ratelimit request is above threshold '%d', resuming", upperAPIlimit)
			g.exBackoff.Reset()
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

func (g Source) RateHandler(w http.ResponseWriter, r *http.Request) {
	rate, _, err := g.client.RateLimit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(rate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
