package config

import (
	"crypto/tls"
	"os"
	"time"

	"log"

	"github.com/jessevdk/go-flags"
)

// CLI Options
type Options struct {
	Port         int           `short:"p" long:"port" description:"Port to listen on" default:"443" env:"BARYON_PORT"`
	Bind         string        `short:"b" long:"bind" description:"Ip address to bind to" default:"0.0.0.0" env:"BARYON_BIND"`
	Secret       string        `short:"s" long:"secret" description:"The web-hook secret" env:"BARYON_HOOK_SECRET"`
	Key          string        `short:"k" long:"key" description:"Specify a Key file to enable server to start TLS" env:"BARYON_KEY"`
	Cert         string        `short:"c" long:"cert" description:"Cert file for TLS" env:"BARYON_CERT"`
	Org          string        `short:"o" long:"org" description:"Github Org to find cookbooks" env:"BARYON_GITHUB_ORG"`
	Token        string        `short:"t" long:"token" description:"Github API token to use when connecting" env:"BARYON_GITHUB_TOKEN"`
	SyncInterval time.Duration `short:"i" long:"interval" description:"Interval to perform full sync against github repos. Supports Golang duration formatting '1h2m'... etc." default:"12h" env:"BARYON_INTERVAL"`
	NoSync       bool          `long:"no-sync" description:"Do NOT perform a github scan/sync when starting. Periodic sync will still fire" env:"BARYON_NOSYNC"`
	BerksOnly    bool          `long:"berks-only" description:"Only use berks compatable version tags in the universe" env:"BARYON_BERKSONLY"`
	TLS          bool
}

// Opts is the application config struct
var Opts Options

func init() {
	_, err := flags.Parse(&Opts)
	if err != nil {
		os.Exit(1)
	}

	// Load client cert
	if Opts.Key != "" {
		Opts.TLS = true
		_, err = tls.LoadX509KeyPair(Opts.Cert, Opts.Key)
		if err != nil {
			log.Fatal("Error loading client keypair: ", err)
		}
	}

	if Opts.Org == "" {
		log.Fatal("Please specify a github org with '-o' or use '-h' to get help")
	}
}
