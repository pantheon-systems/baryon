package universe_test

import (
	"log"
	"runtime"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/pantheon-systems/baryon/universe"
)

func checkFatal(t *testing.T, err error) {
	if err == nil {
		return
	}

	// The failure happens at wherever we were called, not here
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatalf("Unable to get caller")
	}
	t.Fatalf("Fail at %v:%v; %v", file, line, err)
}

func newTestUniverse(t *testing.T) *universe.Universe {
	c := universe.Config{
		BerksOnly: true,
	}

	return universe.NewCache(c)
}

func ParseVersion(t *testing.T) {

	for k, v := range map[string]string{
		"v1":              "1.1.0",
		"v1.0":            "1.1.0",
		"v1.0.0":          "1.1.0",
		"v1.0.0-patch":    "1.1.0",
		"v1.0.0+build123": "1.1.0",
	} {
		tv, err := universe.ParseVersion(k)
		if err != nil {
			t.Fatalf("Error processing version for %s : %s", k, err.Error())
		}

		if tv.String() != v {
			t.Fatalf("Expected '%s' got '%s' for input '%s'", v, tv.String(), k)
		}
	}
}

func TestCache(t *testing.T) {
	tv, err := version.NewVersion("1.0.0")
	if err != nil {
		log.Println("BAD VERSION: ", err)
	}

	log.Printf("%#v", tv)

	entry := universe.Entry{
		Dependencies: map[string]string{"http": "~> 1.0.0", "dnf": "1.0.0"},
	}

	rEntry := universe.ResolvedEntry{
		Source:  "test",
		Name:    "test_cook",
		Version: *tv,
		Entry:   entry,
	}

	u := newTestUniverse(t)
	u.AddEntry(rEntry)
	//	TODO(spew.Dump(u)

}
