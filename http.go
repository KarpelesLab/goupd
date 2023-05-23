package goupd

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func httpGet(p string) ([]byte, error) {
	resp, err := http.Get(p)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP status code: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func httpGetFields(p string) ([]string, error) {
	res, err := httpGet(p)
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(res)), nil
}
