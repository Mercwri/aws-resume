package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	acm "github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	cloudfront "github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	origins "github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	route53 "github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	route53targets "github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	s3 "github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	s3assets "github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"

	s3deploy "github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type AwsResumeStackConfigs struct {
	HostedZoneName string `field:"optional"`
	Subdomain      string `field:"optional"`
}

type AwsResumeStackProps struct {
	awscdk.StackProps
	stackDetails AwsResumeStackConfigs
}

func NewAwsResumeStack(scope constructs.Construct, id string, props *AwsResumeStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	hostedZoneName := props.stackDetails.HostedZoneName
	subdomain := props.stackDetails.Subdomain
	oai := cloudfront.NewOriginAccessIdentity(stack, jsii.String("OAI"), &cloudfront.OriginAccessIdentityProps{})

	resumeBucket := s3.NewBucket(stack, jsii.String("ResumeBucket"), &s3.BucketProps{
		BlockPublicAccess: s3.BlockPublicAccess_BLOCK_ALL(),
		PublicReadAccess:  jsii.Bool(false),
		Versioned:         jsii.Bool(true),
	})
	resumeBucket.GrantRead(oai, nil)

	cloudfrontBehavior := &cloudfront.BehaviorOptions{
		Origin: origins.NewS3Origin(resumeBucket, &origins.S3OriginProps{
			OriginId:             jsii.String("CFS3Access"),
			OriginAccessIdentity: oai,
		}),
		ViewerProtocolPolicy: cloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
	}
	hostedZone := route53.HostedZone_FromLookup(stack, jsii.String("HostedZone"), &route53.HostedZoneProviderProps{
		DomainName:  jsii.String(hostedZoneName),
		PrivateZone: jsii.Bool(false),
	})
	tlsCert := acm.NewCertificate(stack, jsii.String("ResumeSiteCert"), &acm.CertificateProps{
		DomainName: jsii.String(fmt.Sprintf("%s.%s", subdomain, hostedZoneName)),
		Validation: acm.CertificateValidation_FromDns(hostedZone),
	})
	cloudFrontDistro := cloudfront.NewDistribution(stack, jsii.String("ResumeDistro"), &cloudfront.DistributionProps{
		DefaultRootObject: jsii.String("index.html"),
		DefaultBehavior:   cloudfrontBehavior,
		Certificate:       tlsCert,
		DomainNames:       &[]*string{jsii.String(fmt.Sprintf("%s.%s", subdomain, hostedZoneName))},
	})

	endpoint := route53.NewARecord(stack, jsii.String("DNS"), &route53.ARecordProps{
		Zone:       hostedZone,
		RecordName: jsii.String(subdomain),
		Target:     route53.RecordTarget_FromAlias(route53targets.NewCloudFrontTarget(cloudFrontDistro)),
	})
	s3deploy.NewBucketDeployment(stack, jsii.String("ResumeContent"), &s3deploy.BucketDeploymentProps{
		DestinationBucket: resumeBucket,
		Sources: &[]s3deploy.ISource{
			s3deploy.Source_Asset(jsii.String("./hugo/public"), &s3assets.AssetOptions{}),
		},
		Distribution: cloudFrontDistro,
		DistributionPaths: &[]*string{
			jsii.String("/*"),
		},
	})
	awscdk.NewCfnOutput(stack, jsii.String("Route53Endpoint"), &awscdk.CfnOutputProps{
		Value: endpoint.DomainName(),
	})
	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewAwsResumeStack(app, "AwsResumeStack", &AwsResumeStackProps{
		awscdk.StackProps{
			Env: env(),
		},
		AwsResumeStackConfigs{
			HostedZoneName: "mikeell.com",
			Subdomain:      "resume",
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
