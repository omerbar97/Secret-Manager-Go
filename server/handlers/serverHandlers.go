package handlers

import (
	"encoding/json"
	"fmt"
	"golang-secret-manager/api/aws"
	"golang-secret-manager/utils/types"
	"net/http"
)

func GetAllSecretsHandlers(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()
	contextKey := types.GetContextInforamtionKey()
	fromContext, ok := ctx.Value(contextKey).(*types.FromGetAllSecretsMiddlewareToHandler)
	if !ok {
		fmt.Println("HANDLER: failed to convert the context value to the handler")
		json.NewEncoder(rw).Encode(map[string]string{"error": "Error Converting the context to the handler"})
		// setting response code
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !fromContext.FoundedArnList {
		// there was not ArnList for that user in the cache
		val, err := aws.RetriveAllSecretsWithAccessLog(ctx, fromContext.PublicKey, fromContext.SecretKey, fromContext.Region)
		if err != nil {
			// failed to retrive all of them, return bad request
			fmt.Println("HANDLER: failed to retrive all the secrets and access log")
			rw.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(map[string]string{"error": "Error while trying to retrive all Secrets + Access Logs"})
			return
		}

		// sending back to the client the anwser
		toSend := types.GetAllSecretsResponse{
			Secrets:   val.Secrets,
			AccessLog: val.AccessLog,
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(toSend)
		return
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

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(toSend)
}

func GetReportsHandler(rw http.ResponseWriter, r *http.Request) {
	// TODO
}
