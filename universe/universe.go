package universe

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/julienschmidt/httprouter"
)

const altVersionRegexRaw string = `^v([\d]+\.?.*)`

// Source is the backend cookbook data source Interface for providing other sources of universe data in a pluggable manner
type Source interface {
	Name() string
	Sync() error
	SyncInterval() time.Duration
}

var altVersionRegex *regexp.Regexp

func init() {
	altVersionRegex = regexp.MustCompile(altVersionRegexRaw)
}

// Universe is the struct that holds the universe data and config

type Universe struct {
	// ugh ruby     cook    version   data
	Universe map[string]map[string]Entry
	Config   Config
	sync.Mutex
}

// Config is a universe config for setting the cachpath/github token and berks compat
type Config struct {
	BerksOnly bool
}

type Entry struct {
	//	EndpointPriority int               `json:"endpoint_priority"`
	// Platforms    map[string]string `json:"platforms"`
	Dependencies map[string]string `json:"dependencies"`
	LocationType string            `json:"location_type"`
	LocationPath string            `json:"location_path"`
	DownloadUrl  string            `json:"download_url"`
	source       string
	name         string
	version      version.Version
}

type ResolvedEntry struct {
	Source  string
	Name    string
	Version version.Version
	Entry   Entry
}

// NewCache sets up a Universe object
func NewCache(conf Config) *Universe {
	u := Universe{}
	u.Config = conf
	u.Universe = make(map[string]map[string]Entry) // prevent nil map

	return &u
}

func (u *Universe) AddEntry(r ResolvedEntry) {
	// we don't want to expose this in the json, but we want to cary this info
	r.Entry.source = r.Source
	r.Entry.version = r.Version
	r.Entry.name = r.Name

	u.Lock()
	defer u.Unlock()

	if u.Universe[r.Name] == nil { // nil maps are bad m'Kay
		u.Universe[r.Name] = make(map[string]Entry)
	} else {
		// we Want to ensure the cook belongs to the same source, no mixed versions
		e := func() (e Entry) { // can't think of a better way to get a random entry
			for _, e := range u.Universe[r.Name] {
				return e
			}
			return
		}()

		// if it isn't same source
		if e.source != "" && e.source != r.Source {
			log.Printf("Conflicts: '%s' from '%s' is not same source as existing '%s'", r.Name, r.Source, e.source)
			return
		}
	}
	u.Universe[r.Name][r.Version.String()] = r.Entry
	log.Println("Added:  ", r)
	return
}

func (r ResolvedEntry) String() string {
	return fmt.Sprintf("Name=%s Version=%s Dependencies=%v LocationType=%s LocationPath=%s DownloadUrl=%s", r.Name, r.Version.String(), r.Entry.Dependencies, r.Entry.LocationType, r.Entry.LocationPath, r.Entry.DownloadUrl)
}

// ParseVersion returns a version object from a parsed string. This normalizes semver strings, and adds the ability to parse strings with 'v' leader. so that `v1.0.1`-> `1.0.1`  which we need for berkshelf to work
func ParseVersion(v string) (*version.Version, error) {
	if altVersionRegex.MatchString(v) {
		m := altVersionRegex.FindStringSubmatch(v)
		if len(m) >= 2 {
			v = m[1]
		}
	}

	nVersion, err := version.NewVersion(v)
	if err != nil {
		// no version available
		return nil, err
	}

	return nVersion, nil
}

// universeHandler responds with the current known universe of cookbook metadata
// https://github.com/chef/chef-rfc/blob/master/rfc0)14-universe-endpoint.md
func (u *Universe) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(u.Universe)
	if err != nil {
		http.Error(w, "error processing universe data", http.StatusInternalServerError)
		log.Println("Error while trying to marshal universe data: ", err)
		return
	}
	w.Write(data)
}

func (u *Universe) CookBookHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	cookbook := ps.ByName("cook")
	data, err := json.Marshal(u.Universe[cookbook])
	if err != nil {
		msg := fmt.Sprintf("error processing universe data for cookbook '%s'", cookbook)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Println(msg, err)
		return
	}
	w.Write(data)
}

func (u Universe) ServiceInfo() map[string]interface{} {
	gCount := 0
	sCount := 0
	total := 0
	for _, data := range u.Universe {
		for _, entry := range data {
			total += 1
			if entry.LocationType == "github" {
				gCount++
			} else {
				sCount++
			}

		}
	}
	extraInfo := make(map[string]interface{})
	extraInfo["Cookbooks"] = total
	extraInfo["Github"] = gCount
	extraInfo["Supermarket"] = sCount

	return extraInfo
}
