package generic

import (
	"fmt"
	"strings"
)

type ErrMissingOutputs struct {
	OutputNames []string
}

func (e ErrMissingOutputs) Error() string {
	return fmt.Sprintf(`The module used for this app is missing outputs. Add the following outputs before deploying this app through the Nullstone CLI: %s`, strings.Join(e.OutputNames, ","))
}
