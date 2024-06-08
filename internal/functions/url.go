package functions

import (
	"net/url"
)

type urlInfo struct {
	Scheme   string
	Username string
	Password string
	Hostname string
	Port     string
	Path     string
	Query    map[string][]string
	Fragment string
}

func parseUrlInfo(s string) (urlInfo, error) {
	u, err := url.Parse(s)
	if err != nil {
		return urlInfo{}, err
	}

	password, _ := u.User.Password()

	return urlInfo{
		Scheme:   u.Scheme,
		Username: u.User.Username(),
		Password: password,
		Hostname: u.Hostname(),
		Port:     u.Port(),
		Path:     u.Path,
		Query:    u.Query(),
		Fragment: u.Fragment,
	}, nil
}
