package awscmds

import (
	"testing"
	"time"
)

func getCmd() AWSCmds {
	return NewAWSCmds("<aws-access-id>", "<aws-secrete>", "us-east-1")
}

const test_bucket = "<test-bucket>"

func TestAWSCmds_ListBuckets(t *testing.T) {
	getCmd().ListBuckets()
}

func TestAWSCmds_ListBucketFiles(t *testing.T) {
	getCmd().ListBucketFiles(test_bucket)
}

func TestAWSCmds_SearchBucketFiles(t *testing.T) {
	getCmd().SearchBucketFiles(test_bucket, "<test-search>")
}

func TestAWSCmds_DownloadFile(t *testing.T) {
	getCmd().DownloadFile(test_bucket, "<test-key>", "/tmp/my-dl")
}

func TestAWSCmds_DownloadBucket(t *testing.T) {
	getCmd().DownloadBucket(test_bucket, "/tmp/"+test_bucket, time.Now().Add(-5*24*time.Hour))
}
