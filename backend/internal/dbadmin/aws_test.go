package dbadmin

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type fakeRDSClient struct {
	describe func(*rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error)
}

func (c fakeRDSClient) DescribeDBInstances(_ context.Context, input *rds.DescribeDBInstancesInput, _ ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	return c.describe(input)
}

type fakeSecretsManagerClient struct {
	describe func(*secretsmanager.DescribeSecretInput) (*secretsmanager.DescribeSecretOutput, error)
	get      func(*secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
	put      func(*secretsmanager.PutSecretValueInput) (*secretsmanager.PutSecretValueOutput, error)
}

func (c fakeSecretsManagerClient) DescribeSecret(_ context.Context, input *secretsmanager.DescribeSecretInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.DescribeSecretOutput, error) {
	return c.describe(input)
}

func (c fakeSecretsManagerClient) GetSecretValue(_ context.Context, input *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return c.get(input)
}

func (c fakeSecretsManagerClient) PutSecretValue(_ context.Context, input *secretsmanager.PutSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error) {
	return c.put(input)
}

func prepareAWSClientTest(t *testing.T) {
	t.Helper()
	t.Setenv("AWS_REGION", "ap-northeast-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
}

func stubRDSClient(describe func(*rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error)) func() {
	original := newRDSClient
	newRDSClient = func(aws.Config) rdsClient {
		return fakeRDSClient{describe: describe}
	}
	return func() {
		newRDSClient = original
	}
}

func stubSecretsManagerClient(client fakeSecretsManagerClient) func() {
	original := newSecretsManagerClient
	newSecretsManagerClient = func(aws.Config) secretsManagerClient {
		return client
	}
	return func() {
		newSecretsManagerClient = original
	}
}

func TestGetRDSEndpointReturnsEndpoint(t *testing.T) {
	prepareAWSClientTest(t)
	restore := stubRDSClient(func(*rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
		return &rds.DescribeDBInstancesOutput{
			DBInstances: []rdstypes.DBInstance{{
				Endpoint: &rdstypes.Endpoint{
					Address: aws.String("db.example"),
					Port:    aws.Int32(5432),
				},
			}},
		}, nil
	})
	defer restore()

	endpoint, err := GetRDSEndpoint(context.Background(), "app-db")
	if err != nil {
		t.Fatal(err)
	}
	if endpoint.Host != "db.example" || endpoint.Port != 5432 {
		t.Fatalf("unexpected endpoint: %#v", endpoint)
	}
}

func TestAWSConfigHidesLoadError(t *testing.T) {
	t.Setenv("AWS_PROFILE", "profile-that-does-not-exist-for-test")
	t.Setenv("AWS_REGION", "ap-northeast-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")

	_, err := awsConfig(context.Background())
	if err == nil || err.Error() != "AWS request failed; check operator credentials, region and permissions" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetRDSEndpointHidesAWSError(t *testing.T) {
	prepareAWSClientTest(t)
	restore := stubRDSClient(func(*rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
		return nil, errors.New("raw aws error")
	})
	defer restore()

	_, err := GetRDSEndpoint(context.Background(), "app-db")
	if err == nil || err.Error() != "AWS request failed; check operator credentials, region and permissions" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetRDSEndpointRejectsMissingEndpoint(t *testing.T) {
	prepareAWSClientTest(t)
	restore := stubRDSClient(func(*rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
		return &rds.DescribeDBInstancesOutput{
			DBInstances: []rdstypes.DBInstance{{}},
		}, nil
	})
	defer restore()

	_, err := GetRDSEndpoint(context.Background(), "app-db")
	if err == nil || err.Error() != "AWS request failed; check operator credentials, region and permissions" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCurrentSecretVersionExistsDetectsAWSCURRENT(t *testing.T) {
	prepareAWSClientTest(t)
	restore := stubSecretsManagerClient(fakeSecretsManagerClient{
		describe: func(*secretsmanager.DescribeSecretInput) (*secretsmanager.DescribeSecretOutput, error) {
			return &secretsmanager.DescribeSecretOutput{
				VersionIdsToStages: map[string][]string{
					"version-1": {"AWSPREVIOUS"},
					"version-2": {"AWSCURRENT"},
				},
			}, nil
		},
	})
	defer restore()

	exists, err := CurrentSecretVersionExists(context.Background(), "db/app_user")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("expected current secret version")
	}
}

func TestCurrentSecretVersionExistsReturnsFalseWhenCurrentVersionIsMissing(t *testing.T) {
	prepareAWSClientTest(t)
	restore := stubSecretsManagerClient(fakeSecretsManagerClient{
		describe: func(*secretsmanager.DescribeSecretInput) (*secretsmanager.DescribeSecretOutput, error) {
			return &secretsmanager.DescribeSecretOutput{
				VersionIdsToStages: map[string][]string{
					"version-1": {"AWSPREVIOUS"},
				},
			}, nil
		},
	})
	defer restore()

	exists, err := CurrentSecretVersionExists(context.Background(), "db/app_user")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("did not expect current secret version")
	}
}

func TestCurrentSecretVersionExistsHidesAWSError(t *testing.T) {
	prepareAWSClientTest(t)
	restore := stubSecretsManagerClient(fakeSecretsManagerClient{
		describe: func(*secretsmanager.DescribeSecretInput) (*secretsmanager.DescribeSecretOutput, error) {
			return nil, errors.New("raw aws error")
		},
	})
	defer restore()

	_, err := CurrentSecretVersionExists(context.Background(), "db/app_user")
	if err == nil || err.Error() != "AWS request failed; check operator credentials, region and permissions" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetSecretStringReturnsSecretString(t *testing.T) {
	prepareAWSClientTest(t)
	restore := stubSecretsManagerClient(fakeSecretsManagerClient{
		get: func(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
			if *input.SecretId != "db/app_user" {
				t.Fatalf("unexpected secret id: %s", *input.SecretId)
			}
			return &secretsmanager.GetSecretValueOutput{SecretString: aws.String(`{"username":"app_user"}`)}, nil
		},
	})
	defer restore()

	secret, err := GetSecretString(context.Background(), "db/app_user")
	if err != nil {
		t.Fatal(err)
	}
	if secret != `{"username":"app_user"}` {
		t.Fatalf("unexpected secret: %s", secret)
	}
}

func TestGetSecretStringRejectsMissingSecretString(t *testing.T) {
	prepareAWSClientTest(t)
	restore := stubSecretsManagerClient(fakeSecretsManagerClient{
		get: func(*secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
			return &secretsmanager.GetSecretValueOutput{}, nil
		},
	})
	defer restore()

	_, err := GetSecretString(context.Background(), "db/app_user")
	if err == nil || err.Error() != "AWS request failed; check operator credentials, region and permissions" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetSecretStringHidesAWSError(t *testing.T) {
	prepareAWSClientTest(t)
	restore := stubSecretsManagerClient(fakeSecretsManagerClient{
		get: func(*secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
			return nil, errors.New("raw aws error")
		},
	})
	defer restore()

	_, err := GetSecretString(context.Background(), "db/app_user")
	if err == nil || err.Error() != "AWS request failed; check operator credentials, region and permissions" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPutSecretStringStoresSecretString(t *testing.T) {
	prepareAWSClientTest(t)
	var storedID, storedSecret string
	restore := stubSecretsManagerClient(fakeSecretsManagerClient{
		put: func(input *secretsmanager.PutSecretValueInput) (*secretsmanager.PutSecretValueOutput, error) {
			storedID = *input.SecretId
			storedSecret = *input.SecretString
			return &secretsmanager.PutSecretValueOutput{}, nil
		},
	})
	defer restore()

	if err := PutSecretString(context.Background(), "db/app_user", `{"username":"app_user"}`); err != nil {
		t.Fatal(err)
	}
	if storedID != "db/app_user" || storedSecret != `{"username":"app_user"}` {
		t.Fatalf("unexpected stored secret: id=%s secret=%s", storedID, storedSecret)
	}
}

func TestPutSecretStringHidesAWSError(t *testing.T) {
	prepareAWSClientTest(t)
	restore := stubSecretsManagerClient(fakeSecretsManagerClient{
		put: func(*secretsmanager.PutSecretValueInput) (*secretsmanager.PutSecretValueOutput, error) {
			return nil, errors.New("raw aws error")
		},
	})
	defer restore()

	err := PutSecretString(context.Background(), "db/app_user", `{"username":"app_user"}`)
	if err == nil || err.Error() != "AWS request failed; check operator credentials, region and permissions" {
		t.Fatalf("unexpected error: %v", err)
	}
}
