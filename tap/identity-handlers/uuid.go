package identityHandlers

import (
	"github.com/gofrs/uuid"
)

func uuid() string {
	id, err := uuid.NewV5()
	if err != nil {
		panic(err)
	}
	return id.String()
}
