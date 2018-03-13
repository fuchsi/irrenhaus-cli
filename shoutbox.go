/*
 * irrenhaus-cli, CLI client for irrenhaus.dyndns.dk
 * Copyright (C) 2018  Daniel Müller
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>
 */

package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	api "github.com/fuchsi/irrenhaus-api"
)

var ShoutboxID = map[string]int{
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

func shoutboxRead(box string) error {
	c := getConnection()

	boxID, ok := ShoutboxID[box]
	if !ok {
		return errors.New("invalid shoutbox name")
	}

	messages, err := api.ShoutboxRead(c, boxID, 0)
	if err != nil {
		return err
	}

	for _, message := range messages {
		// skip control messages
		if message.Event != nil {
			continue
		}
		fmt.Printf("[%s] <%s> %s\n", message.Date.Format("01.02 15:04"), message.User, message.Message)
	}

	return nil
}

func shoutboxWrite(box string, message string) error {
	c := getConnection()

	boxID, ok := ShoutboxID[box]
	if !ok {
		return errors.New("invalid shoutbox name")
	}

	ok, err := api.ShoutboxWrite(c, boxID, message)
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("failed to post message")
	}

	PrintQuiet("Message posted")
	return nil
}

func shoutboxPoll(box string, refresh int) error {
	c := getConnection()

	boxID, ok := ShoutboxID[box]
	if !ok {
		return errors.New("invalid shoutbox name")
	}

	messages, err := api.ShoutboxRead(c, boxID, 0)
	if err != nil {
		return err
	}

	maxID := int64(0)

	extraStatus := ""

	for _, message := range messages {
		// control messages
		if message.Event != nil {
			switch {
			case message.Event.Type&api.ShoutboxEventUserMessage == api.ShoutboxEventUserMessage:
				unreadMessages := message.Event.Data[1]
				extraStatus = fmt.Sprintf(" - %s unread messages", unreadMessages)
			}
			continue
		}
		fmt.Printf("[%s] <%s> %s\n", message.Date.Format("01.02 15:04"), message.User, message.Message)
		if message.Id > maxID {
			maxID = message.Id
		}
	}

	statusbar := "[refresh in %ds%s]%s\r"

	for {
		for i := 0; i < refresh; i++ {
			fmt.Printf(statusbar, refresh-i, extraStatus, strings.Repeat(" ", 65-len(extraStatus)))
			time.Sleep(time.Second * 1)
		}
		fmt.Print(strings.Repeat(" ", 80) + "\r")
		fmt.Print("[refreshing]" + strings.Repeat(" ", 68) + "\r")

		messages, err := api.ShoutboxRead(c, boxID, maxID)
		if err != nil {
			extraStatus = " - last error:" + err.Error()
			continue
		}

		fmt.Print(strings.Repeat(" ", 80) + "\r")
		extraStatus = ""
		for _, message := range messages {
			// control messages
			if message.Event != nil {
				switch {
				case message.Event.Type&api.ShoutboxEventUserMessage == api.ShoutboxEventUserMessage:
					unreadMessages := message.Event.Data[1]
					extraStatus = fmt.Sprintf(" - %s unread messages", unreadMessages)
				}
				continue
			}
			fmt.Printf("[%s] <%s> %s\n", message.Date.Format("01.02 15:04"), message.User, message.Message)
			if message.Id > maxID {
				maxID = message.Id
			}
		}
	}

	return nil // returns never
}
