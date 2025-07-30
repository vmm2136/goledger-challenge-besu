package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vmm2136/besu_challenge/go-app/internal/service"
)

// Handler lida com as requisições HTTP para o contrato
type Handler struct {
	contractService service.ContractService
}

// NewHandler cria um novo Handler
func NewHandler(svc service.ContractService) *Handler {
	return &Handler{
		contractService: svc,
	}
}

// GetValueHandler lida com a requisição GET /value
func (h *Handler) GetValueHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	value, transactorAddress, err := h.contractService.GetCurrentValue(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao obter valor do contrato: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"current_value":      value.String(),
		"transactor_address": transactorAddress.Hex(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SetValueRequest representa o corpo da requisição POST /value.
type SetValueRequest struct {
	Value int64 `json:"value"`
}

// SetValueHandler lida com a requisição POST /value
func (h *Handler) SetValueHandler(w http.ResponseWriter, r *http.Request) {
	var req SetValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Payload inválido: %v", err), http.StatusBadRequest)
		return
	}

	if req.Value < 0 {
		http.Error(w, "O valor não pode ser negativo", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	txHash, err := h.contractService.SetNewValue(ctx, req.Value)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao definir valor no contrato: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message":   "Transação enviada com sucesso",
		"tx_hash":   txHash.Hex(),
		"new_value": strconv.FormatInt(req.Value, 10),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

// SyncValueHandler lida com a requisição POST /sync
func (h *Handler) SyncValueHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	networkValue, dbValue, err := h.contractService.SyncContractValue(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao sincronizar valor do contrato: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message":        "Sincronização concluída",
		"network_value":  networkValue.String(),
		"database_value": dbValue.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CheckValueHandler lida com a requisição GET /check
func (h *Handler) CheckValueHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	areEqual, networkValue, dbValue, err := h.contractService.CheckContractValue(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao verificar valor do contrato: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"match":          areEqual,
		"network_value":  networkValue.String(),
		"database_value": dbValue.String(),
		"message":        "Valores comparados com sucesso.",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
