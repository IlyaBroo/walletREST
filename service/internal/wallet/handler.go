package wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"service/internal/config"
	"service/internal/logger"
	"sync"

	"github.com/go-chi/chi/v5"
)

const (
	DEPOSIT  string = "DEPOSIT"
	WITHDRAW string = "WITHDRAW"
)

type WalletOperationRequest struct {
	WalletID      string `json:"walletId"`
	OperationType string `json:"operationType"`
	Amount        int64  `json:"amount"`
}

type Handler struct {
	repo RepositoryInterface
	mu   sync.Mutex
	lg   logger.Logger
	ctx  context.Context
}

func NewHandler(lg logger.Logger, ctx context.Context, cfg *config.ConfigAdr) *Handler {
	return &Handler{
		repo: NewRepository(lg, ctx, cfg),
		lg:   lg,
		ctx:  ctx,
	}
}

func (h *Handler) Close() {
	h.repo.Close()
}

func (h *Handler) HandleWalletOperation(w http.ResponseWriter, r *http.Request) {
	var request WalletOperationRequest
	h.ctx = r.Context()
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.lg.ErrorCtx(h.ctx, "error decode request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.OperationType == DEPOSIT {
		if err := h.repo.Deposit(request.WalletID, request.Amount, h.ctx); err != nil {
			if err == errWalletid {
				h.lg.ErrorCtx(h.ctx, "insufficient funds or walletid not found")
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			h.lg.ErrorCtx(h.ctx, fmt.Sprintf("deposit err = %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if request.OperationType == WITHDRAW {
		if err := h.repo.Withdraw(request.WalletID, request.Amount, h.ctx); err != nil {
			if err == errWithdraw {
				h.lg.ErrorCtx(h.ctx, "insufficient funds or walletid not found")
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			h.lg.ErrorCtx(h.ctx, fmt.Sprintf("withdraw err = %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		h.lg.ErrorCtx(h.ctx, "invalid operation type")
		http.Error(w, "invalid operation type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	h.lg.InfoCtx(h.ctx, fmt.Sprintf("wallet id = %s, operation = %s , amount = %d is success", request.WalletID, request.OperationType, request.Amount))
}

func (h *Handler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	h.ctx = r.Context()
	walletID := chi.URLParam(r, "id")
	h.lg.DebugCtx(h.ctx, fmt.Sprintf("walletId=%v", walletID))

	balance, err := h.repo.GetBalance(walletID, h.ctx)
	if err == errWalletid {
		h.lg.ErrorCtx(h.ctx, "walletid not found")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		h.lg.ErrorCtx(h.ctx, "error getting balance")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"walletId": walletID,
		"balance":  balance,
	})
	w.WriteHeader(http.StatusOK)
	h.lg.InfoCtx(h.ctx, fmt.Sprintf("wallet id = %s, balance = %d is success", walletID, balance))
}
