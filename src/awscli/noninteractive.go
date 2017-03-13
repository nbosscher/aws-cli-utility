package main

import (
	"awscli/awscmds"
	"fmt"
	"log"
	"time"
)

const (
	NonInteractiveCmdHelp               = "help"
	NonInteractiveCmdBackupBucketHourly = "backup-bucket-hourly"
)

type NonInteractiveCmd string

func executeNonInteractive(api awscmds.AWSCmds, cmd string, arg []string) {
	switch NonInteractiveCmd(cmd) {
	case NonInteractiveCmdBackupBucketHourly:
		var err error
		if len(arg) != 2 && len(arg) != 3 {
			printNonInteractiveHelp()
			return
		}

		bucketName := arg[0]
		localDir := arg[1]
		sinceDate := time.Unix(0, 0)

		if len(arg) == 3 {
			sinceDate, err = time.Parse("2006-01-02", arg[2])
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		for {
			started := time.Now()
			bytes, err := api.DownloadBucketSilent(bucketName, localDir, sinceDate)
			if err != nil {
				log.Println(err)
			} else {
				log.Printf("Downloaded %d bytes"+newline, bytes)
			}

			sinceDate = started
			time.Sleep(time.Now().Sub(sinceDate) + 1*time.Hour)
		}
	case NonInteractiveCmdHelp:
		fallthrough
	default:
		printNonInteractiveHelp()
	}
}

func printNonInteractiveHelp() {
	fmt.Println("Non-interactive commands:")
	fmt.Println("   -command=backup-bucket-hourly -arg=<bucket-name> -arg=<local-dir> [-arg=<since-date in yyyy-mm-dd>]")
}
