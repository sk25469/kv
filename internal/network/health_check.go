package network

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sk25469/kv/internal/comm"
)

type IHealthCheck interface {
	HealthCheck() error
}

type HealthCheckServiceParams struct {
	Port                 int
	NetworkService       INetwork
	CommunicationService comm.ICommunication
}

type HealthCheckService struct {
	port                 int
	networkService       *NetworkService
	communicationService *comm.CommunicationService
}

func NewHealthCheckService(params HealthCheckServiceParams) *HealthCheckService {
	return &HealthCheckService{
		port:                 params.Port,
		networkService:       params.NetworkService.(*NetworkService),
		communicationService: params.CommunicationService.(*comm.CommunicationService),
	}
}

func (h *HealthCheckService) StartHealthCheck() {
	http.HandleFunc("/health", h.healthCheck)

	go func() {
		log.Infof("Starting health check server on port %d", h.port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", h.port), nil); err != nil {
			log.Fatalf("Health check server failed: %v", err)
		}
	}()
}

func (h *HealthCheckService) healthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.checkServicesHealth(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func (h *HealthCheckService) checkServicesHealth() error {
	if !h.networkService.IsListenerActive() {
		return fmt.Errorf("listener is not active")
	}

	err := h.communicationService.GetEtcdClientHealth()
	if err != nil {
		return err
	}
	return nil
}

func (h *HealthCheckService) StartPeriodicHealthChecks(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := h.checkServicesHealth(); err != nil {
				log.Errorf("Health check failed: %v", err)
			} else {
				log.Info("Health check passed")
			}
		}
	}()
}
