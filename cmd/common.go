package cmd

import "errors"

func bindName(args []string, name *string) error {
	if len(args) < 1 {
		return errors.New("A cluster name is required")
	}
	*name = args[0]
	return nil
}
