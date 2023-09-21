package utils

import "net/url"

type QueryParam struct {
	Key   string
	Value string
}

func AddQueryParam(rawURL string, queryParam ...QueryParam) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Extract existing query parameters
	q := u.Query()

	// Add new query parameter
	for _, param := range queryParam {
		q.Add(param.Key, param.Value)
	}

	// Update the URL with the new query parameters
	u.RawQuery = q.Encode()

	return u.String(), nil
}
