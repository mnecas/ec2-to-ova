package client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AwsClient struct {
	cfg       aws.Config
	ec2Client *ec2.Client
	s3Client  *s3.Client
}

func Build(
	cfg aws.Config,
) AwsClient {
	ec2Client := ec2.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)
	return AwsClient{
		cfg,
		ec2Client,
		s3Client,
	}
}

// GetExportImageTask retrieves information about a specific export AMI.
func (client AwsClient) GetExportImageTask(ctx context.Context, exportImageTask string) (*types.ExportImageTask, error) {
	input := &ec2.DescribeExportImageTasksInput{
		ExportImageTaskIds: []string{
			exportImageTask,
		},
	}

	result, err := client.ec2Client.DescribeExportImageTasks(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe image %s: %w", exportImageTask, err)
	}

	if len(result.ExportImageTasks) == 0 {
		return nil, fmt.Errorf("no image found with ID: %s", exportImageTask)
	}

	return &result.ExportImageTasks[0], nil
}

// GetImageInfo retrieves information about a specific AMI.
func (client AwsClient) GetImageInfo(ctx context.Context, imageId string) (*types.Image, error) {
	input := &ec2.DescribeImagesInput{
		ImageIds: []string{
			imageId,
		},
	}

	result, err := client.ec2Client.DescribeImages(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe image %s: %w", imageId, err)
	}

	if len(result.Images) == 0 {
		return nil, fmt.Errorf("no image found with ID: %s", imageId)
	}

	return &result.Images[0], nil
}

// GetInstanceInfo retrieves information about a specific Instance.
func (client AwsClient) GetInstanceInfo(ctx context.Context, instanceId string) (*types.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			instanceId,
		},
	}

	result, err := client.ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance %s: %w", instanceId, err)
	}
	if len(result.Reservations) == 0 {
		return nil, fmt.Errorf("no instance found with ID: %s", instanceId)
	}
	if len(result.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("no instance found with ID: %s", instanceId)
	}

	return &result.Reservations[0].Instances[0], nil
}

// GetInstanceTypeInfo retrieves information about a specific InstanceType.
func (client AwsClient) GetInstanceTypeInfo(ctx context.Context, instanceType types.InstanceType) (*types.InstanceTypeInfo, error) {
	input := &ec2.DescribeInstanceTypesInput{
		InstanceTypes: []types.InstanceType{
			instanceType,
		},
	}

	result, err := client.ec2Client.DescribeInstanceTypes(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance %s: %w", instanceType, err)
	}
	if len(result.InstanceTypes) == 0 {
		return nil, fmt.Errorf("no instance found with ID: %s", instanceType)
	}

	return &result.InstanceTypes[0], nil
}

// UploadDataToBucket uploads a file to an S3 bucket.
func (client AwsClient) UploadDataToBucket(ctx context.Context, bucket, key, data string) error {
	_, err := client.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte(data)),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to bucket %s: %w", bucket, err)
	}

	return nil
}
