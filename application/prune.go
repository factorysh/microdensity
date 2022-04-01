package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// PruneParam represents http parameters
type PruneParam struct {
	Since string `json:"since"`
	Dry   bool   `json:"dry"`
}

// PruneResponse format the PruneHandler resp body
type PruneResponse struct {
	Reclaimed string `json:"reclaimed"`
	Dry       bool   `json:"dry_run"`
}

// PruneHandler handles prune call
func (a *Application) PruneHandler(w http.ResponseWriter, r *http.Request) {
	l := a.logger.With(
		zap.String("url", r.URL.String()),
	)

	var param PruneParam
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&param)
	if err != nil {
		l.Warn("decode prune param error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	duration, err := time.ParseDuration(param.Since)
	if err != nil {
		l.Warn("unable to parse given duration", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	// can i get the lock ?
	if !a.PruneLock.TryAcquire(1) {
		l.Warn("multiple call to prune")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Prune task already runnig"))
		return
	}

	// do not forget to release at the end
	defer a.PruneLock.Release(1)

	// TODO: handle long prune requests
	reclaimedBytes, err := a.storage.Prune(duration, param.Dry)
	if err != nil {
		l.Warn("error on prune", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	reclaimedMb := float64(reclaimedBytes) / 1024.0 / 1024.0

	resp := PruneResponse{
		Reclaimed: fmt.Sprintf("%fMB", reclaimedMb),
		Dry:       param.Dry,
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(resp)
	if err != nil {
		l.Warn("error when encoding prune response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

}
