package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/awslabs/goformation/v5/cloudformation"
	"github.com/awslabs/goformation/v5/cloudformation/ec2"
	"github.com/awslabs/goformation/v5/cloudformation/tags"
	"github.com/urfave/cli/v2"
)

var opts struct {
	Manifest string
	Output   string
	Version  string
	S3       struct {
		Bucket string
		Prefix string
	}
}

func main() {
	app := cli.NewApp()
	app.Usage = "generates s3 templates"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "manifest",
			Usage:       "packer manifest file",
			Required:    true,
			Destination: &opts.Manifest,
		},
		&cli.StringFlag{
			Name:        "output",
			Usage:       "output directory",
			Required:    true,
			Destination: &opts.Output,
		},
		&cli.StringFlag{
			Name:        "s3-bucket",
			Usage:       "target s3 bucket",
			Destination: &opts.S3.Bucket,
			Value:       "sundaeswap-oss",
		},
		&cli.StringFlag{
			Name:        "s3-prefix",
			Usage:       "s3 key prefix",
			Destination: &opts.S3.Prefix,
		},
		&cli.StringFlag{
			Name:        "version",
			Usage:       "build version",
			EnvVars:     []string{"VERSION"},
			Value:       "latest",
			Destination: &opts.Version,
		},
	}
	app.Action = action
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

type Manifest struct {
	Builds []struct {
		ArtifactID string `json:"artifact_id,omitempty"`
	} `json:"builds,omitempty"`
}

func action(_ *cli.Context) error {
	data, err := ioutil.ReadFile(opts.Manifest)
	if err != nil {
		return fmt.Errorf("failed to read manifest, %v: %w", opts.Manifest, err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest, %v: %w", opts.Manifest, err)
	}

	if err := os.MkdirAll(opts.Output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory, %v: %w", opts.Output, err)
	}

	urls := map[string]string{} // region -> url
	for _, build := range manifest.Builds {
		parts := strings.Split(build.ArtifactID, ":")
		if len(parts) != 2 {
			continue
		}

		region, ami := parts[0], parts[1]
		t := makeTemplate(ami, opts.Version)

		data, err := t.YAML()
		if err != nil {
			return fmt.Errorf("failed to generate cloudformation template: %w", err)
		}

		filename := filepath.Join(opts.Output, fmt.Sprintf("alonzo-testnet-%v.template", region))
		if err := ioutil.WriteFile(filename, data, 0644); err != nil {
			return fmt.Errorf("failed to write cloudformation template, %v: %w", filename, err)
		}

		path, _ := filepath.Abs(filename)
		fmt.Printf("wrote %v\n", path)

		url := fmt.Sprintf("https://console.aws.amazon.com/cloudformation/home?region=%v#/stacks/new?stackName=alonzo-testnet&templateURL=https://s3.amazonaws.com/%v",
			region,
			filepath.Join(
				opts.S3.Bucket,
				opts.S3.Prefix,
				filepath.Base(filename),
			),
		)
		urls[region] = url
	}

	if err := makeHTML(urls, opts.Output); err != nil {
		return fmt.Errorf("failed to generate index.html: %w", err)
	}

	return nil
}

//go:embed index.gohtml
var text string

func makeHTML(urls map[string]string, dir string) interface{} {
	type Image struct {
		Region string
		URL    string
	}
	type Data struct {
		Images []Image
	}

	var data Data
	for region, url := range urls {
		data.Images = append(data.Images, Image{
			Region: region,
			URL:    url,
		})
	}

	sort.Slice(data.Images, func(i, j int) bool {
		return data.Images[i].Region < data.Images[j].Region
	})

	t, err := template.New("page").Parse(text)
	if err != nil {
		return fmt.Errorf("failed to render index.html: %w", err)
	}

	f, err := os.OpenFile(filepath.Join(dir, "index.html"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create index.html: %w", err)
	}
	defer f.Close()

	if err := t.Execute(f, data); err != nil {
		return fmt.Errorf("failed to render index.html: %w", err)
	}

	return nil
}

func makeTemplate(ami string, version string) *cloudformation.Template {
	t := cloudformation.NewTemplate()

	t.Description = "launch private alonzo testnet"
	t.Parameters["InstanceType"] = cloudformation.Parameter{
		Type:        "String",
		Description: "instance type to launch with",
		Default:     "t3.xlarge",
		MinLength:   1,
	}
	t.Parameters["AMI"] = cloudformation.Parameter{
		Type:        "AWS::EC2::Image::Id",
		Description: "alonzo-testnet AMI",
		Default:     ami,
		MinLength:   1,
	}
	t.Parameters["VPC"] = cloudformation.Parameter{
		Type:        "AWS::EC2::VPC::Id",
		Description: "vpc to launch instance into",
		MinLength:   1,
	}
	t.Parameters["SecurityGroup"] = cloudformation.Parameter{
		Type:        "AWS::EC2::SecurityGroup::Id",
		Description: "security group to associate with instance",
		MinLength:   1,
	}
	t.Parameters["Subnet"] = cloudformation.Parameter{
		Type:        "AWS::EC2::Subnet::Id",
		Description: "subnet to launch instance into",
		MinLength:   1,
	}
	t.Parameters["AZ"] = cloudformation.Parameter{
		Type:        "AWS::EC2::AvailabilityZone::Name",
		Description: "subnet availability zone",
		MinLength:   1,
	}
	t.Parameters["KeyName"] = cloudformation.Parameter{
		Type:        "AWS::EC2::KeyPair::KeyName",
		Description: "ec2 instance keypair name",
		MinLength:   1,
	}
	t.Parameters["Profile"] = cloudformation.Parameter{
		Type:        "String",
		Description: "optional ec2 instance profile",
	}

	t.Conditions["HasProfile"] = cloudformation.Not(
		[]string{
			cloudformation.Equals("", cloudformation.Ref("Profile")),
		},
	)

	t.Resources["AlonzoTestnet"] = &ec2.Instance{
		IamInstanceProfile: cloudformation.If(
			"HasProfile",
			cloudformation.Ref("Profile"),
			cloudformation.Ref("AWS::NoValue"),
		),
		ImageId:                           cloudformation.Ref("AMI"),
		InstanceInitiatedShutdownBehavior: "terminate",
		InstanceType:                      cloudformation.Ref("InstanceType"),
		KeyName:                           cloudformation.Ref("KeyName"),
		Monitoring:                        true,
		NetworkInterfaces: []ec2.Instance_NetworkInterface{
			{
				AssociatePublicIpAddress: true,
				DeviceIndex:              "0",
				GroupSet:                 []string{cloudformation.Ref("SecurityGroup")},
				SubnetId:                 cloudformation.Ref("Subnet"),
			},
		},
		Tags: []tags.Tag{
			{
				Key:   "Name",
				Value: "alonzo-testnet",
			},
			{
				Key:   "sundaeswap:name",
				Value: "alonzo-testnet",
			},
			{
				Key:   "sundaeswap:ami_id",
				Value: ami,
			},
			{
				Key:   "sundaeswap:version",
				Value: version,
			},
		},
	}
	return t
}
