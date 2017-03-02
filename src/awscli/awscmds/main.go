package awscmds

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"github.com/aws/aws-sdk-go/service/s3"
	"fmt"
	"strconv"
	"os"
	"io"
	"path/filepath"
	"strings"
	"time"
)

type AWSCmds interface {
	ListBuckets()
	ListBucketFiles(bucket string)
	SearchBucketFiles(bucket, prefix string)
	DownloadFile(bucket, key, localFile string)
	DownloadBucket(bucket, localDir string, modifiedAfter time.Time)
}

type awsCmds struct {
	accessId string
	secrete string
	region string

	sess *session.Session
	s3 *s3.S3
}

func NewAWSCmds(accessId, secrete, region string) AWSCmds {
	return &awsCmds{
		accessId: accessId,
		secrete: secrete,
		region: region,
	}
}

func (c *awsCmds) getSession() *session.Session {
	var err error

	if c.sess == nil {
		c.sess, err = session.NewSession(&aws.Config{
			Region:      aws.String(c.region),
			Credentials: credentials.NewStaticCredentials(c.accessId, c.secrete, ""),
		})

		if err != nil {
			fmt.Println(err)
			log.Fatal()
		}
	}

	return c.sess
}

func (c *awsCmds) getS3() *s3.S3 {
	if c.s3 == nil {
		c.s3 = s3.New(c.getSession())
	}

	return c.s3
}

func (c *awsCmds) ListBuckets() {

	fmt.Println("Listing buckets...")

	bu, err := c.getS3().ListBuckets(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println()

	fmt.Printf("Owner: %s"+newline, *bu.Owner.DisplayName)
	fmt.Printf("%-20s %-30s"+newline, "Bucket Name", "Created")
	fmt.Println("-------------------- ------------------------------")

	for _, bu := range bu.Buckets {
		fmt.Printf("%-20s %30s"+newline, *bu.Name, bu.CreationDate.String())
	}

	fmt.Println()
}

func (c *awsCmds) listBucketFiles(bucket string) (*s3.ListObjectsOutput, error) {
	li, err := c.getS3().ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		MaxKeys: aws.Int64(10000),
	})

	if err != nil {
		return nil, err
	}

	lst := li
	var marker string
	if len(li.Contents) != 0 {
		marker = *li.Contents[len(li.Contents)-1].Key
	}

	for *li.IsTruncated {
		fmt.Println("performing additional request... (got", len(lst.Contents), "files so far)")
		li, err = c.getS3().ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(bucket),
			MaxKeys: aws.Int64(10000),
			Marker: aws.String(marker),
		})
		if err != nil {
			return nil, err
		}

		lst.Contents = append(lst.Contents, li.Contents[1:]...)

		if len(li.Contents) != 0 {
			marker = *li.Contents[len(li.Contents)-1].Key
		}
	}

	return lst, nil
}

func (c *awsCmds) ListBucketFiles(bucket string) {
	fmt.Println("Listing bucket files...")

	lst, err := c.listBucketFiles(bucket)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.renderBucketList(lst)
}

func (c *awsCmds) SearchBucketFiles(bucket, prefix string) {
	fmt.Println("Searching bucket files...")

	li, err := c.getS3().ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		MaxKeys: aws.Int64(10000),
		Prefix: aws.String(prefix),
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	lst := li
	var marker string
	if len(li.Contents) != 0 {
		marker = *li.Contents[len(li.Contents)-1].Key
	}

	for *li.IsTruncated {
		fmt.Println("performing additional request... (got", len(lst.Contents), "so far)")
		li, err = c.getS3().ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(bucket),
			MaxKeys: aws.Int64(10000),
			Marker: aws.String(marker),
			Prefix: aws.String(prefix),
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		lst.Contents = append(lst.Contents, li.Contents[1:]...)

		if len(li.Contents) != 0 {
			marker = *li.Contents[len(li.Contents)-1].Key
		}
	}

	fmt.Printf("Term:        %s"+newline, prefix)

	c.renderBucketList(lst)
}

func (c *awsCmds) renderBucketList(lst *s3.ListObjectsOutput){
	sum := int64(0)
	for _, ob := range lst.Contents {
		sum += *ob.Size
	}

	fmt.Printf("Bucket:      %s"+newline, *lst.Name)
	fmt.Printf("Count:       %d"+newline, len(lst.Contents))
	fmt.Printf("Total size:  %s"+newline, formatFileSize(sum))
	fmt.Printf("%-60s %-20s %-30s %7s %10s %-20s"+newline, "Key", "Owner", "Last Modified", "Size", "Size", "Class")
	fmt.Println("--------------------------------------------------------------------------------------------------------------------------------------------")
	for _, ob := range lst.Contents {
		fmt.Printf("%-60s %-20s %-30s %7s %10s %-20s"+newline, *ob.Key, *ob.Owner.DisplayName, ob.LastModified.String(), formatFileSize(*ob.Size), strconv.FormatInt(*ob.Size,10), *ob.StorageClass)
	}

	fmt.Println()
}

func (c *awsCmds) DownloadFile(bucket, key, localFile string) {
	fmt.Println("Downloading file...")

	by, err := c.downloadFile(bucket, key, localFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(by, "bytes written to", localFile)
	fmt.Println()

}

func (c *awsCmds) downloadFile(bucket, key, localFile string) (int64, error) {
	o, err := c.getS3().GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key: aws.String(key),
	})

	if err != nil {
		return 0, err
	}

	fi, err := os.OpenFile(localFile, os.O_CREATE | os.O_RDWR, os.ModePerm)
	if err != nil {
		return 0, err
	}

	by, err := io.Copy(fi, o.Body)
	if err != nil {
		return 0, err
	}

	return by, nil
}

func (c *awsCmds) DownloadBucket(bucket, localDir string, modifiedAfter time.Time) {

	lst, err := c.listBucketFiles(bucket)
	if err != nil {
		fmt.Println(err)
		return
	}



	tot := int64(0)
	totct := 0
	for _, v := range lst.Contents {
		if v.LastModified.After(modifiedAfter) {
			tot += *v.Size
			totct++
		}
	}

	sum := int64(0)

	for i, v := range lst.Contents {
		if !v.LastModified.After(modifiedAfter) {
			continue
		}

		if *v.Size == 0 {
			// for directories
			continue
		}

		path := filepath.Join(localDir, *v.Key)

		err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			log.Println(err)
			return
		}

		by, err := c.downloadFile(bucket, *v.Key, path)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Download: Wrote %-80s %10s"+newline, path, formatFileSize(by))

		sum += *v.Size

		prog := (i+1) * 100 / totct
		fmt.Printf("Progress: %-100s %3d%% %10s of %s"+newline, strings.Repeat("=", prog), prog, formatFileSize(sum), formatFileSize(tot))
	}

	fmt.Println("")
}

func formatFileSize(size int64) string {
	if size < 1e3 {
		return strconv.FormatInt(size, 10) + " B "
	}

	if size < 1e6 {
		return strconv.FormatInt(size / 1e3, 10) + " KB"
	}

	if size < 1e9 {
		return strconv.FormatInt(size / 1e6, 10) + " MB"
	}

	//if size < 10e12 {
	return strconv.FormatInt(size / 1e9, 10) + " GB"
	//}
}