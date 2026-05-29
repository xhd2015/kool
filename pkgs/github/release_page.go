package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CreateReleaseInput struct {
	TagName    string
	Name       string
	Body       string
	Draft      bool
	Prerelease bool
}

func (c *ReleaseClient) CreateRelease(input CreateReleaseInput) (*Release, error) {
	if input.Name == "" {
		input.Name = input.TagName
	}

	body := map[string]interface{}{
		"tag_name":   input.TagName,
		"name":       input.Name,
		"body":       input.Body,
		"draft":      input.Draft,
		"prerelease": input.Prerelease,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := c.do("POST", c.apiURL("releases"), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func (c *ReleaseClient) GetOrCreateRelease(tag string) (*Release, error) {
	rel, err := c.GetReleaseByTag(tag)
	if err == nil {
		return rel, nil
	}
	if err != ErrReleaseNotFound {
		return nil, err
	}

	fmt.Printf("Release for tag %s not found, creating...\n", tag)
	return c.CreateRelease(CreateReleaseInput{
		TagName: tag,
	})
}
