package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/mnecas/ec2-to-ova/pkg/client"
	"github.com/mnecas/ec2-to-ova/pkg/ova"
	"os"
)

func main() {
	// --- Argument Parsing ---
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <AMI_ID>")
		os.Exit(1)
	}
	exportImageTaskId := os.Args[1]

	// --- AWS Configuration ---
	// The config.LoadDefaultConfig function will load credentials and region
	// from your environment (e.g., ~/.aws/credentials, ~/.aws/config, or environment variables).
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to load AWS SDK config: %v\n", err)
		os.Exit(1)
	}

	// --- EC2 Client Initialization ---
	aws := client.Build(cfg)
	exportAmiTask, err := aws.GetExportImageTask(context.TODO(), exportImageTaskId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get export AMI: %v\n", err)
		os.Exit(1)
	}
	imageInfo, err := aws.GetImageInfo(context.TODO(), *exportAmiTask.ImageId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get image info: %v\n", err)
		os.Exit(1)
	}
	instanceInfo, err := aws.GetInstanceInfo(context.TODO(), *imageInfo.SourceInstanceId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get instance info: %v\n", err)
		os.Exit(1)
	}
	instanceTypeInfo, err := aws.GetInstanceTypeInfo(context.TODO(), instanceInfo.InstanceType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get instance type info: %v\n", err)
		os.Exit(1)
	}

	ovf, err := ova.FormatToOVF(exportImageTaskId, instanceInfo, instanceTypeInfo, imageInfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to format OVF: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(ovf)
	path := fmt.Sprintf("%s%s", *exportAmiTask.S3ExportLocation.S3Prefix, "vm.ovf")
	err = aws.UploadDataToBucket(context.TODO(), *exportAmiTask.S3ExportLocation.S3Bucket, path, ovf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to format OVF: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done")
}
