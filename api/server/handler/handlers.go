package handler

import (
	"fmt"
	"golang-secret-manager/api/aws"
	"golang-secret-manager/types"
	GenericEncoding "golang-secret-manager/utils/genericEncoding"
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
				GenericEncoding.WriteJson(rw, http.StatusInternalServerError, types.ApiError{Err: "internal error", Status: http.StatusInternalServerError})
				return
			}
			// apiErr is type apiError
			GenericEncoding.WriteJson(rw, apiErr.Status, apiErr.Err)
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
	// TODO
	return nil
}
