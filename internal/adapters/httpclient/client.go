package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct{ HTTP *http.Client }

func New() Client { return Client{HTTP: &http.Client{Timeout: 60 * time.Second}} }
func (c Client) Download(ctx context.Context, url string) (io.ReadCloser, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("URL must start with http:// or https://")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		res.Body.Close()
		return nil, fmt.Errorf("download failed: %s", res.Status)
	}
	return res.Body, nil
}
