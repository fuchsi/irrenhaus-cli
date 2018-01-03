package main

import (
	"errors"

	api "github.com/fuchsi/irrenhaus-api"
)

func comment(tid int64, message string) (error) {
	c := getConnection()

	ok, err := api.CommentWrite(c, tid, message)
	if err != nil {
		return err
	}

	if ok {
		return nil
	}

	return errors.New("unknown error")
}
