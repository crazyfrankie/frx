package httpresp

import (
	"context"
	"errors"

	"github.com/crazyfrankie/frx/errorx"
	"github.com/crazyfrankie/frx/logs"
)

func parseError(ctx context.Context, err error) *data {
	var customErr errorx.StatusError

	if errors.As(err, &customErr) && customErr.Code() != 0 {
		logs.CtxWarnf(ctx, "[ErrorX] error:  %v %v \n", customErr.Code(), err)
		return &data{Code: customErr.Code(), Message: customErr.Msg()}
	}

	logs.CtxErrorf(ctx, "[InternalError]  error: %v \n", err)
	return &data{Code: 500, Message: "internal server error"}
}

func parseSuccess(resp any) *data {
	return &data{Data: resp}
}
