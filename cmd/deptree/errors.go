package main

import "fmt"

type missingFlagError struct {
	flag string
}

func (e *missingFlagError) Error() string {
	return fmt.Sprintf("required flag --%s is missing", e.flag)
}
