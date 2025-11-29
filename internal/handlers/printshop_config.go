package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/repositories"
)

// PrintShopConfigHandler handles configuration management (frames, sizes, materials)
type PrintShopConfigHandler struct {
	repo *repositories.PrintShopRepository
}

// NewPrintShopConfigHandler creates a new config handler
func NewPrintShopConfigHandler() *PrintShopConfigHandler {
	return &PrintShopConfigHandler{
		repo: repositories.NewPrintShopRepository(firebase.FirestoreClient),
	}
}

// ==================== Frame Management ====================

// GetFrames retrieves all frames for the authenticated shop
func (h *PrintShopConfigHandler) GetFrames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	frames, err := h.repo.GetFramesByShopID(ctx, shop.ID)
	if err != nil {
		log.Printf("❌ Failed to get frames: %v", err)
		http.Error(w, "Failed to get frames", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(frames)
}

// CreateFrame creates a new frame configuration
func (h *PrintShopConfigHandler) CreateFrame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	var frame models.Frame
	if err := json.NewDecoder(r.Body).Decode(&frame); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	frame.ShopID = shop.ID
	frame.IsActive = true

	if err := h.repo.CreateFrame(ctx, &frame); err != nil {
		log.Printf("❌ Failed to create frame: %v", err)
		http.Error(w, "Failed to create frame", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(frame)
}

// UpdateFrame updates a frame configuration
func (h *PrintShopConfigHandler) UpdateFrame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	pathParts := strings.Split(r.URL.Path, "/")
	frameID := pathParts[len(pathParts)-1]

	// Get frame to verify ownership
	frames, err := h.repo.GetFramesByShopID(ctx, "")
	if err == nil {
		// Find frame and verify shop ownership
		shop, _ := h.repo.GetShopByOwnerID(ctx, ownerID)
		for _, f := range frames {
			if f.ID == frameID && f.ShopID == shop.ID {
				var updates map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}
				delete(updates, "shopId")
				delete(updates, "id")

				if err := h.repo.UpdateFrame(ctx, frameID, updates); err != nil {
					http.Error(w, "Failed to update frame", http.StatusInternalServerError)
					return
				}

				updatedFrame, _ := h.repo.GetFramesByShopID(ctx, shop.ID)
				for _, f := range updatedFrame {
					if f.ID == frameID {
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(f)
						return
					}
				}
			}
		}
	}

	http.Error(w, "Frame not found or access denied", http.StatusNotFound)
}

// DeleteFrame deletes a frame configuration
func (h *PrintShopConfigHandler) DeleteFrame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	pathParts := strings.Split(r.URL.Path, "/")
	frameID := pathParts[len(pathParts)-1]

	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	frames, _ := h.repo.GetFramesByShopID(ctx, shop.ID)
	for _, f := range frames {
		if f.ID == frameID {
			if err := h.repo.DeleteFrame(ctx, frameID); err != nil {
				http.Error(w, "Failed to delete frame", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "Frame not found", http.StatusNotFound)
}

// ==================== Size Management ====================

// GetSizes retrieves all sizes for the authenticated shop
func (h *PrintShopConfigHandler) GetSizes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	sizes, err := h.repo.GetSizesByShopID(ctx, shop.ID)
	if err != nil {
		log.Printf("❌ Failed to get sizes: %v", err)
		http.Error(w, "Failed to get sizes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sizes)
}

// CreateSize creates a new size configuration
func (h *PrintShopConfigHandler) CreateSize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	var size models.PrintSize
	if err := json.NewDecoder(r.Body).Decode(&size); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	size.ShopID = shop.ID
	size.IsActive = true

	if err := h.repo.CreateSize(ctx, &size); err != nil {
		log.Printf("❌ Failed to create size: %v", err)
		http.Error(w, "Failed to create size", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(size)
}

// UpdateSize updates a size configuration
func (h *PrintShopConfigHandler) UpdateSize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	pathParts := strings.Split(r.URL.Path, "/")
	sizeID := pathParts[len(pathParts)-1]

	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	sizes, _ := h.repo.GetSizesByShopID(ctx, shop.ID)
	for _, s := range sizes {
		if s.ID == sizeID {
			var updates map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			delete(updates, "shopId")
			delete(updates, "id")

			if err := h.repo.UpdateSize(ctx, sizeID, updates); err != nil {
				http.Error(w, "Failed to update size", http.StatusInternalServerError)
				return
			}

			updatedSizes, _ := h.repo.GetSizesByShopID(ctx, shop.ID)
			for _, s := range updatedSizes {
				if s.ID == sizeID {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(s)
					return
				}
			}
		}
	}

	http.Error(w, "Size not found", http.StatusNotFound)
}

// DeleteSize deletes a size configuration
func (h *PrintShopConfigHandler) DeleteSize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	pathParts := strings.Split(r.URL.Path, "/")
	sizeID := pathParts[len(pathParts)-1]

	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	sizes, _ := h.repo.GetSizesByShopID(ctx, shop.ID)
	for _, s := range sizes {
		if s.ID == sizeID {
			if err := h.repo.DeleteSize(ctx, sizeID); err != nil {
				http.Error(w, "Failed to delete size", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "Size not found", http.StatusNotFound)
}

// ==================== Material Management ====================

// GetMaterials retrieves all materials for the authenticated shop
func (h *PrintShopConfigHandler) GetMaterials(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	materials, err := h.repo.GetMaterialsByShopID(ctx, shop.ID)
	if err != nil {
		log.Printf("❌ Failed to get materials: %v", err)
		http.Error(w, "Failed to get materials", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(materials)
}

// CreateMaterial creates a new material configuration
func (h *PrintShopConfigHandler) CreateMaterial(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	var material models.Material
	if err := json.NewDecoder(r.Body).Decode(&material); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	material.ShopID = shop.ID
	material.IsActive = true

	if err := h.repo.CreateMaterial(ctx, &material); err != nil {
		log.Printf("❌ Failed to create material: %v", err)
		http.Error(w, "Failed to create material", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(material)
}

// UpdateMaterial updates a material configuration
func (h *PrintShopConfigHandler) UpdateMaterial(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	pathParts := strings.Split(r.URL.Path, "/")
	materialID := pathParts[len(pathParts)-1]

	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	materials, _ := h.repo.GetMaterialsByShopID(ctx, shop.ID)
	for _, m := range materials {
		if m.ID == materialID {
			var updates map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			delete(updates, "shopId")
			delete(updates, "id")

			if err := h.repo.UpdateMaterial(ctx, materialID, updates); err != nil {
				http.Error(w, "Failed to update material", http.StatusInternalServerError)
				return
			}

			updatedMaterials, _ := h.repo.GetMaterialsByShopID(ctx, shop.ID)
			for _, m := range updatedMaterials {
				if m.ID == materialID {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(m)
					return
				}
			}
		}
	}

	http.Error(w, "Material not found", http.StatusNotFound)
}

// DeleteMaterial deletes a material configuration
func (h *PrintShopConfigHandler) DeleteMaterial(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	pathParts := strings.Split(r.URL.Path, "/")
	materialID := pathParts[len(pathParts)-1]

	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	materials, _ := h.repo.GetMaterialsByShopID(ctx, shop.ID)
	for _, m := range materials {
		if m.ID == materialID {
			if err := h.repo.DeleteMaterial(ctx, materialID); err != nil {
				http.Error(w, "Failed to delete material", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "Material not found", http.StatusNotFound)
}

