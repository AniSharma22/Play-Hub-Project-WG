package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"time"
)

type SlotHandler struct {
	slotService service_interfaces.SlotService
}

func NewSlotHandler(slotService service_interfaces.SlotService) *SlotHandler {
	return &SlotHandler{
		slotService: slotService,
	}
}

func (s *SlotHandler) GetSlotByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameIdStr := vars["id"]

	slotId, err := uuid.Parse(gameIdStr)
	if err != nil {
		logger.Logger.Errorw("Error parsing slot ID", "slotID", gameIdStr, "error", err, "method", r.Method, "time", time.Now())
		errs.NewInternalServerError("Couldn't parse slot id").ToJSON(w)
		return
	}

	slot, err := s.slotService.GetSlotByID(r.Context(), slotId)
	if err != nil {
		logger.Logger.Errorw("Error retrieving slot", "slotID", slotId.String(), "error", err, "method", r.Method, "time", time.Now())
		errs.NewInternalServerError("Couldn't get slot").ToJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"slot":    slot,
	}
	if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
		logger.Logger.Errorw("Error encoding response", "slotID", slotId.String(), "error", err, "method", r.Method, "time", time.Now())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Logger.Infow("Slot sent successfully", "slotID", slotId.String(), "method", r.Method, "time", time.Now())
}

func (s *SlotHandler) GetTodaySlotsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameIdStr := vars["gameID"]

	gameId, err := uuid.Parse(gameIdStr)
	if err != nil {
		logger.Logger.Errorw("Error parsing game ID", "gameID", gameIdStr, "error", err, "method", r.Method, "time", time.Now())
		errs.NewInternalServerError("Couldn't parse game id").ToJSON(w)
		return
	}

	slots, err := s.slotService.GetCurrentDayGameSlots(r.Context(), gameId)
	if err != nil {
		logger.Logger.Errorw("Error retrieving today's slots", "gameID", gameId.String(), "error", err, "method", r.Method, "time", time.Now())
		errs.NewInternalServerError("Couldn't get slots").ToJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"slots":   slots,
	}
	if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
		logger.Logger.Errorw("Error encoding response", "gameID", gameId.String(), "error", err, "method", r.Method, "time", time.Now())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Logger.Infow("All slots for today sent successfully", "gameID", gameId.String(), "method", r.Method, "time", time.Now())
}
