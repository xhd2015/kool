package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var ErrReleaseNotFound = errors.New("release not found")

type ReleaseClient struct {
	Token string
	Owner string
	Repo  string
}

func NewReleaseClient(token, owner, repo string) *ReleaseClient {
	return &ReleaseClient{
		Token: token,
		Owner: owner,
		Repo:  repo,
	}
}

type Release struct {
	ID  int64  `json:"id"`
	Tag string `json:"tag_name"`
}

func (c *ReleaseClient) do(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return http.DefaultClient.Do(req)
}

func (c *ReleaseClient) apiURL(path string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/%s", c.Owner, c.Repo, path)
}

func (c *ReleaseClient) GetReleaseByTag(tag string) (*Release, error) {
	resp, err := c.do("GET", c.apiURL("releases/tags/"+tag), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrReleaseNotFound
	}
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(data))
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func (c *ReleaseClient) UploadReleaseAsset(releaseID int64, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	name := filepath.Base(filePath)
	url := fmt.Sprintf("https://uploads.github.com/repos/%s/%s/releases/%d/assets?name=%s", c.Owner, c.Repo, releaseID, name)

	req, err := http.NewRequest("POST", url, f)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}
