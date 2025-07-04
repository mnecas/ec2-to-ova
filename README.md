# EC2 to OVA Converter

This command-line tool facilitates the conversion of an AWS EC2 AMI export task into an OVF (Open Virtualization Format) file. It retrieves metadata about the exported AMI, the source instance, and instance type to generate a valid OVF descriptor.

## Overview
When you export an AMI from AWS EC2, it's typically stored in an S3 bucket as a disk image (e.g., VMDK or VHD). To use this image in virtualization platforms like VMware or VirtualBox, you often need an OVF file that describes the virtual machine's hardware and configuration.
This tool automates the process of generating that OVF file by:
- Fetching details of the AMI export task.
- Gathering information about the original EC2 instance from which the AMI was created.
- Collecting specifications of the instance type.
- Assembling this information into a standard OVF file format.

## Prerequisites
Before you can use this tool, ensure you have the following set up:

1. Go: You need to have Go installed on your system to build and run the application.
2. AWS Account: You must have an active AWS account.
3. AWS Credentials: Your AWS credentials must be configured on the machine where you're running the tool. The application uses the default AWS SDK credential chain, which looks for credentials in:
    - Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN).
    - The shared credentials file (~/.aws/credentials).
    - The shared configuration file (~/.aws/config).
4. An Exported AMI Task: You must have already initiated an "Export Image" task in the EC2 console and have the ExportImageTaskId (e.g., `export-i-0123456789abcdef0`).

## How to Create an AMI and Export it
Here is an example of how to create an AMI from an existing EC2 instance and then export it using the AWS CLI.

Create an image from an instance:
```bash
aws ec2 create-image \
--instance-id i-055848feba9b2fe05 \
--name "my-web-server" \
--description "My web server image" \
--no-reboot
```
Export the created image to an S3 bucket:

```bash
aws ec2 export-image \
--role-name mnecas-ec2-export-role \
--image-id ami-05b6f33d70f7a2ad5 \
--disk-image-format raw \
--s3-export-location S3Bucket=mnecas,S3Prefix=exports/
```

This command will return an ExportImageTaskId which you can then use with this tool.

## Usage
The generated OVF content will be printed to the standard output. You can redirect this to a file:
```
go run main.go export-i-01a2b3c4d5e6f7g8h > my-virtual-machine.ovf
```
