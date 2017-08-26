package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type clusterBackup struct {
	session   *session.Session
	client    *s3.S3
	fileName  string
	bucket    string
	maxDaily  int
	maxHourly int
}

func main() {
	var bucketName string
	if bucketName = os.Getenv("BACKUP_BUCKET"); bucketName == "" {
		log.Fatal("No upload bucket")
	}

	maxDaily := 10
	maxHourly := 24
	if tmp := os.Getenv("MAX_DAILY"); tmp != "" {
		maxDaily, _ = strconv.Atoi(tmp)
	}
	if tmp := os.Getenv("MAX_HOURLY"); tmp != "" {
		maxHourly, _ = strconv.Atoi(tmp)
	}

	backup := clusterBackup{
		fileName:  "cluster-backup.tar",
		bucket:    bucketName,
		maxDaily:  maxDaily,
		maxHourly: maxHourly,
	}
	backup.connectAWS()
	backup.backup()
	backup.upload()
	backup.cleanup()
	log.Println("Backup complete")
}

func (s *clusterBackup) connectAWS() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")})
	if err != nil {
		fatalError("Unable to create aws session", err)
	}

	svc := s3.New(sess)
	s.session = sess
	s.client = svc
}

func (s clusterBackup) backup() {
	log.Println("Backing up")
	args := []string{"cluster", "backup", "--file", s.fileName}
	cmd := exec.Command("flynn", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fatalError("Unable to backup cluster", err)
	}
}

func (s clusterBackup) upload() {
	now := time.Now()
	daily := fmt.Sprintf("daily/%s.tar", now.Format("2006-01-02"))
	hourly := fmt.Sprintf("hourly/%s.tar", now.Format("02_15-04-05"))

	s.uploadFile(hourly)
	// TODO don't do this every time
	s.uploadFile(daily)
}

func (s clusterBackup) uploadFile(dest string) {
	log.Printf("Uploading backup to %s", dest)
	uploader := s3manager.NewUploader(s.session)

	file, err := os.Open(s.fileName)
	if err != nil {
		fatalError("Unable to open backup file", err)
	}
	defer file.Close()

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(dest),
		Body:   file,
	})
	if err != nil {
		fatalError("Unable to upload backup", err)
	}
	log.Println("Upload complete")
}

func (s clusterBackup) cleanup() {
	log.Println("Cleanup local files")
	if err := os.Remove(s.fileName); err != nil {
		log.Print("WARNING: Unable to remove local file")
	}

	s.cleanupPrefix("daily", s.maxDaily)
	s.cleanupPrefix("hourly", s.maxHourly)
}

//Sorter interface for key
type Sorter []*s3.Object

func (s Sorter) Len() int {
	return len(s)
}
func (s Sorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Sorter) Less(i, j int) bool {
	return *(s[i].Key) > *(s[j].Key)
}

func (s *clusterBackup) cleanupPrefix(prefix string, maxVersions int) {
	log.Printf("Cleanup %s backup files", prefix)
	resp, err := s.client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(s.bucket),
		Prefix: &prefix,
	})
	if err != nil {
		fatalError("Unable to list items in bucket", err)
	}

	objects := resp.Contents
	sort.Sort(Sorter(objects))

	// Don't clean if not needed
	if len(objects) <= maxVersions {
		return
	}
	destroy := objects[maxVersions:]

	log.Printf("Removing %d files", len(destroy))
	for _, item := range destroy {
		_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(*item.Key),
		})
		if err != nil {
			log.Printf("Unable to delete object %q %v", *item.Key, err)
		}
	}
}

func fatalError(msg string, err error) {
	log.Printf("ERROR: %s", msg)
	log.Fatal(err)
}
