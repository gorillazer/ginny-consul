package consul

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type healthFilter int

const (
	healthFilterUndefined healthFilter = iota
	healthFilterOnlyHealthy
	healthFilterFallbackToUnhealthy
)

const scheme = "consul"

func parseEndpoint(url *url.URL) (serviceName, scheme string, tags []string, health healthFilter, token string, err error) {
	const defHealthFilter = healthFilterOnlyHealthy

	// url.Path contains a leading "/", when the URL is in the form
	// scheme://host/path, remove it
	serviceName = strings.TrimPrefix(url.Path, "/")
	if serviceName == "" {
		return "", "", nil, health, "", errors.New("path is missing in url")
	}

	scheme, tags, health, token, err = extractOpts(url.Query())
	if err != nil {
		return "", "", nil, health, "", err
	}

	if health == healthFilterUndefined {
		health = defHealthFilter
	}

	return serviceName, scheme, tags, health, token, nil
}

func extractOpts(opts url.Values) (scheme string, tags []string, health healthFilter, token string, err error) {
	for key, values := range opts {
		if len(values) == 0 {
			continue
		}
		value := values[len(values)-1]

		switch strings.ToLower(key) {
		case "scheme":
			scheme = strings.ToLower(value)
			if scheme != "http" && scheme != "https" {
				return "", nil, healthFilterUndefined, "", fmt.Errorf("unsupported scheme '%s'", value)
			}

		case "tags":
			tags = strings.Split(value, ",")

		case "health":
			switch strings.ToLower(value) {
			case "healthy":
				health = healthFilterOnlyHealthy
			case "fallbacktounhealthy":
				health = healthFilterFallbackToUnhealthy
			default:
				return "", nil, healthFilterUndefined, "", fmt.Errorf("unsupported health parameter value: '%s'", value)
			}
		case "token":
			token = value

		default:
			return "", nil, healthFilterUndefined, "", fmt.Errorf("unsupported parameter: '%s'", key)
		}
	}

	return scheme, tags, health, token, err
}
