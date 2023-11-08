package aws

import (
	"context"
	"fmt"
	"golang-secret-manager/types"
	"golang-secret-manager/utils/storage"
	"log"
)

// Api for handling the aws secretmanger request

// Key Generator for cache
func GetCacheSecretKey(secretID string) string {
	return "secret" + secretID
}
func GetCacheAccessKey(secretID string) string {
	return "access" + secretID
}
func GetCacheARNKey(publicKey string) string {
	return "arnlst" + publicKey
}

func RetriveAllSecretsWithAccessLog(ctx context.Context, publicKey string, secretKey string, region string) (*types.AllSecretWithAccessLog, error) {
	// creating the AWS client
	client, err := NewAWSClient(ctx, publicKey, secretKey, region)
	if err != nil {
		// failed to create AWSClient
		log.Println("API-AWS: failed to create AWSClient")
		return nil, err
	}

	cacheInstance := storage.GetCacheInstance()

	trys := 5
	var nextToken *string = nil
	var allSecrets []types.Secret
	for {
		result, err := client.GetAllSecrets(nextToken)
		if err != nil {
			// failed to retrive all the secrets
			if trys == 0 {
				// failed to retrive all secrets
				return nil, err
			}
			trys--
			continue
		} else if result.NextToken == nil {
			// got all secrets
			for _, secret := range result.Secrets {
				// caching the secrets
				key := GetCacheSecretKey(secret.ARN)
				storage.SetCacheValue[types.Secret](cacheInstance, key, secret)
				if err != nil {
					log.Println("API-AWS: failed to cache the secret", secret.ARN)
				}
			}
			allSecrets = append(allSecrets, result.Secrets...)
			break
		}

		for _, secret := range result.Secrets {
			key := GetCacheSecretKey(secret.ARN)
			log.Println("found secret: ", secret.ARN) // TEST
			storage.SetCacheValue[types.Secret](cacheInstance, key, secret)
			if err != nil {
				log.Println("API-AWS: failed to cache the secret", secret.ARN)
			}
		}
		nextToken = result.NextToken
		allSecrets = append(allSecrets, result.Secrets...)
	}

	// for each secrets retriving the access log
	accessLogMap := make(map[string][]types.AccessLog)
	for _, secret := range allSecrets {
		accesslog, err := getAccessLog(client, secret.ARN)
		if err != nil {
			// failed to retrived all the access log
			continue
		}
		accessLogMap[secret.ARN] = accesslog
		key := GetCacheAccessKey(secret.ARN)
		err = storage.SetCacheValue[[]types.AccessLog](cacheInstance, key, accesslog)
		if err != nil {
			log.Println(err.Error())
		}
	}

	// saving the ARN list for each user that request it
	key := GetCacheARNKey(publicKey)
	lst := createARNList(allSecrets)

	// caching the value
	err = storage.SetCacheValue[[]string](cacheInstance, key, lst)
	if err != nil {
		// failed to cache the value printing the error
		log.Println(err.Error())
	}

	var retVal types.AllSecretWithAccessLog
	retVal.Secrets = allSecrets
	retVal.AccessLog = accessLogMap
	return &retVal, nil
}

func createARNList(secrets []types.Secret) []string {
	var lst []string
	for _, s := range secrets {
		lst = append(lst, s.ARN)
	}
	return lst
}

func getAccessLog(client IAWSClient, secretID string) ([]types.AccessLog, error) {
	var accessLogList []types.AccessLog
	var nextToken *string = nil
	trys := 5
	for {
		accessLogs, err := client.GetAccessLog(secretID, nextToken)
		if err != nil {
			// failed to retrive all
			if trys == 0 {
				// failed to retrive all secrets
				log.Println("API-AWS: failed to retrive all access log to secret id: ", secretID)
				return nil, err
			}
			trys--
			continue
		} else if accessLogs.NextToken == nil {
			// retrive all the accesslog
			accessLogList = append(accessLogList, accessLogs.AccessLog...)
			break
		}

		// adding all to the global variable
		nextToken = accessLogs.NextToken
		accessLogList = append(accessLogList, accessLogs.AccessLog...)
	}
	return accessLogList, nil
}

func GetSecretByIdWithAccessLog(ctx context.Context, publicKey string, secretKey string, secretID string, region string) (*types.SingleSecretWithAccessLog, error) {
	fmt.Println("API-AWS: Getting Report For", secretID)
	client, err := NewAWSClient(ctx, publicKey, secretKey, region)
	if err != nil {
		// failed to create AWSClient
		log.Println("API-AWS: failed to create AWSClient")
	}

	cache := storage.GetCacheInstance()

	// defining number of trys
	trys := 5
	var secret *types.Secret
	for {
		// getting the secrets
		secret, err = client.GetSecretById(secretID)
		if err != nil {
			// failed to retrive the secret
			trys--
		} else {
			// success
			break
		}
		if trys == 0 {
			// failed to retrive after all succes
			return nil, err
		}
		trys--
	}

	// getting the secret accesslog
	accessLogList, err := getAccessLog(client, secretID)
	if err != nil {
		// failed to retrive all access log
		return nil, err
	}

	// caching the access log
	key := GetCacheAccessKey(secretID)
	err = storage.SetCacheValue[[]types.AccessLog](cache, key, accessLogList)
	if err != nil {
		log.Println("API-AWS: failed to cache the Access Log of secret", secretID)
	}

	// The func GetSecretById won't return lastAccessTime to the secret
	if len(accessLogList) > 0 {
		// updating the lastAccess
		lastUse := accessLogList[0].EventTime
		for _, log := range accessLogList {
			if log.EventTime.After(lastUse) {
				lastUse = log.EventTime
			}
		}
		secret.LastAccessed = lastUse
	}

	// caching the secret
	key = GetCacheSecretKey(secretID)
	err = storage.SetCacheValue[types.Secret](cache, key, *secret)
	if err != nil {
		log.Println("API-AWS: failed to cache the secret", secretID)
	}

	toReturn := types.SingleSecretWithAccessLog{
		Secret:    *secret,
		AccessLog: accessLogList,
	}

	return &toReturn, nil
}

func GenerateReportStringBySecret(secret types.Secret, accessLog []types.AccessLog) string {
	fmt.Println("API-AWS: Generating Report For", secret.ARN)
	// MetaData
	report := " # Secret Metadata: \n"
	report += fmt.Sprintf(" - Secret Name: %s\n", secret.Name)
	report += fmt.Sprintf(" - Secret Created At: %s\n", secret.CreatedAt)
	report += fmt.Sprintf(" - Secret Last Accessed: %s\n", secret.LastAccessed)
	report += fmt.Sprintf(" - Secret ARN: %s\n", secret.ARN)

	// AccessLog
	report += " - Secret Access Log: \n"
	for _, accessLog := range accessLog {
		report += fmt.Sprintf("{\n	User: %s\n", accessLog.User)
		report += fmt.Sprintf("	Event Time: %s\n", accessLog.EventTime)
		report += fmt.Sprintf("	Event Name: %s\n", accessLog.EventName)
		report += fmt.Sprintf("	Event Source: %s\n}\n", accessLog.EventSource)
	}
	return report
}
