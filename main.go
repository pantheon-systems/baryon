package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pantheon-systems/baryon/config"
	"github.com/pantheon-systems/baryon/hook"
	"github.com/pantheon-systems/baryon/source/gh"
	"github.com/pantheon-systems/baryon/source/supermarket"
	"github.com/pantheon-systems/baryon/universe"
	"github.com/wblakecaldwell/profiler"
)

var (
	ghook *hook.Server
)

func main() {
	u := universe.NewCache(universe.Config{
		BerksOnly: config.Opts.BerksOnly,
	})

	// add handlers to help us track memory usage - they don't track memory until they're told to
	profiler.AddMemoryProfilingHandlers()
	profiler.RegisterExtraServiceInfoRetriever(u.ServiceInfo)
	profiler.StartProfiling()

	ghsConfig := gh.Config{
		Token:        config.Opts.Token,
		Org:          config.Opts.Org,
		SyncInterval: config.Opts.SyncInterval,
		MaxVersions:  5,
	}
	ghs := gh.New(ghsConfig, u)
	var wg sync.WaitGroup
	wg.Add(1) // we want to do some stuff after github is done, this is so we can block on the gh repo processing before we do something like add in other sources

	// Add supermarket to the universe after github
	d, _ := time.ParseDuration("6h") // not exposed as a tunable yet
	conf := supermarket.Config{SyncInterval: d}
	sm := supermarket.New(conf, u)

	// ye yea hate the ! no sync, but default behavior should be to sync
	if !config.Opts.NoSync {
		go func() {
			if err := ghs.Sync(); err != nil {
				log.Printf("Error syncing github org %s %s", ghs.Conf.Org, err.Error())
			} else {
				wg.Done()
			}
		}()
	}
	go periodicSync(ghs)
	go periodicSync(sm)

	// create our custom http server/router
	router := httprouter.New()
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Opts.Bind, config.Opts.Port),
		Handler: Log(router),
	}

	ghook = hook.NewServer()
	ghook.Secret = config.Opts.Secret
	ghook.Path = "/hook"

	router.GET("/", rootHandler)

	router.HandlerFunc("GET", "/profiler/info.html", profiler.MemStatsHTMLHandler)
	router.HandlerFunc("GET", "/profiler/info", profiler.ProfilingInfoJSONHandler)
	router.HandlerFunc("GET", "/profiler/start", profiler.StartProfilingHandler)
	router.HandlerFunc("GET", "/profiler/stop", profiler.StopProfilingHandler)

	// github rate info
	router.HandlerFunc("GET", "/rate", ghs.RateHandler)
	// throw these standard handlers in the router
	router.HandlerFunc("GET", "/universe", u.Handler)
	router.HandlerFunc("PUT", "/hook", ghook.ServeHTTP)
	router.HandlerFunc("POST", "/hook", ghook.ServeHTTP)

	// TODO:(jnelson) the EventProcessor is tied to a gh source, but really it should be not a global event stream, but configerd per-org since the gh sources can be multi-org.
	//                Potentially we should setup a WebHook on the Source interface, so each source could have a configured hook/hook path.. Sun Sep 20 03:25:05 2015
	go ghs.EventProcessor(ghook.Events)

	if !config.Opts.NoSync {
		wg.Wait()                         // make sure we have processed everything from github before other sources
		if err := sm.Sync(); err != nil { // sync supermarket
			log.Println("Error syncing supermarket ", err)
		}
	}

	log.Printf("starting server on %s:%d\n", config.Opts.Bind, config.Opts.Port)
	if config.Opts.TLS {
		log.Fatal(server.ListenAndServeTLS(config.Opts.Cert, config.Opts.Key))
	} else {
		log.Fatal(server.ListenAndServe())
	}
}

// periodicSync handles waking and syncing the universe against the org
func periodicSync(s universe.Source) {
	log.Printf("Sync Timer for source %s setup. Duration %v", s.Name(), s.SyncInterval())
	for {
		time.Sleep(s.SyncInterval())
		log.Println("Sync Time arrived! Indexing: ", s.Name())
		s.Sync()
	}
}

// rootHandler presents a simple pingable root
func rootHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "OHAI")
}

// Generic request logging, cause thats probably a good idea
func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ff := r.Header.Get("X-Forwarded-For")
		log.Printf("client='%s' forward-for='%s' method='%s' request='%s' agent='%s'", r.RemoteAddr, ff, r.Method, r.URL, r.UserAgent())
		handler.ServeHTTP(w, r)
	})
}
