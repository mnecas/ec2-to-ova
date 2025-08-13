# Setup of the AWS

```bash
aws iam create-role --role-name vmimport --assume-role-policy-document file://vmimport-trust-policy.json
```

Important! Please update the vmimport-role-policy.json with you bucket.
```bash
aws iam put-role-policy --role-name vmimport --policy-name vmimport --policy-document file://vmimport-role-policy.json
```

If there are no changes in the MTV inventory Run
```bash
aws storagegateway refresh-cache --file-share-arn arn:aws:storagegateway:us-east-1:441275399559:share/share-BA5875D1
```
## Documentations
- Configuring storage gateway in private VPC https://docs.aws.amazon.com/storagegateway/latest/vgw/gateway-private-link.html
- Troubleshooting of the storage gateway creation: https://docs.aws.amazon.com/filegateway/latest/files3/troubleshooting-gateway-activation.html
- Exporting AMI image: https://docs.aws.amazon.com/vm-import/latest/userguide/vmexport_image.html
- Role for exporting AMI image: https://docs.aws.amazon.com/vm-import/latest/userguide/required-permissions.html#vmimport-role
