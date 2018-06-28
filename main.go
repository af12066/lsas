package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
)

func main() {
	// TODO Should set some AWS settings by flag
	// FIXME Need to set flexible amount of option
	if len(os.Args) != 3 {
		panic("must set two key=value sets")
	}
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic(err)
	}

	svc := autoscaling.New(cfg)
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		MaxRecords: aws.Int64(100),
	}
	req := svc.DescribeAutoScalingGroupsRequest(input)
	result, err := req.Send()
	if err != nil {
		panic(err)
	}
	tag1 := strings.Split(os.Args[1], "=")
	tag2 := strings.Split(os.Args[2], "=")
	results := []autoscaling.Group{}
	results = append(results, result.AutoScalingGroups...)

	var nt *string
	nt = result.NextToken

	for {
		if nt != nil {
			nt = func(nt *string) *string {
				// fmt.Println("exist more group yet" + aws.StringValue(nt))
				ri := &autoscaling.DescribeAutoScalingGroupsInput{
					MaxRecords: aws.Int64(100),
					NextToken:  nt,
				}
				rreq := svc.DescribeAutoScalingGroupsRequest(ri)
				rresult, err := rreq.Send()
				if err != nil {
					panic(err)
				}
				results = append(results, rresult.AutoScalingGroups...)
				return rresult.NextToken
			}(nt)
		} else {
			break
		}
	}

	filterd := []autoscaling.Group{}
	for _, asg := range results {
		// FIXME Need async output
		var matched int
		for _, tag := range asg.Tags {
			if aws.StringValue(tag.Key) == tag1[0] {
				// fmt.Printf("asg.Tags = %#+v\n", asg.Tags)
				// fmt.Printf("%s = %+v\n", tag1[0], aws.StringValue(tag.Value))
				if aws.StringValue(tag.Value) == tag1[1] {
					matched++
				}
			} else if aws.StringValue(tag.Key) == tag2[0] && aws.StringValue(tag.Value) == tag2[1] {
				// fmt.Printf("%s = %+v\n", tag2[0], aws.StringValue(tag.Value))
				matched++
			}
		}
		if matched == 2 {
			filterd = append(filterd, asg)
		}
	}
	fmt.Println("autoscaling-group-name | desired-capacity | min-size | max-size | Launch-configuration-name")
	// autoscaling group name
	// desired capacity
	// min
	// max
	// launch configuration名
	for _, asg := range filterd {
		fmt.Printf("%s    %d    %d    %d    %s\n",
			aws.StringValue(asg.AutoScalingGroupName),
			aws.Int64Value(asg.DesiredCapacity),
			aws.Int64Value(asg.MaxSize), aws.Int64Value(asg.MinSize),
			aws.StringValue(asg.LaunchConfigurationName))
	}
}
