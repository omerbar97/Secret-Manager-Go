package middleware

import (
	"context"
	"fmt"
	"golang-secret-manager/api/aws"
	"golang-secret-manager/types"
	GenericEncoding "golang-secret-manager/utils/genericEncoding"
	"golang-secret-manager/utils/storage"
	"net/http"
)

// Before handling the request checking if the value of the secrets in the cache
func GetAllSecretsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Decoding the request body
		defer r.Body.Close()
		reqBody, err := GenericEncoding.JsonBodyDecoder[types.GetAllSecretsRequest](r.Body)
		if err != nil {
			GenericEncoding.WriteJson(rw, http.StatusBadRequest, types.ApiError{Err: "request body didn't matched!", Status: http.StatusBadRequest})
			return
		}

		ctx := r.Context()

		// retriving the cache instance
		cacheInstance := storage.GetCacheInstance()

		// extraction publicKey from body
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
		contextKey := types.GetContextInforamtionKey()
		arnList, err := storage.GetCacheValue[[]string](cacheInstance, KeyArn)
		if err != nil {
			// was not found in cache, calling to next function
			fmt.Println("MIDDILEWARE: User", publicKey, "ARN list was not found in the cache")
			// Setting the pass by value to the context
			ctx := context.WithValue(ctx, contextKey, &toContext)
			next.ServeHTTP(rw, r.WithContext(ctx))
			return
		}

		allFound := true
		foundedSecrets := make(map[string]types.Secret)
		for _, arn := range *arnList {
			key := aws.GetCacheSecretKey(arn)
			val, err := storage.GetCacheValue[types.Secret](cacheInstance, key)
			if err != nil {
				// not in cache
				allFound = false
				continue
			}
			foundedSecrets[arn] = *val
		}

		foundedAccessLog := make(map[string][]types.AccessLog)
		for _, arn := range *arnList {
			key := aws.GetCacheAccessKey(arn)
			val, err := storage.GetCacheValue[[]types.AccessLog](cacheInstance, key)
			if err != nil {
				// not in cache
				allFound = false
				continue
			}
			foundedAccessLog[arn] = *val
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

			if err := GenericEncoding.WriteJson(rw, http.StatusOK, toSend); err != nil {
				fmt.Println("MIDDILEWARE: failed to send back to client information")
			}
			return
		}

		toContext.FoundedAccessLog = foundedAccessLog
		toContext.FoundedSecrets = foundedSecrets
		toContext.FoundedArnList = true
		toContext.ArnList = *arnList

		ctx = context.WithValue(ctx, contextKey, &toContext)

		// Calling handler
		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}

// Before handling the request checking if the Secret already in cache for fast access
func GetReportMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		reqBody, err := GenericEncoding.JsonBodyDecoder[types.GetReportRequest](r.Body)
		if err != nil {
			GenericEncoding.WriteJson(rw, http.StatusBadRequest, types.ApiError{Err: "request body didn't matched!", Status: http.StatusBadRequest})
			return
		}

		ctx := r.Context()

		// retriving the cache instance
		cacheInstance := storage.GetCacheInstance()

		keyForSecret := aws.GetCacheSecretKey(reqBody.SecretID)
		keyForAccessLog := aws.GetCacheAccessKey(reqBody.SecretID)

		allFound := true

		toContext := types.FromGetReportMiddlewareToHandler{
			FoundedSecret:    nil,
			FoundedAccessLog: nil,
			PublicKey:        reqBody.PublicKey,
			SecretKey:        reqBody.SecretKey,
			SecretID:         reqBody.SecretID,
			Region:           reqBody.Region,
		}

		if secret, err := storage.GetCacheValue[types.Secret](cacheInstance, keyForSecret); err != nil {
			// secret not in memory
			fmt.Println("MIDDILEWARE: Secret id:", reqBody.SecretID, "not in cache")
			allFound = false
		} else {
			toContext.FoundedSecret = secret
		}

		if access, err := storage.GetCacheValue[[]types.AccessLog](cacheInstance, keyForAccessLog); err != nil {
			fmt.Println("MIDDILEWARE: Secret id:", reqBody.SecretID, "Access log was not in cache")
			allFound = false
		} else {
			toContext.FoundedAccessLog = *access
		}

		if !allFound {
			// need to call the handler to retrive the missing information
			contextKey := types.GetContextInforamtionKey()
			ctx = context.WithValue(ctx, contextKey, &toContext)
			next.ServeHTTP(rw, r.WithContext(ctx))
		} else {
			// sending to the user the report
			report := aws.GenerateReportStringBySecret(*toContext.FoundedSecret, toContext.FoundedAccessLog)
			toReturn := types.GetReportResponse{
				Report: report,
			}
			if err := GenericEncoding.WriteJson(rw, http.StatusOK, toReturn); err != nil {
				fmt.Println("MIDDILEWARE: failed to send back to client information")
			}
		}
	})
}
