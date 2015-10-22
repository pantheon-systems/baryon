[![Circle CI](https://circleci.com/gh/pantheon-systems/baryon/tree/master.svg?style=svg&circle-token=92ed13ff052f213b007977e8ef356831b9c63e0d)](https://circleci.com/gh/pantheon-systems/baryon/tree/master)

Baryon
-----
Chef cookbook compositor (aka universe endpoint)

This is an implementation of the Chef Universe server API that layers and combines multiple universe sources like the Chef Supermarket and GitHub organizations.

By default Baryon makes use of GitHub releases (tags) as the primary source for cookbook version information. Any remaining dependencies are resolved with the community Supermarket data. This provides a simple mechanism to combine private version controlled storage with public community cookbooks, with preference for the internal name-space.

A GitHub web-hook endpoint for processing cookbook version data (typically automatic tagging during continuous integration) as well as polling GitHub repositories on a sync interval is provided.

Features:
  * Exclusive Merge of sources (Github org before supermarket universe)
  * GitHub hook Universe updates
  * understands tags beyond /^v\d.\d.\d/
  * Efficient usage of Github api
  * Stable service

## Getting Started
Pull a release from releases for your architecture.


## How to use
 The most simple setup is to just `./baryon -p 80 -o mygithuborg -t mytoken -s hooksecret` this starts a server that will listen on port 80, and index 'mygithuborg' using your github token and waiting for github hook payloads with 'hooksecret' as githubs auth request to baryon.

```
  ./baryon --help
Usage:
  baryon [OPTIONS]

Application Options:
  -p, --port=       Port to listen on (443) [$BARYON_PORT]
  -b, --bind=       Ip address to bind to (0.0.0.0) [$BARYON_BIND]
  -s, --secret=     The web-hook secret [$BARYON_HOOK_SECRET]
  -k, --key=        Specify a Key file to enable server to start TLS [$BARYON_KEY]
  -c, --cert=       Cert file for TLS [$BARYON_CERT]
  -o, --org=        Github Org to find cookbooks [$BARYON_GITHUB_ORG]
  -t, --token=      Github API token to use when connecting [$BARYON_GITHUB_TOKEN]
  -i, --interval=   Interval to perform full sync against github repos. Supports Golang duration formatting '1h2m'... etc. (12h)
                    [$BARYON_INTERVAL]
      --no-sync     Do NOT perform a github scan/sync when starting. Periodic sync will still fire [$BARYON_NOSYNC]
      --berks-only  Only use berks compatable version tags in the universe [$BARYON_BERKSONLY]

Help Options:
  -h, --help        Show this help message
```


## Build from source
This project is using the Go 1.5 vendor experiment to manage dependencies. Fetch the repo source per normal with go get:
```Shell
go get github.com/pantheon-systems/baryon
```

#### Vendored Deps
Then run `make` in the source directory. Make will 'go get' gvt, which is used to manage vendoring dependencies into the `./vendor` directory, and build from source
```Shell
cd $GOPATH/src/github.com/pantheon-systems/baryon && make
```

## Contributing
See the [CONTRIBUTING.md](CONTRIBUTING.md) documentation
