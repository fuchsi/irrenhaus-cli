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
	"fmt"
	"os"

	"bufio"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pborman/getopt/v2"
)

const VERSION = "v0.0.1"

var CONFIGPATH = ".irrenhaus-cli/"
var config Configuration
var configFile string

var helpFlag = getopt.BoolLong("help", 'h', "Show this help message and exit")
var versionFlag = getopt.BoolLong("version", 'V', "Print version and quit")
var verboseFlag = getopt.BoolLong("verbose", 'v', "verbose output")
var quietFlag = getopt.BoolLong("quiet", 'q', "no output")
var configOpt = getopt.StringLong("config", 'C', "", "Path to the config file")
var categoryOpt = getopt.ListLong("category", 'c', "", "Torrent category. See 'categories' for help.")
var nameOpt = getopt.StringLong("name", 'n', "", "Torrent name. See 'upload' for details.")
var deadFlag = getopt.BoolLong("dead", 'd', "Include dead torrents")

func main() {
	getopt.SetParameters("command args")

	// Parse the program arguments
	getopt.Parse()

	if *versionFlag {
		fmt.Println("irrenhaus-cli " + VERSION)
		return
	}
	if *helpFlag || getopt.NArgs() == 0 {
		getopt.Usage()
		return
	}

	command := getopt.Arg(0)

	if *configOpt == "" {
		CONFIGPATH = os.Getenv("HOME") + CONFIGPATH
		if _, err := os.Stat(CONFIGPATH); err != nil {
			os.Mkdir(CONFIGPATH, 0772)
		}
		configFile = CONFIGPATH + "config.json"
	} else {
		configFile = *configOpt
	}
	var err error
	config, err = loadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read config file: %s\n", err.Error())
		if command != "init" {
			command = "commands"
		}
	}

	if command != "init" && command != "commands" {
		newConnection()
	}

	switch command {
	case "init":
		var username, password, pin, url string
		url = "https://irrenhaus.dyndns.dk"
		switch getopt.NArgs() {
		case 1:
			AskFor("Username", &username)
			AskFor("Password", &password)
			AskFor("Pin", &pin)
		case 2:
			username = getopt.Arg(1)
			AskFor("Password", &password)
			AskFor("Pin", &pin)
		case 3:
			username = getopt.Arg(1)
			password = getopt.Arg(2)
			AskFor("Pin", &pin)
		case 4:
			username = getopt.Arg(1)
			password = getopt.Arg(2)
			pin = getopt.Arg(3)
		default:
			username = getopt.Arg(1)
			password = getopt.Arg(2)
			pin = getopt.Arg(3)
			url = getopt.Arg(4)
		}

		err := initConfig(username, password, pin, url)
		if err != nil {
			PrintError("failed to write config file:", err.Error())
		}
	case "download":
		if getopt.NArgs() < 2 {
			PrintError("Missing Torrent-ID")
		}
		tid, err := strconv.ParseInt(getopt.Arg(1), 10, 64)
		if err != nil {
			PrintError("TID is not a valid ID")
		}
		dest := ""
		if getopt.NArgs() > 2 {
			dest = getopt.Arg(2)
		}
		err = download(tid, dest)
		if err != nil {
			PrintError(err.Error())
		}
	case "upload":
		if getopt.NArgs() < 5 {
			PrintError("Missing parameters")
		}

		meta := getopt.Arg(1)
		nfo := getopt.Arg(2)
		description := getopt.Arg(3)
		image1 := getopt.Arg(4)
		image2 := ""
		if getopt.NArgs() > 5 {
			image2 = getopt.Arg(5)
		}
		name := ""
		if *nameOpt != "" {
			name = *nameOpt
		} else {
			name = filepath.Base(meta)
		}
		category, err := strconv.ParseInt((*categoryOpt)[0], 10, 32)
		if err != nil {
			PrintError(err.Error())
		}

		err = upload(meta, nfo, image1, image2, name, description, int(category))
		if err != nil {
			PrintError(err.Error())
		}
	case "search":
		if getopt.NArgs() < 2 {
			PrintError("Missing Search string")
		}
		needle := strings.Join(getopt.Args()[1:], " ")
		categories := make([]int, 0)
		for _, c := range *categoryOpt {
			ci, err := strconv.ParseInt(c, 10, 32)
			if err != nil {
				continue
			}
			categories = append(categories, int(ci))
		}

		err = search(needle, categories, *deadFlag)
		if err != nil {
			PrintError(err.Error())
		}
	case "details":
		if getopt.NArgs() < 3 {
			fmt.Println("details <tid> subcommand")

			fmt.Println("\tinfo")
			fmt.Println("\t\tShows the basic informations")

			fmt.Println("\tfiles")
			fmt.Println("\t\tList the files")

			fmt.Println("\tpeers")
			fmt.Println("\t\tList the Peers")

			fmt.Println("\tsnatch")
			fmt.Println("\t\tList the Snatchers")

			fmt.Println("\tall")
			fmt.Println("\t\tShow all informations")
			break
		}

		tid, err := strconv.ParseInt(getopt.Arg(1), 10, 64)
		if err != nil {
			PrintError("TID is not a valid ID")
		}

		subcommand := getopt.Arg(2)
		err = details(tid, subcommand)
		if err != nil {
			PrintError(err.Error())
		}
	case "thank":
		if getopt.NArgs() < 2 {
			PrintError("Missing Torrent-ID")
		}
		tid, err := strconv.ParseInt(getopt.Arg(1), 10, 64)
		if err != nil {
			PrintError("TID is not a valid ID")
		}

		err = thank(tid)
		if err != nil {
			PrintError(err.Error())
		}
	case "comment":
		if getopt.NArgs() < 3 {
			PrintError("Missing Parameters")
		}
		tid, err := strconv.ParseInt(getopt.Arg(1), 10, 64)
		if err != nil {
			PrintError("TID is not a valid ID")
		}
		message := strings.Join(getopt.Args()[2:], " ")

		err = comment(tid, message)
		if err != nil {
			PrintError(err.Error())
		}
	case "shout":
		if getopt.NArgs() < 2 {
			shoutboxUsage()
			break
		}

		subcommand := getopt.Arg(1)
		box := "user"
		argOffset := 2
		if getopt.Arg(2) == "team" {
			box = "team"
			argOffset = 3
		}

		switch subcommand {
		case "read":
			err := shoutboxRead(box)
			if err != nil {
				PrintError(err.Error())
			}
		case "write":
			if getopt.NArgs() > argOffset {
				message := strings.Join(getopt.Args()[argOffset:], " ")
				err := shoutboxWrite(box, message)
				if err != nil {
					PrintError(err.Error())
				}
				break
			}

			shoutboxUsage()
			PrintError("Too few parameters")
		case "poll":
			refresh := 10
			if getopt.Arg(argOffset) != "" {
				temp, err := strconv.ParseInt(getopt.Arg(argOffset), 10, 32)
				if err != nil {
					shoutboxUsage()
					PrintError("refresh is not a number")
				}
				refresh = int(temp)
			}

			err := shoutboxPoll(box, refresh)
			if err != nil {
				PrintError(err.Error())
			}
		}
	case "commands":
		fallthrough
	default:
		fmt.Println("commands:")
		fmt.Println("\tinit [username] [password] [pin] [url]")
		fmt.Println("\t\tInitialize the config file")

		fmt.Println("\tdownload <tid> [destination]")
		fmt.Println("\t\tDownload a torrent file")

		fmt.Println("\tupload -c category [-n name] <torrent> <nfo> <description> <image1> [image2]")
		fmt.Println("\t\tUpload a torrent file")

		fmt.Println("\tsearch [-c category] [-d] <search>")
		fmt.Println("\t\tSearch for torrents")

		fmt.Println("\tdetails <tid> <subcommand>")
		fmt.Println("\t\tShow the details of a torrent")

		fmt.Println("\tthank <tid>")
		fmt.Println("\t\tThank the uploader for the torrent")

		fmt.Println("\tcomment <tid> <message>")
		fmt.Println("\t\tWrite a comment for a torrent")

		fmt.Println("\tshout <subcommand>")
		fmt.Println("\t\tShoutbox/Chat commands")

		fmt.Println("\tmessage <subcommand>")
		fmt.Println("\t\tMessage commands")

		fmt.Println("\tcommands")
		fmt.Println("\t\tPrint this command list")
	}

	if connection.GetCookies().Uid > 0 {
		dumpCookies(connection.GetCookies())
	}
}

// Initialize and write the config
func initConfig(username, password, pin, url string) (error) {
	config.Username = username
	config.Password = password
	config.Pin = pin
	config.Url = url

	return dumpConfig(config, configFile)
}

// Print a line to stdout if the verbose flag is set.
// It returns the number of bytes written and any write error encountered.
func PrintVerbose(a ...interface{}) (n int, err error) {
	if *verboseFlag {
		return fmt.Println(a)
	}

	return 0, nil
}

// Print a line to stdout if the quiet flag is not set.
// It returns the number of bytes written and any write error encountered.
func PrintQuiet(a ...interface{}) (n int, err error) {
	if !*quietFlag {
		fmt.Println(a)
	}

	return 0, nil
}

// Print a line to stderr and exit with status 1
func PrintError(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a)
	os.Exit(1)
}

// Ask the user for some input
//  prompt: Prompt/Question for the user
//  result: Input from the user
// It returns any error encountered.
func Ask(prompt string, result *string) (error) {
	fmt.Printf("%s: ", prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	*result = scanner.Text()

	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}

// Ask the user for some input, until result is not empty
//  prompt: Prompt/Question for the user
//  result: Input from the user
// It returns any error encountered.
func AskFor(prompt string, result *string) (error) {
	for len(*result) == 0 {
		err := Ask(prompt, result)
		if err != nil {
			return err
		}
	}

	return nil
}
