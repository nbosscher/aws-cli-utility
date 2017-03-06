package main

import (
	"awscli/awscmds"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var api awscmds.AWSCmds

func main() {

	secret := os.Getenv("AWS_SECRETE")
	access := os.Getenv("AWS_ACCESS_ID")

	ky := flag.String("aws-secrete", "", "AWS secrete key to use for this session. Alternatively AWS_SECRETE environment variable can be used")
	acc := flag.String("aws-access-id", "", "AWS access id to use for this session. Alternatively AWS_ACCESS_ID environment variable can be used")
	rg := flag.String("aws-region", "us-east-1", "AWS region-id to use for this session")

	flag.Parse()

	if *ky != "" {
		secret = *ky
	}

	if *acc != "" {
		access = *acc
	}

	if secret == "" {
		fmt.Println("missing -aws-secrete")
		flag.Usage()
		log.Fatal()
	}

	if access == "" {
		fmt.Println("missing -aws-access-id")
		flag.Usage()
		log.Fatal()
	}

	api = awscmds.NewAWSCmds(access, secret, *rg)

	fmt.Println("Enter a command:")
	for {
		cmd, args := getCommand()
		exec(cmd, args...)
	}
}

func printCommands() {
	fmt.Println(`Commands:
s3-object-ls <bucket> [<search-prefix>]
   Lists the objects in a bucket and optionally searches for a specific prefix

s3-bucket-ls
   Lists buckets

s3-object-dl <bucket> <object-key> <local-file>
   Downloads a specific file to the local file system

s3-bucket-dl <bucket> <local-directory> [<modified-after>]
   Downloads an entire bucket to a local directory.
   <modified-after> should be in the format "2017-12-31"

help
   Prints this information

exit
	`)
}

func getCommand() (string, []string) {

	fmt.Print("awscli # ")
	input := make([]byte, 100)

	n, err := os.Stdin.Read(input)
	if err != nil {
		fmt.Println(err)
	}

	vals := strings.Trim(string(input[0:n]), newline+" ")
	cmd := vals

	k := strings.IndexRune(cmd, ' ')
	if k == -1 {
		return cmd, nil
	}

	cmd = vals[0:k]
	vals = vals[k+1:]

	args := []string{}
	start := 0
	quoted := false

	for i := 0; i < len(vals); i++ {
		switch vals[i] {
		case ' ':
			if !quoted {
				args = append(args, strings.Trim(vals[start:i], `" `))
				start = i + 1
			}
		case '"':
			quoted = !quoted
		}
	}

	args = append(args, strings.Trim(vals[start:], `" `))

	return cmd, args
}

func exec(cmd string, args ...string) {
	switch cmd {
	case "s3-object-ls":
		if len(args) == 1 {
			api.ListBucketFiles(args[0])
		} else if len(args) == 2 {
			api.SearchBucketFiles(args[0], args[1])
		} else {
			fmt.Println("object-ls requires 1 parameter. got", strings.Join(args, ","))
			return
		}
	case "s3-bucket-ls":
		api.ListBuckets()
	case "s3-object-dl":
		if len(args) != 3 {
			fmt.Println("s3-object-dl requires 3 parameters")
			return
		}
		api.DownloadFile(args[0], args[1], args[2])
	case "s3-bucket-dl":
		var mod time.Time

		if len(args) == 3 {

			var err error
			mod, err = time.ParseInLocation("2006-01-02", args[2], time.UTC)
			if err != nil {
				fmt.Println(err)
				fmt.Println("date parameter should be in format 2017-12-31")
				return
			}

		} else if len(args) == 2 {
			mod = time.Unix(0, 0)
		} else {
			fmt.Println("s3-bucket-dl requires 2 or 3 parameters")
			return
		}

		api.DownloadBucket(args[0], args[1], mod)
	case "exit":
		os.Exit(1)
	case "help":
		fallthrough
	default:
		printCommands()
	}
}
