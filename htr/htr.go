package htr

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/frx/errorx"
	"github.com/crazyfrankie/frx/htr/errno"
	"github.com/crazyfrankie/frx/htr/httpresp"
)

type Option[A, B any] struct {
	// BindAfter is called after the req is bind from ctx.
	BindAfter func(*A) error
	// RespAfter is called after the resp is return from rpc.
	RespAfter func(*B) error
}

func Call[A, B, C any](c *gin.Context, rpc func(client C, ctx context.Context, req *A, opts ...grpc.DialOption) (*B, error), client C, opts ...*Option[A, B]) {
	req, err := ginParseRequest[A](c)
	if err != nil {
		httpresp.GinError(c, err)
		return
	}

	for _, opt := range opts {
		if opt.BindAfter == nil {
			continue
		}
		if err := opt.BindAfter(req); err != nil {
			httpresp.GinError(c, err) // args option error
			return
		}
	}

	resp, err := rpc(client, c.Request.Context(), req)
	if err != nil {
		httpresp.GinError(c, err)
		return
	}

	for _, opt := range opts {
		if opt.RespAfter == nil {
			continue
		}
		if err := opt.RespAfter(resp); err != nil {
			httpresp.GinError(c, err) // resp option error
			return
		}
	}

	httpresp.GinSuccess(c, resp) // rpc call success
}

func CallHttp[A, B, C any](w http.ResponseWriter, r *http.Request, rpc func(client C, ctx context.Context, req *A, opts ...grpc.DialOption) (*B, error), client C, opts ...*Option[A, B]) {
	ctx := r.Context()
	req, err := httpParseRequest[A](r)
	if err != nil {
		httpresp.HttpError(ctx, w, err)
		return
	}

	for _, opt := range opts {
		if opt.BindAfter != nil {
			if err := opt.BindAfter(req); err != nil {
				httpresp.HttpError(ctx, w, err)
				return
			}
		}
	}

	resp, err := rpc(client, ctx, req)
	if err != nil {
		httpresp.HttpError(ctx, w, err)
		return
	}

	for _, opt := range opts {
		if opt.RespAfter != nil {
			if err := opt.RespAfter(resp); err != nil {
				httpresp.HttpError(ctx, w, err)
				return
			}
		}
	}

	httpresp.HttpSuccess(w, resp)
}

func ginParseRequest[T any](c *gin.Context) (*T, error) {
	var req T
	if err := c.ShouldBind(&req); err != nil {
		return nil, errorx.New(errno.ErrInvalidParamsCode)
	}

	return &req, nil
}

func httpParseRequest[T any](r *http.Request) (*T, error) {
	var req T
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errorx.New(errno.ErrInvalidParamsCode)
	}

	err := validator.New().Struct(&req)
	if err != nil {
		return nil, errorx.New(errno.ErrInvalidParamsCode)
	}

	return &req, nil
}
