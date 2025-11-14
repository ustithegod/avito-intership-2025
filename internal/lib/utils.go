package lib

import "fmt"

// Err returns formatted error in "op: err" template.
func Err(op string, err error) error {
	return fmt.Errorf("%s: %w", op, err)
}
