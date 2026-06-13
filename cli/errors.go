package cli

import (
	"errors"

	"github.com/tamnd/clinicaltrials-cli/clinicaltrials"
)

func isNotFound(err error) bool {
	return errors.Is(err, clinicaltrials.ErrNotFound)
}
