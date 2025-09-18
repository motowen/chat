package point

import (
	"context"
	"errors"
	"net/http"

	"internal/pkg/errors/errorv2"
	"internal/pkg/htcaccount"
	"internal/pkg/http/client"
	"pkg/log"
)

type PointOperationType string
type PointType string
type PointDescription string

const (
	PointOperationTypeAdd PointOperationType = "add"
	PointOperationTypeSub PointOperationType = "subtract"

	PointTypeAdd PointType = "task"
	PointTypeSub PointType = "consume"

	PointDescriptionOutfitCreatorToolCreate PointDescription = "Outfit Creator Tool Create"
	PointDescriptionOutfitCreatorToolImage  PointDescription = "Outfit Creator Tool Image"
)

var (
	instance *Manager
)

func GetInstance() *Manager {
	return instance
}

type Manager struct {
	PointServiceHost                string
	OutfitCreatorCreateOutfitPoints int
	AuthClientID                    string
}

type Config struct {
	PointServiceHost                string
	OutfitCreatorCreateOutfitPoints int
	AuthClientID                    string
}

func (manager *Manager) Setup(setupConfig Config) error {
	instance = &Manager{
		PointServiceHost:                setupConfig.PointServiceHost,
		OutfitCreatorCreateOutfitPoints: setupConfig.OutfitCreatorCreateOutfitPoints,
		AuthClientID:                    setupConfig.AuthClientID,
	}
	return nil
}

func GetPointsBalance(ctx context.Context, accountID string) (res GetPointsBalanceResp, err error) {
	serviceToken, err := htcaccount.GetInstance().GetServiceToken(ctx)
	if err != nil {
		return res, err
	}

	restyReq := client.NewHTTPRequest(ctx).
		SetHeader("authkey", serviceToken).
		SetHeader("Content-Type", "application/json").
		SetPathParams(map[string]string{
			"userId": accountID,
		}).
		SetResult(&res)

	// API DOC: //https://htcsense.jira.com/wiki/spaces/NEOSTORE/pages/3479175176/Point+Service+API+Doc#%5BGET%5D-%2Fpriv%2Fpoint-service%2Fv1%2Fpoints
	restyResp, err := restyReq.Get(GetInstance().PointServiceHost + "/priv/point-service/v1/points?userId={userId}")
	if err != nil {
		return
	}
	if restyResp.StatusCode() != http.StatusOK {
		err = errors.New(errorv2.PointGetPointNotOK)
		return
	}
	return
}

func OperatePoints(ctx context.Context, req OperatePointsReq) (res OperatePointsResp, err error) {
	serviceToken, err := htcaccount.GetInstance().GetServiceToken(ctx)
	if err != nil {
		return res, err
	}

	restyReq := client.NewHTTPRequest(ctx).
		SetHeader("authkey", serviceToken).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&res)

	// API DOC: //https://htcsense.jira.com/wiki/spaces/NEOSTORE/pages/3479175176/Point+Service+API+Doc#%5BPOST%5D-%2Fpriv%2Fpoint-service%2Fv1%2Fpoints
	restyResp, err := restyReq.Post(GetInstance().PointServiceHost + "/priv/point-service/v1/points")
	if err != nil {
		return
	}
	if restyResp.StatusCode() != http.StatusOK {
		if res.CommonRes.Error == ErrInsufficientPoint {
			err = errors.New(errorv2.PointOperatePointInsifficientPointBalance)
			return
		}
		switch restyResp.StatusCode() {
		case http.StatusBadRequest:
			err = errors.New(errorv2.PointOperatePointInvalidRequest)
		default:
			err = errors.New(errorv2.PointOperatePointNotOK)
		}
		return
	}
	log.C(ctx).Infof("OperatePoints success, res: %+v", res)
	return
}
