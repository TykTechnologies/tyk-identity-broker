package identityHandlers

import (
	"github.com/gofrs/uuid"
)

func newUUID() string {
	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return id.String()
}
