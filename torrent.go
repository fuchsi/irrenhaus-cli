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
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/c2h5oh/datasize"
	api "github.com/fuchsi/irrenhaus-api"
	"github.com/fuchsi/irrenhaus-api/Category"
	"github.com/olekukonko/tablewriter"
)

func download(tid int64, destination string) (error) {
	PrintVerbose("Downloading torrent", tid)
	c := getConnection()

	body, filename, err := api.DownloadTorrent(c, tid)
	if err != nil {
		return err
	}

	PrintVerbose("Filename from Server:", filename)
	if destination == "" {
		destination = filename
	} else if !strings.HasSuffix(destination, ".torrent") {
		destination = destination + "/" + filename
	}

	file, err := os.Create(destination)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	writer.Write(body)

	PrintQuiet("Download to", destination, "complete")

	return nil
}

func upload(meta string, nfo string, image1 string, image2 string, name string, description string, category int) (error) {
	metard, err := os.Open(meta)
	if err != nil {
		return err
	}
	defer metard.Close()

	nford, err := os.Open(nfo)
	if err != nil {
		return err
	}
	defer nford.Close()

	imagerd, err := os.Open(image1)
	if err != nil {
		return err
	}
	defer imagerd.Close()

	temp, err := ioutil.ReadFile(description)
	if err != nil {
		return err
	}
	descriptionString := string(temp)

	t, err := api.NewUpload(getConnection(), metard, nford, imagerd, name, category, descriptionString)
	if err != nil {
		return err
	}
	if image2 != "" {
		image2rd, err := os.Open(image2)
		if err != nil {
			return err
		}
		defer image2rd.Close()
		t.Image2 = image2rd
	}

	if err := t.Upload(); err != nil {
		return err
	}

	fmt.Printf("Upload successful: %s/details.php?id=%d\n", config.Url, t.Id)
	return nil
}

func search(needle string, categories []int, dead bool) (error) {
	c := getConnection()

	entries, err := api.Search(c, needle, categories, dead)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d Torrents\n", len(entries))
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Size", "Date", "S", "L"})

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Added.Unix() > entries[j].Added.Unix()
	})

	for _, entry := range entries {
		table.Append([]string{
			fmt.Sprintf("%d", entry.Id),
			entry.Name,
			datasize.ByteSize(entry.Size).HumanReadable(),
			entry.Added.Format("02.01.2006 15:04:05"),
			fmt.Sprintf("%d", entry.SeederCount),
			fmt.Sprintf("%d", entry.LeecherCount),
		})
	}

	table.Render()

	return nil
}

func details(tid int64, subcommand string) (error) {
	c := getConnection()

	info := false
	files := false
	peers := false
	snatches := false

	switch subcommand {
	case "info":
		info = true
	case "files":
		files = true
	case "peers":
		peers = true
	case "snatcher":
		snatches = true
	case "all":
		info = true
		files = true
		peers = true
		snatches = true
	}

	entry, err := api.Details(c, tid, files, peers, snatches)
	if err != nil {
		return err
	}

	fmt.Println(entry.Name)

	if info {
		category, err := Category.ToString(entry.Category)
		if err != nil {
			category = "-"
		}
		fmt.Printf("ID:        %d\n", entry.Id)
		fmt.Printf("Info Hash: %s\n", entry.InfoHash)
		fmt.Printf("Category:  %s\n", category)
		fmt.Printf("Size:      %s\n", datasize.ByteSize(entry.Size).HumanReadable())
		fmt.Printf("Added:     %s\n", entry.Added.Format("02.01.2006 15:04:05"))
		fmt.Printf("#Files:    %d\n", entry.FileCount)
		fmt.Printf("#Seeders:  %d\n", entry.SeederCount)
		fmt.Printf("#Leechers: %d\n", entry.LeecherCount)
		fmt.Printf("#Snatched: %d\n", entry.SnatchCount)

		fmt.Println("Description:")
		fmt.Println(entry.Description)
	}

	if files {
		fmt.Println("Files:")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Size"})
		for _, file := range entry.Files {
			table.Append([]string{file.Name, datasize.ByteSize(file.Size).HumanReadable()})
		}
		table.Render()
	}

	if peers {
		fmt.Println("Seeders:")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Con", "ULed", "Up Rate", "DLed", "Down Rate", "Ratio", "Client"})

		for _, peer := range entry.Peers {
			if peer.Seeder {
				conStr := "No"
				ratioStr := "Inf."
				if peer.Connectable {
					conStr = "Yes"
				}
				if peer.Ratio > 0 {
					ratioStr = fmt.Sprintf("%0.3f", peer.Ratio)
				}
				table.Append([]string{
					peer.Name,
					conStr,
					datasize.ByteSize(peer.Uploaded).HumanReadable(),
					fmt.Sprintf("%s/s", datasize.ByteSize(peer.Ulrate).HumanReadable()),
					datasize.ByteSize(peer.Downloaded).HumanReadable(),
					fmt.Sprintf("%s/s", datasize.ByteSize(peer.Dlrate).HumanReadable()),
					ratioStr,
					peer.Client,
				})
			}
		}

		table.Render()

		fmt.Println("Leechers:")
		table2 := tablewriter.NewWriter(os.Stdout)
		table2.SetHeader([]string{"Name", "Con", "ULed", "Up Rate", "DLed", "Down Rate", "Ratio", "Complete", "Client"})

		for _, peer := range entry.Peers {
			if !peer.Seeder {
				conStr := "No"
				if peer.Connectable {
					conStr = "Yes"
				}
				table2.Append([]string{
					peer.Name,
					conStr,
					datasize.ByteSize(peer.Uploaded).HumanReadable(),
					fmt.Sprintf("%s/s", datasize.ByteSize(peer.Ulrate).HumanReadable()),
					datasize.ByteSize(peer.Downloaded).HumanReadable(),
					fmt.Sprintf("%s/s", datasize.ByteSize(peer.Dlrate).HumanReadable()),
					fmt.Sprintf("%0.3f", peer.Ratio),
					fmt.Sprintf("%0.2f%%", peer.Completed),
					peer.Client,
				})
			}
		}

		table2.Render()
	}

	if snatches {
		fmt.Println("Snatches:")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "ULed", "DLed", "Ratio", "Stopped"})

		for _, snatch := range entry.Snatches {
			stoppedStr := "No"
			if !snatch.Seeding {
				stoppedStr = snatch.Stopped.Format("02.01.2006 15:04:05")
			}
			table.Append([]string{
				snatch.Name,
				datasize.ByteSize(snatch.Uploaded).HumanReadable(),
				datasize.ByteSize(snatch.Downloaded).HumanReadable(),
				fmt.Sprintf("%0.3f", snatch.Ratio),
				stoppedStr,
			})
		}

		table.Render()
	}

	return nil
}

func thank(tid int64) (error) {
	c := getConnection()

	ok, err := api.Thank(c, tid)
	if err != nil {
		return err
	}

	if ok {
		return nil
	}

	return errors.New("unknown error")
}
