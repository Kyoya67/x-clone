package dbaccess

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type rdsClient interface {
	DescribeDBInstances(context.Context, *rds.DescribeDBInstancesInput, ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
}

type secretsManagerClient interface {
	DescribeSecret(context.Context, *secretsmanager.DescribeSecretInput, ...func(*secretsmanager.Options)) (*secretsmanager.DescribeSecretOutput, error)
	GetSecretValue(context.Context, *secretsmanager.GetSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
	PutSecretValue(context.Context, *secretsmanager.PutSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error)
}

var (
	newRDSClient            = func(cfg aws.Config) rdsClient { return rds.NewFromConfig(cfg) }
	newSecretsManagerClient = func(cfg aws.Config) secretsManagerClient { return secretsmanager.NewFromConfig(cfg) }
)

func awsConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, errors.New("AWS request failed; check operator credentials, region and permissions")
	}
	return cfg, nil
}

func GetRDSEndpoint(ctx context.Context, instance string) (RDSEndpoint, error) {
	var rdsEndpoint RDSEndpoint
	cfg, err := awsConfig(ctx)
	if err != nil {
		return rdsEndpoint, err
	}
	out, err := newRDSClient(cfg).DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instance),
	})
	if err != nil || len(out.DBInstances) == 0 || out.DBInstances[0].Endpoint == nil {
		return rdsEndpoint, errors.New("AWS request failed; check operator credentials, region and permissions")
	}
	endpoint := out.DBInstances[0].Endpoint
	if endpoint.Address != nil {
		rdsEndpoint.Host = *endpoint.Address
	}
	if endpoint.Port != nil {
		rdsEndpoint.Port = int(*endpoint.Port)
	}
	return rdsEndpoint, nil
}

func CurrentSecretVersionExists(ctx context.Context, secretID string) (bool, error) {
	cfg, err := awsConfig(ctx)
	if err != nil {
		return false, err
	}
	out, err := newSecretsManagerClient(cfg).DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(secretID),
	})
	if err != nil {
		return false, errors.New("AWS request failed; check operator credentials, region and permissions")
	}
	for _, stages := range out.VersionIdsToStages {
		for _, stage := range stages {
			if stage == "AWSCURRENT" {
				return true, nil
			}
		}
	}
	return false, nil
}

func GetSecretString(ctx context.Context, secretID string) (string, error) {
	cfg, err := awsConfig(ctx)
	if err != nil {
		return "", err
	}
	out, err := newSecretsManagerClient(cfg).GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	})
	if err != nil || out.SecretString == nil {
		return "", errors.New("AWS request failed; check operator credentials, region and permissions")
	}
	return *out.SecretString, nil
}

func PutSecretString(ctx context.Context, secretID, secretString string) error {
	cfg, err := awsConfig(ctx)
	if err != nil {
		return err
	}
	_, err = newSecretsManagerClient(cfg).PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(secretID),
		SecretString: aws.String(secretString),
	})
	if err != nil {
		return errors.New("AWS request failed; check operator credentials, region and permissions")
	}
	return nil
}
