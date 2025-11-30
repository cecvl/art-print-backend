package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

// PrintShopRepository handles all Firestore operations for print shops
type PrintShopRepository struct {
	client *firestore.Client
}

// NewPrintShopRepository creates a new repository instance
func NewPrintShopRepository(client *firestore.Client) *PrintShopRepository {
	return &PrintShopRepository{
		client: client,
	}
}

// ==================== Shop Operations ====================

// GetShopByOwnerID retrieves a print shop by owner's user ID
func (r *PrintShopRepository) GetShopByOwnerID(ctx context.Context, ownerID string) (*models.PrintShopProfile, error) {
	iter := r.client.Collection("printshops").
		Where("ownerId", "==", ownerID).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, errors.New("print shop not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get shop: %w", err)
	}

	var shop models.PrintShopProfile
	if err := doc.DataTo(&shop); err != nil {
		return nil, fmt.Errorf("failed to parse shop data: %w", err)
	}
	shop.ID = doc.Ref.ID

	return &shop, nil
}

// GetShopByID retrieves a print shop by its ID
func (r *PrintShopRepository) GetShopByID(ctx context.Context, shopID string) (*models.PrintShopProfile, error) {
	doc, err := r.client.Collection("printshops").Doc(shopID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get shop: %w", err)
	}

	if !doc.Exists() {
		return nil, errors.New("print shop not found")
	}

	var shop models.PrintShopProfile
	if err := doc.DataTo(&shop); err != nil {
		return nil, fmt.Errorf("failed to parse shop data: %w", err)
	}
	shop.ID = doc.Ref.ID

	return &shop, nil
}

// CreateShop creates a new print shop profile
func (r *PrintShopRepository) CreateShop(ctx context.Context, shop *models.PrintShopProfile) error {
	if shop.ID == "" {
		shop.ID = uuid.New().String()
	}
	if shop.CreatedAt.IsZero() {
		shop.CreatedAt = time.Now()
	}
	if shop.UpdatedAt.IsZero() {
		shop.UpdatedAt = time.Now()
	}

	_, err := r.client.Collection("printshops").Doc(shop.ID).Set(ctx, shop)
	if err != nil {
		return fmt.Errorf("failed to create shop: %w", err)
	}

	return nil
}

// UpdateShop updates a print shop profile
func (r *PrintShopRepository) UpdateShop(ctx context.Context, shopID string, updates map[string]interface{}) error {
	updates["updatedAt"] = time.Now()

	// Convert map to []firestore.Update
	var firestoreUpdates []firestore.Update
	for path, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{
			Path:  path,
			Value: value,
		})
	}

	_, err := r.client.Collection("printshops").Doc(shopID).Update(ctx, firestoreUpdates)
	if err != nil {
		return fmt.Errorf("failed to update shop: %w", err)
	}

	return nil
}

// GetActiveShops retrieves all active print shops
func (r *PrintShopRepository) GetActiveShops(ctx context.Context) ([]*models.PrintShopProfile, error) {
	iter := r.client.Collection("printshops").
		Where("isActive", "==", true).
		Documents(ctx)

	var shops []*models.PrintShopProfile
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate shops: %w", err)
		}

		var shop models.PrintShopProfile
		if err := doc.DataTo(&shop); err != nil {
			continue // Skip invalid documents
		}
		shop.ID = doc.Ref.ID
		shops = append(shops, &shop)
	}

	return shops, nil
}

// ==================== Service Operations ====================

// GetServicesByShopID retrieves all services for a print shop
func (r *PrintShopRepository) GetServicesByShopID(ctx context.Context, shopID string) ([]*models.PrintService, error) {
	iter := r.client.Collection("services").
		Where("shopId", "==", shopID).
		Documents(ctx)

	var services []*models.PrintService
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate services: %w", err)
		}

		var service models.PrintService
		if err := doc.DataTo(&service); err != nil {
			continue
		}
		service.ID = doc.Ref.ID
		services = append(services, &service)
	}

	return services, nil
}

// GetServiceByID retrieves a service by its ID
func (r *PrintShopRepository) GetServiceByID(ctx context.Context, serviceID string) (*models.PrintService, error) {
	doc, err := r.client.Collection("services").Doc(serviceID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	if !doc.Exists() {
		return nil, errors.New("service not found")
	}

	var service models.PrintService
	if err := doc.DataTo(&service); err != nil {
		return nil, fmt.Errorf("failed to parse service data: %w", err)
	}
	service.ID = doc.Ref.ID

	return &service, nil
}

// CreateService creates a new print service
func (r *PrintShopRepository) CreateService(ctx context.Context, service *models.PrintService) error {
	if service.ID == "" {
		service.ID = uuid.New().String()
	}
	if service.CreatedAt.IsZero() {
		service.CreatedAt = time.Now()
	}
	if service.UpdatedAt.IsZero() {
		service.UpdatedAt = time.Now()
	}

	// Initialize empty PriceMatrix if not set
	if service.PriceMatrix.SizeModifiers == nil {
		service.PriceMatrix.SizeModifiers = make(map[string]float64)
	}
	if service.PriceMatrix.MaterialMarkups == nil {
		service.PriceMatrix.MaterialMarkups = make(map[string]float64)
	}
	if service.PriceMatrix.MediumMarkups == nil {
		service.PriceMatrix.MediumMarkups = make(map[string]float64)
	}
	if service.PriceMatrix.FramePrices == nil {
		service.PriceMatrix.FramePrices = make(map[string]float64)
	}

	_, err := r.client.Collection("services").Doc(service.ID).Set(ctx, service)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Update shop's services list
	shopRef := r.client.Collection("printshops").Doc(service.ShopID)
	shopDoc, err := shopRef.Get(ctx)
	if err == nil && shopDoc.Exists() {
		var shop models.PrintShopProfile
		shopDoc.DataTo(&shop)
		shop.Services = append(shop.Services, service.ID)
		shopRef.Update(ctx, []firestore.Update{
			{Path: "services", Value: shop.Services},
			{Path: "updatedAt", Value: time.Now()},
		})
	}

	return nil
}

// UpdateService updates a print service
func (r *PrintShopRepository) UpdateService(ctx context.Context, serviceID string, updates map[string]interface{}) error {
	updates["updatedAt"] = time.Now()

	// Convert map to []firestore.Update
	var firestoreUpdates []firestore.Update
	for path, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{
			Path:  path,
			Value: value,
		})
	}

	_, err := r.client.Collection("services").Doc(serviceID).Update(ctx, firestoreUpdates)
	if err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	return nil
}

// DeleteService deletes a print service
func (r *PrintShopRepository) DeleteService(ctx context.Context, serviceID string) error {
	// Get service to find shop ID
	service, err := r.GetServiceByID(ctx, serviceID)
	if err != nil {
		return err
	}

	// Delete the service
	_, err = r.client.Collection("services").Doc(serviceID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	// Remove from shop's services list
	shopRef := r.client.Collection("printshops").Doc(service.ShopID)
	shopDoc, err := shopRef.Get(ctx)
	if err == nil && shopDoc.Exists() {
		var shop models.PrintShopProfile
		shopDoc.DataTo(&shop)

		// Remove service ID from list
		var newServices []string
		for _, sID := range shop.Services {
			if sID != serviceID {
				newServices = append(newServices, sID)
			}
		}

		shopRef.Update(ctx, []firestore.Update{
			{Path: "services", Value: newServices},
			{Path: "updatedAt", Value: time.Now()},
		})
	}

	return nil
}

// ==================== Frame Operations ====================

// GetFramesByShopID retrieves all frames for a print shop
func (r *PrintShopRepository) GetFramesByShopID(ctx context.Context, shopID string) ([]*models.Frame, error) {
	iter := r.client.Collection("frames").
		Where("shopId", "==", shopID).
		Documents(ctx)

	var frames []*models.Frame
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate frames: %w", err)
		}

		var frame models.Frame
		if err := doc.DataTo(&frame); err != nil {
			continue
		}
		frame.ID = doc.Ref.ID
		frames = append(frames, &frame)
	}

	return frames, nil
}

// CreateFrame creates a new frame configuration
func (r *PrintShopRepository) CreateFrame(ctx context.Context, frame *models.Frame) error {
	if frame.ID == "" {
		frame.ID = uuid.New().String()
	}
	if frame.CreatedAt.IsZero() {
		frame.CreatedAt = time.Now()
	}
	if frame.UpdatedAt.IsZero() {
		frame.UpdatedAt = time.Now()
	}

	_, err := r.client.Collection("frames").Doc(frame.ID).Set(ctx, frame)
	if err != nil {
		return fmt.Errorf("failed to create frame: %w", err)
	}

	return nil
}

// UpdateFrame updates a frame configuration
func (r *PrintShopRepository) UpdateFrame(ctx context.Context, frameID string, updates map[string]interface{}) error {
	updates["updatedAt"] = time.Now()

	// Convert map to []firestore.Update
	var firestoreUpdates []firestore.Update
	for path, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{
			Path:  path,
			Value: value,
		})
	}

	_, err := r.client.Collection("frames").Doc(frameID).Update(ctx, firestoreUpdates)
	if err != nil {
		return fmt.Errorf("failed to update frame: %w", err)
	}

	return nil
}

// DeleteFrame deletes a frame configuration
func (r *PrintShopRepository) DeleteFrame(ctx context.Context, frameID string) error {
	_, err := r.client.Collection("frames").Doc(frameID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete frame: %w", err)
	}

	return nil
}

// ==================== Size Operations ====================

// GetSizesByShopID retrieves all print sizes for a print shop
func (r *PrintShopRepository) GetSizesByShopID(ctx context.Context, shopID string) ([]*models.PrintSize, error) {
	iter := r.client.Collection("sizes").
		Where("shopId", "==", shopID).
		Documents(ctx)

	var sizes []*models.PrintSize
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate sizes: %w", err)
		}

		var size models.PrintSize
		if err := doc.DataTo(&size); err != nil {
			continue
		}
		size.ID = doc.Ref.ID
		sizes = append(sizes, &size)
	}

	return sizes, nil
}

// CreateSize creates a new print size configuration
func (r *PrintShopRepository) CreateSize(ctx context.Context, size *models.PrintSize) error {
	if size.ID == "" {
		size.ID = uuid.New().String()
	}
	if size.CreatedAt.IsZero() {
		size.CreatedAt = time.Now()
	}
	if size.UpdatedAt.IsZero() {
		size.UpdatedAt = time.Now()
	}

	_, err := r.client.Collection("sizes").Doc(size.ID).Set(ctx, size)
	if err != nil {
		return fmt.Errorf("failed to create size: %w", err)
	}

	return nil
}

// UpdateSize updates a print size configuration
func (r *PrintShopRepository) UpdateSize(ctx context.Context, sizeID string, updates map[string]interface{}) error {
	updates["updatedAt"] = time.Now()

	// Convert map to []firestore.Update
	var firestoreUpdates []firestore.Update
	for path, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{
			Path:  path,
			Value: value,
		})
	}

	_, err := r.client.Collection("sizes").Doc(sizeID).Update(ctx, firestoreUpdates)
	if err != nil {
		return fmt.Errorf("failed to update size: %w", err)
	}

	return nil
}

// DeleteSize deletes a print size configuration
func (r *PrintShopRepository) DeleteSize(ctx context.Context, sizeID string) error {
	_, err := r.client.Collection("sizes").Doc(sizeID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete size: %w", err)
	}

	return nil
}

// ==================== Material Operations ====================

// GetMaterialsByShopID retrieves all materials for a print shop
func (r *PrintShopRepository) GetMaterialsByShopID(ctx context.Context, shopID string) ([]*models.Material, error) {
	iter := r.client.Collection("materials").
		Where("shopId", "==", shopID).
		Documents(ctx)

	var materials []*models.Material
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate materials: %w", err)
		}

		var material models.Material
		if err := doc.DataTo(&material); err != nil {
			continue
		}
		material.ID = doc.Ref.ID
		materials = append(materials, &material)
	}

	return materials, nil
}

// CreateMaterial creates a new material configuration
func (r *PrintShopRepository) CreateMaterial(ctx context.Context, material *models.Material) error {
	if material.ID == "" {
		material.ID = uuid.New().String()
	}
	if material.CreatedAt.IsZero() {
		material.CreatedAt = time.Now()
	}
	if material.UpdatedAt.IsZero() {
		material.UpdatedAt = time.Now()
	}

	_, err := r.client.Collection("materials").Doc(material.ID).Set(ctx, material)
	if err != nil {
		return fmt.Errorf("failed to create material: %w", err)
	}

	return nil
}

// UpdateMaterial updates a material configuration
func (r *PrintShopRepository) UpdateMaterial(ctx context.Context, materialID string, updates map[string]interface{}) error {
	updates["updatedAt"] = time.Now()

	// Convert map to []firestore.Update
	var firestoreUpdates []firestore.Update
	for path, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{
			Path:  path,
			Value: value,
		})
	}

	_, err := r.client.Collection("materials").Doc(materialID).Update(ctx, firestoreUpdates)
	if err != nil {
		return fmt.Errorf("failed to update material: %w", err)
	}

	return nil
}

// DeleteMaterial deletes a material configuration
func (r *PrintShopRepository) DeleteMaterial(ctx context.Context, materialID string) error {
	_, err := r.client.Collection("materials").Doc(materialID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete material: %w", err)
	}

	return nil
}

// ==================== Matching Operations ====================

// GetShopsByService retrieves all shops that offer a specific service
func (r *PrintShopRepository) GetShopsByService(ctx context.Context, serviceID string) ([]*models.PrintShopProfile, error) {
	// First get the service to find shops that have it in their services list
	service, err := r.GetServiceByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Get the shop that owns this service
	shop, err := r.GetShopByID(ctx, service.ShopID)
	if err != nil {
		return nil, err
	}

	// Check if service is in shop's services list
	hasService := false
	for _, sID := range shop.Services {
		if sID == serviceID {
			hasService = true
			break
		}
	}

	if !hasService {
		return []*models.PrintShopProfile{}, nil
	}

	return []*models.PrintShopProfile{shop}, nil
}
