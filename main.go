package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	jsdelivr  = 0
	sizeLimit = 1024 * 1024 * 100 // 100MB
	HOST      = "127.0.0.1"
	PORT      = ":8001"
	ASSET_URL = "https://hunshcn.github.io/gh-proxy"
	exps      = []*regexp.Regexp{
		regexp.MustCompile(`^(?:https?://)?github\.com/(?P<author>.+?)/(?P<repo>.+?)/(?:releases|archive)/.*$`),
		regexp.MustCompile(`^(?:https?://)?github\.com/(?P<author>.+?)/(?P<repo>.+?)/(?:blob|raw)/.*$`),
		regexp.MustCompile(`^(?:https?://)?github\.com/(?P<author>.+?)/(?P<repo>.+?)/(?:info|git-).*$`),
		regexp.MustCompile(`^(?:https?://)?raw\.(?:githubusercontent|github)\.com/(?P<author>.+?)/(?P<repo>.+?)/.*$`),
		regexp.MustCompile(`^(?:https?://)?gist\.(?:githubusercontent|github)\.com/(?P<author>.+?)/.+?/.+$`),
	}
	whiteList = []string{}
	blackList = []string{}
)

func init() {
	fmt.Println("Starting server on", HOST+PORT)
}

func main() {
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(HOST+PORT, nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/https:/") {
		u := strings.Replace(r.URL.Path, "/https:/", "https://", 1)
		fmt.Println("Request URL Received:", u)
		// Match the URL against regular expressions
		for _, exp := range exps {
			if exp.MatchString(u) {
				names := exp.SubexpNames()
				params := make(map[string]string)
				match := exp.FindStringSubmatch(u)
				for i, name := range names {
					if i != 0 && name != "" {
						params[name] = match[i]
					}
				}
				// Check if the URL is in the whitelist
				for _, wl := range whiteList {
					if strings.ContainsAny(wl, "/") {
						author := params["author"]
						repo := params["repo"]
						wl_author := strings.Split(wl, "/")[0]
						wl_repo := strings.Split(wl, "/")[1]
						if author == wl_author && repo == wl_repo {
							handleGitHubRedirect(w, r, u)
							return
						}
					} else {
						author := params["author"]
						if author == wl {
							handleGitHubRedirect(w, r, u)
							return
						}
					}
				}
				// Check if the URL is in the blacklist
				for _, bl := range blackList {
					if strings.ContainsAny(bl, "/") {
						author := params["author"]
						repo := params["repo"]
						bl_author := strings.Split(bl, "/")[0]
						bl_repo := strings.Split(bl, "/")[1]
						if author == bl_author && repo == bl_repo {
							http.Error(w, "URL is blacklisted", http.StatusForbidden)
							return
						}
					} else {
						author := params["author"]
						if author == bl {
							http.Error(w, "URL is blacklisted", http.StatusForbidden)
							return
						}
					}
				}
				handleGitHubRedirect(w, r, u)
				return
			}
		}
	} else {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
	}
}

func handleGitHubRedirect(w http.ResponseWriter, r *http.Request, u string) {
	if jsdelivr == 1 {
		_u := strings.Replace(u, "raw.githubusercontent.com", "cdn.jsdelivr.net/gh", 1)
		u = _u
		if _u == u {
			u = strings.Replace(u, "raw.github.com", "cdn.jsdelivr.net/gh", 1)
		}
		http.Redirect(w, r, u, http.StatusFound)
		return
	} else if exps[1].MatchString(u) {
		u = strings.Replace(u, "/blob/", "/raw/", 1)
	}

	quotedURL := quoteURL(u)
	proxyRequest(w, r, quotedURL)
}

func quoteURL(u string) string {
	quotedURL, err := url.QueryUnescape(u)
	if err != nil {
		quotedURL = u
	}
	return quotedURL
}

func proxyRequest(w http.ResponseWriter, r *http.Request, u string) {
	req, err := http.NewRequest(r.Method, u, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header = r.Header
	req.Host = ""
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.ContentLength > int64(sizeLimit) {
		http.Redirect(w, r, u+r.URL.String()[len(r.URL.Path):], http.StatusFound)
		return
	}

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
