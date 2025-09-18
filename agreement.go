package external

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"internal/pkg/config"
	"internal/pkg/http/client"
	"internal/pkg/model"
)

const upsertApiPath = "/priv/recurringagreement/v1/agreement/upsert"
const removeApiPath = "/priv/recurringagreement/v1/agreement/remove"

func UpsertAgreement(req model.UpsertAgreementReq, requestLogger *logrus.Entry) (resp model.UpsertAgreementResp, err error) {
	authToken, getAuthTokenErr := GetAuthToken(requestLogger)
	if getAuthTokenErr != nil {
		msg := fmt.Sprintf("[UpsertAgreement] failed to get auth token, err: %s", getAuthTokenErr.Error())
		requestLogger.Error(msg)
		err = fmt.Errorf(msg)
		return
	}

	uri := fmt.Sprintf("%s%s", config.EnvVariable.AgreementServiceHost, upsertApiPath)

	requestLogger = requestLogger.WithFields(logrus.Fields{
		"uri": uri,
		"req": fmt.Sprintf("%+v", req),
	})
	requestLogger.Info("[UpsertAgreement] sending request to upsert agreement api")

	httpResp, err := client.NewRequest().
		SetHeader("Content-Type", "application/json").
		SetHeader("authKey", authToken).
		SetBody(req).
		Post(uri)

	if err != nil {
		msg := fmt.Sprintf("[UpsertAgreement] failed to send api request, err: %s", err.Error())
		requestLogger.Error(msg)
		err = fmt.Errorf(msg)
		return
	}

	if httpResp.StatusCode() != http.StatusCreated {
		msg := fmt.Sprintf("[UpsertAgreement] call %s fail, got status code: %d, resp body: %s", uri, httpResp.StatusCode(), httpResp.Body())
		requestLogger.Error(msg)
		err = fmt.Errorf(msg)
		return
	}

	if err = json.Unmarshal(httpResp.Body(), &resp); err != nil {
		msg := fmt.Sprintf("[UpsertAgreement] failed to parse response body, err: %s, resp body: %s", err.Error(), httpResp.Body())
		requestLogger.Error(msg)
		err = fmt.Errorf(msg)
		return
	}

	return
}

func RemoveAgreement(req model.RemoveAgreementReq, requestLogger *logrus.Entry) (resp model.RemoveAgreementResp, err error) {
	authToken, getAuthTokenErr := GetAuthToken(requestLogger)
	if getAuthTokenErr != nil {
		msg := fmt.Sprintf("[RemoveAgreement] failed to get auth token, err: %s", getAuthTokenErr.Error())
		requestLogger.Error(msg)
		err = fmt.Errorf(msg)
		return
	}

	uri := fmt.Sprintf("%s%s", config.EnvVariable.AgreementServiceHost, removeApiPath)

	requestLogger = requestLogger.WithFields(logrus.Fields{
		"uri": uri,
		"req": fmt.Sprintf("%+v", req),
	})
	requestLogger.Info("[RemoveAgreement] sending request to remove agreement api")

	httpResp, err := client.NewRequest().
		SetHeader("Content-Type", "application/json").
		SetHeader("authKey", authToken).
		SetBody(req).
		Post(uri)

	if err != nil {
		msg := fmt.Sprintf("[RemoveAgreement] failed to send api request, err: %s", err.Error())
		requestLogger.Error(msg)
		err = fmt.Errorf(msg)
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		msg := fmt.Sprintf("[RemoveAgreement] call %s fail, got status code: %d, resp body: %s", uri, httpResp.StatusCode(), httpResp.Body())
		requestLogger.Error(msg)
		err = fmt.Errorf(msg)
		return
	}

	if err = json.Unmarshal(httpResp.Body(), &resp); err != nil {
		msg := fmt.Sprintf("[RemoveAgreement] failed to parse response body, err: %s, resp body: %s", err.Error(), httpResp.Body())
		requestLogger.Error(msg)
		err = fmt.Errorf(msg)
		return
	}

	return
}
