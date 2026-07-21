package sandbox

import (
	"encoding/json"
	"time"
)

// PackBlob is the cleartext payload sealed into a sandbox binary.
type PackBlob struct {
	Version   int               `json:"version"`
	Name      string            `json:"name"`
	CreatedAt time.Time         `json:"created_at"`
	ExpiresAt *time.Time        `json:"expires_at,omitempty"`
	Comment   string            `json:"comment,omitempty"`
	Files     []PackFile        `json:"files"`
	Env       map[string]string `json:"env"`
}

// PackFile is one packed file entry.
type PackFile struct {
	Path    string `json:"path"`
	Mode    uint32 `json:"mode"`
	Content []byte `json:"content"`
}

const packBlobVersion = 1

func marshalPackBlob(b *PackBlob) ([]byte, error) {
	return json.Marshal(b)
}

func unmarshalPackBlob(data []byte) (*PackBlob, error) {
	var b PackBlob
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	return &b, nil
}
