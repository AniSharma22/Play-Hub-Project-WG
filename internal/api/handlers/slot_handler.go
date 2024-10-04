package handlers

import (
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"project2/pkg/utils"
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
		errs.ValidationError("Couldn't parse slot id").ToJson2(w)
		return
	}

	slot, err := s.slotService.GetSlotByID(r.Context(), slotId)
	if err != nil {
		logger.Logger.Errorw("Error retrieving slot", "slotID", slotId.String(), "error", err, "method", r.Method, "time", time.Now())
		errs.DBError("Couldn't get slot").ToJson2(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"slot":    slot,
	}
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
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
		errs.ValidationError("Couldn't parse game id").ToJson2(w)
		return
	}

	slots, err := s.slotService.GetCurrentDayGameSlots(r.Context(), gameId)
	if err != nil {
		if errors.Is(err, errs.ErrGameNotFound) {
			logger.Logger.Errorw("No slots found for the given game id", "gameID", gameId.String(), "error", err, "method", r.Method, "time", time.Now())
			errs.InvalidRequestError("No slots found for the game").ToJson2(w)
			return
		}
		logger.Logger.Errorw("Error retrieving today's slots", "gameID", gameId.String(), "error", err, "method", r.Method, "time", time.Now())
		errs.DBError("Couldn't get slots").ToJson2(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"slots":   slots,
	}
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("All slots for today sent successfully", "gameID", gameId.String(), "method", r.Method, "time", time.Now())
}
