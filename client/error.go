package client

import (
	"encoding/json"
	"fmt"

	"github.com/getcarina/carina/common"
)

// UserError is a user-ready error container
type UserError struct {
	error
	Context map[string]interface{}
}

func newClientError(err error) *UserError {
	return &UserError{
		error:   err,
		Context: common.Log.ErrorContext,
	}
}

// Cause returns the underlying cause of the error, if possible.
func (err *UserError) Cause() error {
	return err.error
}

func (err *UserError) Error() string {
	var hint string
	if !common.Log.DebugEnabled() {
		hint = "\n\nFor additional troubleshooting, re-run the command with --debug specified."
	}
	return fmt.Sprintf("%s%s%s", err.error.Error(), err.formatContext(), hint)
}

func (err *UserError) formatContext() string {
	if len(err.Context) == 0 {
		return ""
	}

	var context string
	result, jsonErr := json.MarshalIndent(err.Context, "", "\t")
	if jsonErr == nil {
		context = string(result)
	} else {
		context = common.Log.SDump(err.Context)
	}

	return fmt.Sprintf("\nContext:\n%s", context)
}
