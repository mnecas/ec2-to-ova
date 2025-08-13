package ova

import (
	"encoding/xml"
	"fmt"
	"path"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func getInstanceName(instance *types.Instance) string {
	for _, tag := range instance.Tags {
		if tag.Key != nil && *tag.Key == "Name" {
			return *tag.Value
		}
	}
	return *instance.InstanceId
}

// FormatToOVF generates an OVF XML string from instance, instance type, and AMI details.
// This version dynamically creates disk and network sections.
func FormatToOVF(exportImageTaskId string, instance *types.Instance, instanceTypeInfo *types.InstanceTypeInfo, image *types.Image) (string, error) {
	ovfID := fmt.Sprintf("export-%s", aws.ToString(instance.InstanceId))

	// --- Dynamic Hardware Configuration ---
	var files []File
	var disks []Disk
	var networks []Network
	var hardwareItems []Item

	// Instance ID counter for hardware items
	itemInstanceID := 1

	// 1. CPU
	hardwareItems = append(hardwareItems, Item{
		InstanceID:      strconv.Itoa(itemInstanceID),
		ResourceType:    3, // CPU
		Description:     "Number of Virtual CPUs",
		AllocationUnits: "hertz * 10^6",
		ElementName:     fmt.Sprintf("%d virtual CPU(s)", aws.ToInt32(instance.CpuOptions.CoreCount)),
		VirtualQuantity: int64(aws.ToInt32(instance.CpuOptions.CoreCount)),
	})
	itemInstanceID++

	// 2. Memory
	hardwareItems = append(hardwareItems, Item{
		InstanceID:      strconv.Itoa(itemInstanceID),
		ResourceType:    4, // Memory
		Description:     "Memory Size",
		AllocationUnits: "byte * 2^20",
		ElementName:     fmt.Sprintf("%dMB of memory", aws.ToInt64(instanceTypeInfo.MemoryInfo.SizeInMiB)),
		VirtualQuantity: aws.ToInt64(instanceTypeInfo.MemoryInfo.SizeInMiB),
	})
	itemInstanceID++

	// 3. IDE Controller
	ideControllerInstanceID := strconv.Itoa(itemInstanceID)
	hardwareItems = append(hardwareItems, Item{
		InstanceID:   ideControllerInstanceID,
		ResourceType: 5, // IDE Controller
		Address:      "0",
		Description:  "IDE Controller",
		ElementName:  "VirtualIDEController 0",
	})
	itemInstanceID++

	// 4. Disks (from Image BlockDeviceMappings)
	for i, bdm := range image.BlockDeviceMappings {
		if bdm.Ebs == nil {
			continue
		}
		diskIndex := i + 1
		diskCapacity := int64(*bdm.Ebs.VolumeSize) * 1024 * 1024 * 1024
		// Create File reference
		fileRefID := fmt.Sprintf("file%d", diskIndex)
		files = append(files, File{
			// export-ami-1545f5ce7236bf35t-dev-sda1.raw
			Href: fmt.Sprintf("%s-dev-%s.raw", exportImageTaskId, path.Base(*bdm.DeviceName)),
			Size: diskCapacity,
			Id:   fileRefID,
		})

		// Create Disk section entry
		diskID := fmt.Sprintf("vmdisk%d", diskIndex)
		disks = append(disks, Disk{
			Capacity:                diskCapacity,
			CapacityAllocationUnits: "byte",
			DiskID:                  diskID,
			FileRef:                 fileRefID,
			Format:                  "http://www.vmware.com/interfaces/specifications/vmdk.html#streamOptimized",
		})

		// Create Virtual Hardware Item for the disk
		hardwareItems = append(hardwareItems, Item{
			InstanceID:      strconv.Itoa(itemInstanceID),
			ResourceType:    17, // Hard Disk
			ElementName:     fmt.Sprintf("Hard Disk %d", diskIndex),
			HostResource:    fmt.Sprintf("ovf:/disk/%s", diskID),
			Parent:          ideControllerInstanceID,
			AddressOnParent: strconv.Itoa(i),
		})
		itemInstanceID++
	}

	// 5. Networks (from Instance NetworkInterfaces)
	for i, netInterface := range instance.NetworkInterfaces {
		networkIndex := i + 1
		networkName := fmt.Sprintf("VM Network %d", networkIndex)
		if netInterface.SubnetId != nil {
			networkName = aws.ToString(netInterface.SubnetId)
		}

		// Create Network section entry
		networks = append(networks, Network{
			Name:        networkName,
			Description: fmt.Sprintf("Network interface %d", networkIndex),
		})

		// Create Virtual Hardware Item for the network adapter
		autoAlloc := true
		hardwareItems = append(hardwareItems, Item{
			InstanceID:          strconv.Itoa(itemInstanceID),
			ResourceType:        10, // Ethernet Adapter
			ResourceSubType:     "E1000",
			ElementName:         fmt.Sprintf("Ethernet %d", networkIndex),
			Description:         fmt.Sprintf("E1000 ethernet adapter on \"%s\"", networkName),
			Connection:          networkName,
			AutomaticAllocation: &autoAlloc,
		})
		itemInstanceID++
	}

	// --- Determine OS Type ---
	osType := "otherLinux64Guest" // Default
	if image.PlatformDetails != nil {
		if *image.PlatformDetails == "Windows" {
			osType = "windows9_64Guest"
		} else if *image.PlatformDetails == "Red Hat Enterprise Linux" {
			osType = "rhel8_64Guest"
		}
	}

	// --- Assemble the final OVF Envelope ---
	envelope := Envelope{
		Xmlns:          "http://schemas.dmtf.org/ovf/envelope/1",
		Cim:            "http://schemas.dmtf.org/wbem/wscim/1/common",
		Ovf:            "http://schemas.dmtf.org/ovf/envelope/1",
		Rasd:           "http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData",
		Vmw:            "http://www.vmware.com/schema/ovf",
		Vssd:           "http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData",
		Xsi:            "http://www.w3.org/2001/XMLSchema-instance",
		References:     References{Files: files},
		DiskSection:    DiskSection{Info: "List of the virtual disks", Disks: disks},
		NetworkSection: NetworkSection{Info: "The list of logical networks", Networks: networks},
		VirtualSystem: VirtualSystem{
			ID:   ovfID,
			Info: "A virtual machine",
			Name: getInstanceName(instance),
			OperatingSystem: OperatingSystemSection{
				ID:          94,
				OsType:      osType,
				Info:        "The kind of installed guest operating system",
				Description: aws.ToString(image.Description),
			},
			VirtualHardware: VirtualHardwareSection{
				Info: "Virtual hardware requirements",
				System: System{
					ElementName:             "Virtual Hardware Family",
					InstanceID:              0,
					VirtualSystemIdentifier: ovfID,
					VirtualSystemType:       "vmx-07",
				},
				Items: hardwareItems,
			},
		},
	}

	output, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return "", err
	}

	return `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + string(output), nil
}
