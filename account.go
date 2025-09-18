package account

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
	"internal/pkg/http/client"
	"pkg/log"
)

var (
	instance *Manager
)

func GetInstance() *Manager {
	return instance
}

type Manager struct {
	localTokenCache        AuthCache
	authBasePath           string
	tokenTTL               int64
	getServiceTokenAPIPath string
	getServiceTokenXMLBody string
	accountProfileBasePath string
}

type Config struct {
	AuthServiceHost     string
	AuthClientID        string
	AuthClientSecret    string
	AuthServiceCacheTTL int64
	AccountProfileHost  string
}

func (manager *Manager) Setup(setupConfig Config) error {
	instance = &Manager{
		authBasePath:           setupConfig.AuthServiceHost,
		getServiceTokenAPIPath: fmt.Sprintf("%s%s", setupConfig.AuthServiceHost, "/$SS$/Services/OAuth/Token"),
		getServiceTokenXMLBody: fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s", setupConfig.AuthClientID, setupConfig.AuthClientSecret),
		tokenTTL:               setupConfig.AuthServiceCacheTTL,
		accountProfileBasePath: setupConfig.AccountProfileHost,
	}
	return nil
}

func (manager *Manager) FlushLocalCache() {
	manager.localTokenCache = AuthCache{IsSet: false}
}

func (manager *Manager) GetWalletAddress(ctx context.Context, accountID string) (string, error) {
	serviceToken, err := manager.GetServiceToken(ctx)
	if err != nil {
		return "", err
	}
	httpRes, err := client.NewHTTPRequest(ctx).
		EnableTrace().
		SetHeader("authkey", serviceToken).
		SetPathParams(map[string]string{
			"account-id": accountID,
		}).
		SetQueryParam("fields", "accountProvider").
		SetResult(GetAccountProfileResponse{}).
		Get(manager.accountProfileBasePath + "/SS/Profiles/v3/{account-id}")
	if err != nil {
		log.C(ctx).Errorf("GetWalletAddress Get Err: %v", err)
		return "", err
	}
	if httpRes.StatusCode() != http.StatusOK {
		log.C(ctx).Errorf("GetWalletAddress Not ok: code=%d, resp=%v", httpRes.StatusCode(), string(httpRes.Body()))
		return "", errors.New("get service token not ok")
	}
	getProfileResp := httpRes.Result().(*GetAccountProfileResponse)
	return getProfileResp.WalletAddress, nil
}

func (manager *Manager) GetServiceToken(ctx context.Context) (string, error) {
	tokenFromCache, err := manager.getTokenFromCache()
	if err == nil {
		return tokenFromCache, nil
	}
	newToken, err := manager.getNewToken(ctx)
	if err != nil {
		return "", err
	}
	manager.localTokenCache = AuthCache{
		IsSet:          true,
		Token:          newToken,
		StartTimeInSec: time.Now().Unix(),
	}
	return newToken, nil
}

func (manager *Manager) getTokenFromCache() (string, error) {
	if manager.localTokenCache.IsSet && manager.localTokenCache.StartTimeInSec+manager.tokenTTL > time.Now().Unix() {
		return manager.localTokenCache.Token, nil
	} else {
		return "", errors.New("no token or token expired")
	}
}

func (manager *Manager) getNewToken(ctx context.Context) (string, error) {
	req := client.NewHTTPRequest(ctx).
		SetBody(manager.getServiceTokenXMLBody).
		SetHeader("Content-Type", "application/x-www-form-urlencoded")

	resp, err := req.Post(manager.getServiceTokenAPIPath)
	if err != nil {
		log.C(ctx).Errorf("getNewToken Post Err: %v", err)
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		log.C(ctx).Errorf("getNewToken Not ok: code=%d, resp=%v", resp.StatusCode(), string(resp.Body()))
		return "", errors.New("get service token not ok")
	}
	getTokenResp := GetAuthTokenResponse{}
	if err := jsoniter.Unmarshal(resp.Body(), &getTokenResp); err != nil {
		log.C(ctx).Errorf("getNewToken parse failed: resp=%v, err:%v", string(resp.Body()), err)
		return "", err
	}
	return getTokenResp.AccessToken, nil
}

type AuthCache struct {
	IsSet          bool
	Token          string
	StartTimeInSec int64
}

type GetAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type GetAccountProfileResponse struct {
	WalletAddress string `json:"ethereumPublicAddr"`
}
