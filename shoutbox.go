package main

import (
	"errors"
	"fmt"
	"time"

	api "github.com/fuchsi/irrenhaus-api"
)

var ShoutboxId = map[string]int{
	"user": 1,
	"team": 2,
}

func shoutboxUsage() {
	fmt.Println("shout subcommand")

	fmt.Println("\tread [box]")
	fmt.Println("\t\tList the messages in [box]")

	fmt.Println("\twrite [box] <message>")
	fmt.Println("\t\tWrite a message to [box]")

	fmt.Println("\tpoll [box] [refresh]")
	fmt.Println("\t\tPoll the [box] evenry [refresh] seconds and display new messages")

	fmt.Println("\tbox can be 'user' or 'team' and defaults to 'user' if ommited")
}

func shoutboxRead(box string) (error) {
	c := getConnection()

	boxId, ok := ShoutboxId[box]
	if !ok {
		return errors.New("invalid shoutbox name")
	}

	messages, err := api.ShoutboxRead(c, boxId, 0)
	if err != nil {
		return err
	}

	for _, message := range messages {
		fmt.Printf("[%s] <%s> %s\n", message.Date.Format("01.02 15:04"), message.User, message.Message)
	}

	return nil
}

func shoutboxWrite(box string, message string) (error) {
	c := getConnection()

	boxId, ok := ShoutboxId[box]
	if !ok {
		return errors.New("invalid shoutbox name")
	}

	ok, err := api.ShoutboxWrite(c, boxId, message)
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("failed to post message")
	}

	PrintQuiet("Message posted")
	return nil
}

func shoutboxPoll(box string, refresh int) (error) {
	c := getConnection()

	boxId, ok := ShoutboxId[box]
	if !ok {
		return errors.New("invalid shoutbox name")
	}

	messages, err := api.ShoutboxRead(c, boxId, 0)
	if err != nil {
		return err
	}

	maxId := int64(0)

	for _, message := range messages {
		fmt.Printf("[%s] <%s> %s\n", message.Date.Format("01.02 15:04"), message.User, message.Message)
		if message.Id > maxId {
			maxId = message.Id
		}
	}

	for {
		messages, err := api.ShoutboxRead(c, boxId, maxId)
		if err != nil {
			return err
		}

		for _, message := range messages {
			fmt.Printf("[%s] <%s> %s\n", message.Date.Format("01.02 15:04"), message.User, message.Message)
			if message.Id > maxId {
				maxId = message.Id
			}
		}

		time.Sleep(time.Second * time.Duration(refresh))
	}

	return nil
}
