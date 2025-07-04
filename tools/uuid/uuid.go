package uuid

import (
	"fmt"

	"github.com/google/uuid"
)

func Handle(args []string) error {
	//  "github.com/google/uuid"
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	fmt.Println(id)
	return nil
}
