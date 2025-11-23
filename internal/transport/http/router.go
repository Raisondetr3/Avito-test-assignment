package http

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Raisondetr3/Avito-test-assignment/internal/transport/http/handlers"
)

func NewRouter(teamHandler *handlers.TeamHandler, userHandler *handlers.UserHandler, prHandler *handlers.PRHandler, statsHandler *handlers.StatsHandler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/team/add", teamHandler.CreateTeam).Methods(http.MethodPost)
	r.HandleFunc("/team/get", teamHandler.GetTeam).Methods(http.MethodGet)
	r.HandleFunc("/team/deactivateUsers", teamHandler.BulkDeactivateUsers).Methods(http.MethodPost)

	r.HandleFunc("/users/setIsActive", userHandler.SetActive).Methods(http.MethodPost)
	r.HandleFunc("/users/getReview", userHandler.GetReview).Methods(http.MethodGet)

	r.HandleFunc("/pullRequest/create", prHandler.CreatePR).Methods(http.MethodPost)
	r.HandleFunc("/pullRequest/merge", prHandler.MergePR).Methods(http.MethodPost)
	r.HandleFunc("/pullRequest/reassign", prHandler.ReassignPR).Methods(http.MethodPost)

	r.HandleFunc("/stats", statsHandler.GetStatistics).Methods(http.MethodGet)

	r.HandleFunc("/health", healthCheck).Methods(http.MethodGet)

	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK")) //nolint:errcheck
}
