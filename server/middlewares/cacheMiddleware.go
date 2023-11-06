package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-secret-manager/server/api/aws"
	"golang-secret-manager/server/storage"
	"golang-secret-manager/utils/types"
	"net/http"
)

// Before handling the request checking if the value of the secrets in the cache
func GetAllSecretsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Decoding the request body
		defer r.Body.Close()
		var reqBody types.GetAllSecretsRequest
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&reqBody)
		if err != nil {
			http.Error(rw, "Failed to decode JSON data", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		fmt.Println("inside handler1")

		// retriving the cache instance
		cache := storage.GetCacheInstance()
		fmt.Println("inside handler2")

		// extraction body information
		publicKey := reqBody.PublicKey

		// Structure that will be pass to the handler
		toContext := types.FromGetAllSecretsMiddlewareToHandler{
			FoundedAccessLog: nil,
			FoundedSecrets:   nil,
			ArnList:          nil,
			FoundedArnList:   false,
			PublicKey:        publicKey,
			SecretKey:        reqBody.SecretKey,
			Region:           reqBody.Region,
		}

		// checking if user ARN list in cache
		KeyArn := aws.GetCacheARNKey(publicKey)

		value, err := cache.Get(KeyArn)
		if err != nil {
			// was not found in cache, calling to next function
			fmt.Println("MIDDILEWARE: User", publicKey, "arn list was not found in the cache")
			// Setting the pass by value to the context
			ctx := context.WithValue(ctx, "info", &toContext)
			next.ServeHTTP(rw, r.WithContext(ctx))
			return
		}

		ArnList, arnOk := value.([]string)
		if !arnOk {
			// failed to convert the ArnList
			fmt.Print("MIDDILEWARE: failed to convert the ARN list to []string that was in the cache")
			// Setting the pass by value to the context
			ctx := context.WithValue(ctx, "info", &toContext)
			next.ServeHTTP(rw, r.WithContext(ctx))
			return
		}

		allFound := true
		foundedSecrets := make(map[string]types.Secret)
		for _, arn := range ArnList {
			key := aws.GetCacheSecretKey(arn)
			val, err := cache.Get(key)
			if err != nil {
				// not in cache
				allFound = false
				continue
			}

			secret, ok := val.(types.Secret)
			if !ok {
				// failed to convert to secret
				fmt.Println("failed to convert secret:", arn, "to type Secret")
				allFound = false
				continue
			} else {
				foundedSecrets[arn] = secret
			}
		}

		foundedAccessLog := make(map[string][]types.AccessLog)
		for _, arn := range ArnList {
			key := aws.GetCacheAccessKey(arn)
			val, err := cache.Get(key)
			if err != nil {
				// not in cache
				allFound = false
				continue
			}

			accessLog, ok := val.([]types.AccessLog)
			if !ok {
				// failed to convert to secret
				fmt.Println("failed to convert Access Log:", arn, "to type []AccessLog")
				allFound = false
				continue
			} else {
				foundedAccessLog[arn] = accessLog
			}
		}

		if allFound {
			// returning the value from the cache if all the value was in cache
			// sending back to the client the anwser

			var secretList []types.Secret

			for _, value := range foundedSecrets {
				secretList = append(secretList, value)
			}

			toSend := types.GetAllSecretsResponse{
				Secrets:   secretList,
				AccessLog: foundedAccessLog,
			}

			json.NewEncoder(rw).Encode(toSend)
			rw.WriteHeader(http.StatusOK)
			return
		}

		toContext.FoundedAccessLog = foundedAccessLog
		toContext.FoundedSecrets = foundedSecrets
		toContext.FoundedArnList = arnOk
		toContext.ArnList = ArnList

		ctx = context.WithValue(ctx, "info", &toContext)

		// Calling handler
		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}

// Before handling the request checking if the Secret already in cache for fast access
func GetReportMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

	})
}
