package photon_download_workflow

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.temporal.io/sdk/activity"

	app_config "github.com/bikehopper/photon-download-workflow/src/app_config"
)

type PhotonDownloadActivityObject struct{}

type CheckForNewArchiveActivityResult struct {
	NewArchiveAvailable bool
}

type DownloadArchiveActivityResult struct {
	FilePath string
	Etag     string
}

type UploadArchiveActivityParams struct {
	FilePath string
	Etag     string
}

type UploadArchiveActivityResult struct {
	Key string
}

type CreateLatestArchiveActivityParams struct {
	Key string
}

type FetchArchiveResult struct {
	File *os.File
	Etag string
}

func getEtagHttp(url string) (*string, error) {
	latestHeadRes, err := http.Head(url)
	if err != nil {
		return nil, err
	}
	if latestHeadRes.StatusCode > 299 {
		return nil, err
	}

	latestEtag := latestHeadRes.Header.Get("ETag")
	return &latestEtag, nil
}

func fetchArchive(url string) (*FetchArchiveResult, error) {
	file, err := os.CreateTemp("", "photon-db-us-latest.*.tar.bz2")
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, err
	}
	return &FetchArchiveResult{
		File: file,
		Etag: resp.Header.Get("ETag"),
	}, nil
}

func getDatedFileName(filePath string, date time.Time) string {
	fileName := filepath.Base(filePath)
	return date.Format("2006-01-02") + "-" + strings.Replace(fileName, "-latest", "", 1)
}

func (o *PhotonDownloadActivityObject) CheckForNewArchiveActivity(ctx context.Context) (*CheckForNewArchiveActivityResult, error) {
	// logger := activity.GetLogger(ctx)
	conf := app_config.New()

	// Get Etag of latest from Geofabrik
	result := &CheckForNewArchiveActivityResult{
		NewArchiveAvailable: false,
	}
	latestEtag, err := getEtagHttp(conf.ArchiveUrl)
	if err != nil {
		return nil, err
	}

	cfg, _ := config.LoadDefaultConfig(context.TODO())
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Region = conf.S3Region
		o.BaseEndpoint = aws.String(conf.S3EndpointUrl)
		o.UsePathStyle = true
	})

	lastUpdatedArchiveHead, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(conf.Bucket),
		Key:    aws.String(conf.ArchiveKey),
	})
	if err != nil {
		var nf *types.NotFound
		if errors.As(err, &nf) {
			result.NewArchiveAvailable = true
			return result, nil
		}
		return result, err
	}

	lastUpdatedArchiveEtag := lastUpdatedArchiveHead.Metadata["ETag"]
	if *latestEtag != lastUpdatedArchiveEtag {
		result.NewArchiveAvailable = true
	}

	return result, nil
}

func (o *PhotonDownloadActivityObject) DownloadArchiveActivity(ctx context.Context) (*DownloadArchiveActivityResult, error) {
	conf := app_config.New()
	fetchResult, err := fetchArchive(conf.ArchiveUrl)
	if err != nil {
		return nil, err
	}
	defer fetchResult.File.Close()

	result := &DownloadArchiveActivityResult{
		FilePath: fetchResult.File.Name(),
		Etag:     fetchResult.Etag,
	}

	return result, nil
}

func (o *PhotonDownloadActivityObject) UploadArchiveActivity(ctx context.Context, param UploadArchiveActivityParams) (*UploadArchiveActivityResult, error) {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	conf := app_config.New()
	s3Client := s3.NewFromConfig(cfg, func(opt *s3.Options) {
		opt.Region = conf.S3Region
		opt.BaseEndpoint = aws.String(conf.S3EndpointUrl)
		opt.UsePathStyle = true
	})

	datedFileName := getDatedFileName(conf.ArchiveKey, activity.GetInfo(ctx).ScheduledTime)
	objectKey := strings.Replace(conf.ArchiveKey, path.Base(conf.ArchiveKey), datedFileName, 1)
	fileToUpload, err := os.Open(param.FilePath)
	if err != nil {
		return nil, err
	}
	// not really necc. when run in a container but useful otherwise
	defer os.Remove(fileToUpload.Name())
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(conf.Bucket),
		Key:    aws.String(objectKey),
		Body:   fileToUpload,
		Metadata: map[string]string{
			"geofabrik-etag": param.Etag,
		},
	})
	if err != nil {
		return nil, err
	}

	return &UploadArchiveActivityResult{
		Key: objectKey,
	}, nil
}

func (o *PhotonDownloadActivityObject) CreateLatestArchiveActivity(ctx context.Context, params CreateLatestArchiveActivityParams) error {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	conf := app_config.New()
	s3Client := s3.NewFromConfig(cfg, func(opt *s3.Options) {
		opt.Region = conf.S3Region
		opt.BaseEndpoint = aws.String(conf.S3EndpointUrl)
		opt.UsePathStyle = true
	})

	_, err := s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		CopySource: aws.String(filepath.Join(conf.Bucket, params.Key)),
		Bucket:     aws.String(conf.Bucket),
		Key:        aws.String(conf.ArchiveKey),
	})
	if err != nil {
		return err
	}

	return nil
}
