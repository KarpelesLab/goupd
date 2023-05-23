package goupd

import (
	"fmt"
	"io"
	"net/http"
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
