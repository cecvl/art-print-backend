package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cecvl/art-print-backend/internal/services/catalog"
)

type PrintOptionsHandler struct {
	catalog *catalog.CatalogService
}

func NewPrintOptionsHandler() *PrintOptionsHandler {
	return &PrintOptionsHandler{
		catalog: catalog.NewCatalogService(),
	}
}

func (h *PrintOptionsHandler) GetPrintOptions(w http.ResponseWriter, r *http.Request) {
	opts := h.catalog.GetPrintOptions()
	json.NewEncoder(w).Encode(opts)
}
