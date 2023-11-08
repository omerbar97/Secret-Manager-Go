package aws

import (
	"context"
	"errors"
	"fmt"
	"golang-secret-manager/types"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// for this package use only
func isErrWithCode(err error, code int) bool {
	if err := errors.Unwrap(err); err != nil {
		awsErr, ok := err.(interface{ HTTPStatusCode() int })
		if ok {
			return awsErr.HTTPStatusCode() == code
		}
	}
	return false
}

type IAWSClient interface {
	GetAllSecrets(nextToken *string) (types.AllSecrets, error)
	GetAccessLog(secretID string, nextToken *string) (types.AllAccessLog, error)
	GetSecretById(secretID string) (*types.Secret, error)
}

type client struct {
	ctx       context.Context
	PublicKey string
	SecretKey string
	Region    string
	Session   *session.Session
}

func NewAWSClient(ctx context.Context, publicKey string, secretKey string, region string) (IAWSClient, error) {
	if region == "" {
		// setting default value
		region = "us-east-1"
	}
	creds := credentials.NewStaticCredentials(publicKey, secretKey, "")
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	})
	if err != nil {
		// failed to create client
		return nil, err
	}
	return &client{
		ctx:       ctx,
		PublicKey: publicKey,
		SecretKey: secretKey,
		Session:   sess,
	}, nil
}

func (c *client) GetAllSecrets(nextToken *string) (types.AllSecrets, error) {
	// this function will retrive all the secrets from the AWS services
	svc := secretsmanager.New(c.Session)

	input := &secretsmanager.ListSecretsInput{
		MaxResults: aws.Int64(100),
		NextToken:  nextToken,
	}

	listSecretOutput := &secretsmanager.ListSecretsOutput{}
	var err error

	var secrets []types.Secret
	var returnValue types.AllSecrets

	for {
		// retriving the secrets
		listSecretOutput, err = svc.ListSecrets(input)
		if err != nil {
			// checking if the error is due to rate limiting status code 429
			if isErrWithCode(err, 429) {
				// rate limiting the api calls
				retryAfter := err.(interface{ RetryDelay() time.Duration }).RetryDelay()
				fmt.Println("Rate limited, retrying after:", retryAfter)
				time.Sleep(retryAfter)
				continue
			} else {
				fmt.Println("Error retriving secrets:", err)
				// returning the nextToken
				returnValue.Secrets = secrets
				returnValue.NextToken = nextToken
				return returnValue, err
			}
		}

		// adding all the secrets to the list
		for _, secret := range listSecretOutput.SecretList {
			var s types.Secret
			s.Name = *secret.Name
			s.ARN = *secret.ARN
			s.CreatedAt = *secret.CreatedDate
			s.LastAccessed = *secret.LastAccessedDate
			secrets = append(secrets, s)
		}

		if nextToken == nil {
			// no more secrets to fetch
			break
		}

		input.NextToken = listSecretOutput.NextToken
	}
	returnValue.Secrets = secrets
	returnValue.NextToken = nil
	return returnValue, nil
}

func (c *client) GetSecretById(secretID string) (*types.Secret, error) {

	svc := secretsmanager.New(c.Session)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	}

	output, err := svc.GetSecretValue(input)
	if err != nil {
		// failed to retrive the secrets
		return nil, err
	}

	secret := &types.Secret{
		Name:         *output.Name,
		CreatedAt:    *output.CreatedDate,
		LastAccessed: time.Time{},
		ARN:          *output.ARN,
		Version:      *output.VersionId,
	}

	return secret, nil
}

func (c *client) GetAccessLog(secretID string, nextToken *string) (types.AllAccessLog, error) {

	// creating the cloudtrail svc
	svc := cloudtrail.New(c.Session)

	input := &cloudtrail.LookupEventsInput{
		LookupAttributes: []*cloudtrail.LookupAttribute{
			{
				AttributeKey:   aws.String("ResourceName"),
				AttributeValue: aws.String(secretID),
			},
		},
		MaxResults: aws.Int64(50),
		NextToken:  nextToken,
	}

	var list []types.AccessLog
	var returnValue types.AllAccessLog

	for {
		result, err := svc.LookupEvents(input)
		if err != nil {
			if !isErrWithCode(err, 429) {
				fmt.Println("Error retriving access log:", err)
				// failed to retrive all accesslog
				returnValue.AccessLog = list
				returnValue.NextToken = nextToken
				return returnValue, err
			}
			// rate limiting the api calls
			retryAfter := err.(interface{ RetryDelay() time.Duration }).RetryDelay()
			fmt.Println("Rate limited, retrying after:", retryAfter)
			time.Sleep(retryAfter)
			continue
		}
		for _, event := range result.Events {
			val := types.AccessLog{
				User:        *event.Username,
				EventTime:   *event.EventTime,
				EventName:   *event.EventName,
				EventSource: *event.EventSource,
			}
			list = append(list, val)
		}
		if result.NextToken == nil {
			// finished
			break
		}
		input.NextToken = result.NextToken
	}

	returnValue.AccessLog = list
	returnValue.NextToken = nil
	return returnValue, nil
}
