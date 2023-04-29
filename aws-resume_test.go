package main

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
)

func TestAwsResumeStack(t *testing.T) {
	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewAwsResumeStack(app, "MyStack", &AwsResumeStackProps{
		stackDetails: AwsResumeStackConfigs{
			HostedZoneName: "test.com",
			Subdomain:      "resume",
		},
	})

	// THEN
	template := assertions.Template_FromStack(stack, &assertions.TemplateParsingOptions{})

	template.HasResourceProperties(jsii.String("AWS::SQS::Queue"), map[string]interface{}{
		"VisibilityTimeout": 300,
	})
}
