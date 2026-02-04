package example

import (
	"context"
	"fmt"

	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/log"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/types"
)

// GetRequest is the request for getting an example item
type GetRequest struct {
	ID types.ID `json:"id" form:"id"`
}

// GetResponse is the response for getting an example item
type GetResponse struct {
	ID   types.ID `json:"id"`
	Name string   `json:"name"`
}

// Get handles GET /api/example
// Example usage:
//
//	curl "http://localhost:8080/api/example?id=1"
//
// Response:
//
//	{"code":0,"data":{"id":1,"name":"Example Item 1"}}
func Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	log.Infof(ctx, "getting example item: id=%d", req.ID)

	if req.ID == 0 {
		return nil, fmt.Errorf("id is required")
	}

	// In a real application, you would fetch from database here
	return &GetResponse{
		ID:   req.ID,
		Name: fmt.Sprintf("Example Item %d", req.ID),
	}, nil
}

// CreateRequest is the request for creating an example item
type CreateRequest struct {
	Name string `json:"name"`
}

// CreateResponse is the response for creating an example item
type CreateResponse struct {
	ID types.ID `json:"id"`
}

// Create handles POST /api/example/create
// Example usage:
//
//	curl -X POST "http://localhost:8080/api/example/create" \
//	  -H "Content-Type: application/json" \
//	  -d '{"name":"My Item"}'
//
// Response:
//
//	{"code":0,"data":{"id":1}}
func Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	log.Infof(ctx, "creating example item: name=%s", req.Name)

	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// In a real application, you would insert into database here
	// For demo, return a fake ID
	return &CreateResponse{
		ID: 1,
	}, nil
}
