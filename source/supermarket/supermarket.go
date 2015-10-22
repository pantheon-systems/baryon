package supermarket

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/pantheon-systems/baryon/universe"
)

const sourceName = "Supermarket"

type Source struct {
	Conf     Config
	universe *universe.Universe
}

type Config struct {
	SyncInterval time.Duration
}

func New(conf Config, universe *universe.Universe) Source {
	return Source{conf, universe}
}

func (s Source) Name() string {
	return sourceName
}

func (s Source) SyncInterval() time.Duration {
	return s.Conf.SyncInterval
}

func (s Source) Sync() error {
	url := "https://supermarket.chef.io/universe"
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var entries map[string]map[string]universe.Entry
	err = json.Unmarshal(body, &entries)
	if err != nil {
		fmt.Printf("%T\n%s\n%#v\n", err, err, err)
		switch v := err.(type) {
		case *json.SyntaxError:
			fmt.Println(string(body[v.Offset-40 : v.Offset]))
		}
	}

	log.Printf("%d cookbooks loaded from supermarket", len(entries))

	for name, data := range entries {
		log.Printf("Processing supermarket cookbooks: '%s'", name)
		for ver, entry := range data {
			nVersion, err := version.NewVersion(ver)
			if err != nil {
				log.Printf("Failed to add  '%s@%s': %s", name, ver, err.Error())
				continue
			}

			e := universe.ResolvedEntry{
				Source:  sourceName,
				Name:    name,
				Version: *nVersion,
				Entry:   entry,
			}
			s.universe.AddEntry(e)
		}
	}

	return nil
}
