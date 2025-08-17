package httpresp

import (
	"context"
	"encoding/json"
	"net/http"
)

func httpJson(w http.ResponseWriter, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "json marshal error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func HttpError(ctx context.Context, w http.ResponseWriter, err error) {
	httpJson(w, parseError(ctx, err))
}

func HttpSuccess(w http.ResponseWriter, data any) {
	httpJson(w, parseSuccess(data))
}
