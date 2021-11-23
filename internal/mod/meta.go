package mod

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/mod/module"
	"golang.org/x/net/html"
)

// Credentials represents the credentials required to get a module.
type Credentials struct {
	User  string
	Token string
}

// Meta represents a module meta
type Meta struct {
	RootPath   string
	VCS        string
	RepoURL    string
	Credetials *Credentials
}

func GetMeta(mod module.Version) (*Meta, error) {
	// First look for VCS qualifier
	re := regexp.MustCompile(`(?m)\.(bzr|fossil|git|hg|svn)($|/)`)
	if loc := re.FindStringIndex(mod.Path); loc != nil {
		path := mod.Path[:loc[0]]
		rurl := strings.TrimRight(mod.Path[:loc[1]], "/")
		vcs := rurl[loc[0]+1:]
		return &Meta{
			RootPath:   path,
			VCS:        vcs,
			RepoURL:    rurl,
			Credetials: nil,
		}, nil
	}

	// No VCS qualifier found, start firing GET requests
	private := IsPrivate(mod.Path)
	var usr, pwd string
	if private {
		// private module, get credentials
		usr, pwd = CredentialsFor(strings.SplitN(mod.Path, "/", 2)[0])
	}
	crm := make(chan *Meta)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Cancel requests as soon as we find the repo

	path := mod.Path
	for path != "." {
		// try every paths
		go getMeta(mod.Path, "https://"+path, usr, pwd, ctx, crm)
		path = filepath.Dir(path)
	}

	for {
		select {
		case rm := <-crm:
			if private {
				rm.Credetials = &Credentials{User: usr, Token: pwd}
			}
			return rm, nil
		case <-ctx.Done(): // Timeout
			return nil, fmt.Errorf("finding repo: %w", ctx.Err())
		}
	}
}

func getMeta(mod, url, usr, pwd string, ctx context.Context, res chan<- *Meta) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req = req.WithContext(ctx)

	if pwd != "" {
		req.SetBasicAuth(usr, pwd)
	}

	q := req.URL.Query()
	q.Add("go-get", "1")
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	rm, err := extract(resp.Body)
	if err != nil {
		return
	}

	if rm.RootPath == mod {
		select {
		case res <- rm:
			return
		case <-ctx.Done():
			return
		}
	}
}

func extract(r io.Reader) (*Meta, error) {
	z := html.NewTokenizer(r)

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return nil, z.Err()
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			if t.Data == `body` {
				return nil, fmt.Errorf("Could not find go metadata")
			}
			if t.Data == "meta" {
				found := false
				content := ""
				for _, attr := range t.Attr {
					if attr.Key == "name" && attr.Val == "go-import" {
						found = true
					}
					if attr.Key == "content" {
						content = attr.Val
					}
				}
				if found {
					return parseMeta(content)
				}
			}
		}
	}
}

func parseMeta(content string) (*Meta, error) {
	s := strings.Split(content, " ")
	if len(s) != 3 {
		return nil, fmt.Errorf("Unexpected go-import length")
	}
	return &Meta{
		RootPath: s[0],
		VCS:      s[1],
		RepoURL:  s[2],
	}, nil
}
