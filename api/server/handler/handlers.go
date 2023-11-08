package handler

import (
	"fmt"
	"golang-secret-manager/api/aws"
	"golang-secret-manager/types"
	GenericEncoding "golang-secret-manager/utils/genericEncoding"
	"log"
	"net/http"
)

// Transforitm the apiHandler to the http.HandlerFunc
func MakeHTTPHandleFuncDecoder(handler types.ApiHandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if err := handler(rw, r); err != nil {
			// got error
			apiErr, ok := err.(*types.ApiError)
			if !ok {
				// some other error
				if err := GenericEncoding.WriteJson(rw, http.StatusInternalServerError, types.ApiError{Err: "internal error", Status: http.StatusInternalServerError}); err != nil {
					log.Printf("failed to write json to client %v", err)
				}
				return
			}
			// apiErr is type apiError
			if err := GenericEncoding.WriteJson(rw, apiErr.Status, apiErr.Err); err != nil {
				log.Printf("failed to write json to client %v", err)
			}
		}
	}
}

func GetAllSecretsHandlers(rw http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	ctx := r.Context()
	contextKey := types.GetContextInforamtionKey()
	fromContext, ok := ctx.Value(contextKey).(*types.FromGetAllSecretsMiddlewareToHandler)
	if !ok {
		fmt.Println("HANDLER: failed to convert the context value to the handler")
		return &types.ApiError{Err: "Error Converting the context to the handler", Status: http.StatusInternalServerError}
	}

	if !fromContext.FoundedArnList {
		// there was not ArnList for that user in the cache
		val, err := aws.RetriveAllSecretsWithAccessLog(ctx, fromContext.PublicKey, fromContext.SecretKey, fromContext.Region)
		if err != nil {
			// failed to retrive all of them, return bad request
			fmt.Println("HANDLER: failed to retrive all the secrets and access log")
			return &types.ApiError{Err: "Error while trying to retrive all Secrets + Access Logs", Status: http.StatusInternalServerError}
		}

		// sending back to the client the anwser
		toSend := types.GetAllSecretsResponse{
			Secrets:   val.Secrets,
			AccessLog: val.AccessLog,
		}

		if err = GenericEncoding.WriteJson(rw, http.StatusOK, toSend); err != nil {
			// failed sending back to client
			fmt.Println("HANDLER: failed to send back to client information")
		}
		return nil
	}

	// maybe some of the data was found, for what not using the api to retrive the
	// missing ones
	for _, arn := range fromContext.ArnList {
		if _, ok := fromContext.FoundedSecrets[arn]; !ok {
			// key wasn't found in the map, using the AWS api to retrive it
			fromApi, err := aws.GetSecretByIdWithAccessLog(
				ctx,
				fromContext.PublicKey,
				fromContext.SecretKey,
				arn,
				fromContext.Region)
			if err != nil {
				// failed to retrive the information
				fmt.Println("HANDLER: failed to retrive infromation from API about arn:", arn)
				continue
			}
			fromContext.FoundedSecrets[arn] = fromApi.Secret
			fromContext.FoundedAccessLog[arn] = fromApi.AccessLog
		}
	}

	var secretList []types.Secret
	for _, val := range fromContext.FoundedSecrets {
		secretList = append(secretList, val)
	}

	toSend := types.GetAllSecretsResponse{
		Secrets:   secretList,
		AccessLog: fromContext.FoundedAccessLog,
	}
	if err := GenericEncoding.WriteJson(rw, http.StatusOK, toSend); err != nil {
		// failed sending back to client
		fmt.Println("HANDLER: failed to send back to client information")
	}
	return nil
}

func GetReportsHandler(rw http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	ctx := r.Context()
	contextKey := types.GetContextInforamtionKey()
	fromContext, ok := ctx.Value(contextKey).(*types.FromGetReportMiddlewareToHandler)
	if !ok {
		fmt.Println("HANDLER: failed to convert the context value to the handler")
		return &types.ApiError{Err: "Error Converting the context to the handler", Status: http.StatusInternalServerError}
	}

	if fromContext.FoundedSecret == nil {
		// retriving the secret from AWS api
		secret, err := aws.GetSecretById(ctx, fromContext.PublicKey, fromContext.SecretKey, fromContext.SecretID, fromContext.Region)
		if err != nil {
			// failed to retrive secret from AWS api
			return &types.ApiError{Err: "failed to retrive Secret from API", Status: http.StatusBadRequest}
		}
		fromContext.FoundedSecret = secret
	}

	if fromContext.FoundedAccessLog == nil {
		// retriving the access log from AWS api
		access, err := aws.GetAccessLog(ctx, fromContext.PublicKey, fromContext.SecretKey, fromContext.SecretID, fromContext.Region)
		if err != nil {
			return &types.ApiError{Err: "failed to retrive Access Log from API", Status: http.StatusBadRequest}
		}
		fromContext.FoundedAccessLog = access
	}

	// writing to the client back
	report := aws.GenerateReportStringBySecret(*fromContext.FoundedSecret, fromContext.FoundedAccessLog)
	toSend := types.GetReportResponse{
		Report: report,
	}

	if err := GenericEncoding.WriteJson(rw, http.StatusOK, toSend); err != nil {
		log.Printf("failed to write json to client %v", err)
	}
	return nil
}
