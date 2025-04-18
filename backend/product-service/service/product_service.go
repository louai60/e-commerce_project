package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/louai60/e-commerce_project/backend/product-service/cache"
	"github.com/louai60/e-commerce_project/backend/product-service/models"
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	"github.com/louai60/e-commerce_project/backend/product-service/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ProductService handles business logic for products, brands, and categories
type ProductService struct {
	productRepo  repository.ProductRepository
	brandRepo    repository.BrandRepository
	categoryRepo repository.CategoryRepository
	cacheManager *cache.CacheManager
	logger       *zap.Logger
}

// NewProductService creates a new product service
func NewProductService(
	productRepo repository.ProductRepository,
	brandRepo repository.BrandRepository,
	categoryRepo repository.CategoryRepository,
	cacheManager *cache.CacheManager,
	logger *zap.Logger,
) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		brandRepo:    brandRepo,
		categoryRepo: categoryRepo,
		cacheManager: cacheManager,
		logger:       logger,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.Product, error) {
	if req == nil || req.Product == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request: product is required")
	}

	// Validate required fields
	if req.Product.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	product := &models.Product{
		Title:            req.Product.Title,
		Slug:             req.Product.Slug, // Consider generating slug if empty
		Description:      req.Product.Description,
		ShortDescription: req.Product.ShortDescription,
		Price:            req.Product.Price,
		SKU:              req.Product.Sku,
		InventoryQty:     int(req.Product.InventoryQty),
		InventoryStatus:  "in_stock", // Default to in_stock if inventory_qty > 0
		IsPublished:      req.Product.IsPublished,
		CreatedAt:        time.Now().UTC(), // Use UTC
		UpdatedAt:        time.Now().UTC(), // Use UTC
	}

	// Handle nullable fields
	if req.Product.Weight != nil {
		product.Weight = &req.Product.Weight.Value
	}
	if req.Product.BrandId != nil {
		product.BrandID = &req.Product.BrandId.Value
	}

	// Handle variants
	if len(req.Product.Variants) > 0 {
		product.Variants = make([]models.ProductVariant, len(req.Product.Variants))
		for i, variant := range req.Product.Variants {
			productVariant := models.ProductVariant{
				SKU:          variant.Sku,
				Price:        variant.Price,
				InventoryQty: int(variant.InventoryQty),
			}

			// Handle nullable fields
			if variant.Title != "" {
				productVariant.Title = &variant.Title
			}
			if variant.DiscountPrice != nil {
				productVariant.DiscountPrice = &variant.DiscountPrice.Value
			}

			// Handle attributes
			if len(variant.Attributes) > 0 {
				productVariant.Attributes = make([]models.VariantAttributeValue, len(variant.Attributes))
				for j, attr := range variant.Attributes {
					productVariant.Attributes[j] = models.VariantAttributeValue{
						Name:  attr.Name,
						Value: attr.Value,
					}
				}
			}

			product.Variants[i] = productVariant
		}
	}

	// Handle images
	if len(req.Product.Images) > 0 {
		product.Images = make([]models.ProductImage, len(req.Product.Images))
		for i, img := range req.Product.Images {
			product.Images[i] = models.ProductImage{
				URL:      img.Url,
				AltText:  img.AltText,
				Position: int(img.Position),
			}
		}
	}

	// Handle tags
	if len(req.Product.Tags) > 0 {
		product.Tags = make([]models.ProductTag, len(req.Product.Tags))
		for i, tag := range req.Product.Tags {
			product.Tags[i] = models.ProductTag{
				Tag: tag.Tag,
			}
		}
	}

	// Handle specifications if provided
	if len(req.Product.Specifications) > 0 {
		product.Specifications = make([]models.ProductSpecification, len(req.Product.Specifications))
		for i, spec := range req.Product.Specifications {
			product.Specifications[i] = models.ProductSpecification{
				Name:  spec.Name,
				Value: spec.Value,
				Unit:  spec.Unit,
			}
		}
	}

	// Handle attributes if provided
	if len(req.Product.Attributes) > 0 {
		product.Attributes = make([]models.ProductAttribute, len(req.Product.Attributes))
		for i, attr := range req.Product.Attributes {
			product.Attributes[i] = models.ProductAttribute{
				Name:  attr.Name,
				Value: attr.Value,
			}
		}
	}

	// Handle SEO if provided
	if req.Product.Seo != nil {
		product.SEO = &models.ProductSEO{
			MetaTitle:       req.Product.Seo.MetaTitle,
			MetaDescription: req.Product.Seo.MetaDescription,
			Keywords:        req.Product.Seo.Keywords,
			Tags:            req.Product.Seo.Tags,
		}
	}

	// Handle shipping if provided
	if req.Product.Shipping != nil {
		product.Shipping = &models.ProductShipping{
			FreeShipping:     req.Product.Shipping.FreeShipping,
			EstimatedDays:    int(req.Product.Shipping.EstimatedDays),
			ExpressAvailable: req.Product.Shipping.ExpressAvailable,
		}
	}

	// Handle discount if provided
	if req.Product.Discount != nil {
		product.Discount = &models.ProductDiscount{
			Type:  req.Product.Discount.Type,
			Value: req.Product.Discount.Value,
		}
		if req.Product.Discount.ExpiresAt != nil {
			expiresAt := req.Product.Discount.ExpiresAt.AsTime()
			product.Discount.ExpiresAt = &expiresAt
		}
	}

	// Handle inventory locations if provided
	if len(req.Product.InventoryLocations) > 0 {
		product.InventoryLocations = make([]models.InventoryLocation, len(req.Product.InventoryLocations))
		for i, loc := range req.Product.InventoryLocations {
			product.InventoryLocations[i] = models.InventoryLocation{
				WarehouseID:  loc.WarehouseId,
				AvailableQty: int(loc.AvailableQty),
			}
		}
	}

	if err := s.productRepo.CreateProduct(ctx, product); err != nil {
		s.logger.Error("Failed to create product", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	// Create related entities after the product is created
	// These need the product ID which is generated during product creation
	// Create specifications
	for i := range product.Specifications {
		spec := &product.Specifications[i]
		spec.ProductID = product.ID
		if err := s.productRepo.AddProductSpecification(ctx, spec); err != nil {
			s.logger.Error("Failed to create product specification",
				zap.Error(err),
				zap.String("product_id", product.ID),
				zap.String("spec_name", spec.Name))
			// Continue even if one specification fails
		}
	}

	// Create attributes
	for i := range product.Attributes {
		attr := &product.Attributes[i]
		attr.ProductID = product.ID
		if err := s.productRepo.AddProductAttribute(ctx, attr); err != nil {
			s.logger.Error("Failed to create product attribute",
				zap.Error(err),
				zap.String("product_id", product.ID),
				zap.String("attr_name", attr.Name))
			// Continue even if one attribute fails
		}
	}

	// Create SEO
	if product.SEO != nil {
		product.SEO.ProductID = product.ID
		if err := s.productRepo.UpsertProductSEO(ctx, product.SEO); err != nil {
			s.logger.Error("Failed to create product SEO",
				zap.Error(err),
				zap.String("product_id", product.ID))
			// Continue even if SEO fails
		}
	}

	// Create shipping
	if product.Shipping != nil {
		product.Shipping.ProductID = product.ID
		if err := s.productRepo.UpsertProductShipping(ctx, product.Shipping); err != nil {
			s.logger.Error("Failed to create product shipping",
				zap.Error(err),
				zap.String("product_id", product.ID))
			// Continue even if shipping fails
		}
	}

	// Create discount
	if product.Discount != nil {
		product.Discount.ProductID = product.ID
		if err := s.productRepo.AddProductDiscount(ctx, product.Discount); err != nil {
			s.logger.Error("Failed to create product discount",
				zap.Error(err),
				zap.String("product_id", product.ID))
			// Continue even if discount fails
		}
	}

	// Create inventory locations
	for i := range product.InventoryLocations {
		loc := &product.InventoryLocations[i]
		loc.ProductID = product.ID
		if err := s.productRepo.UpsertInventoryLocation(ctx, loc); err != nil {
			s.logger.Error("Failed to create inventory location",
				zap.Error(err),
				zap.String("product_id", product.ID),
				zap.String("warehouse_id", loc.WarehouseID))
			// Continue even if one location fails
		}
	}

	// Enhanced cache invalidation
	if err := s.cacheManager.InvalidateProductAndRelated(ctx, product.ID); err != nil {
		s.logger.Warn("Failed to invalidate caches after product creation",
			zap.String("id", product.ID),
			zap.Error(err))
	}

	// Fetch the complete product with all related entities to return
	createdProduct, err := s.productRepo.GetByID(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get created product", zap.Error(err))
		// Still return the basic product even if we can't fetch the complete one
		return convertModelToProto(product), nil
	}

	// Populate all related entities
	if err := s.populateProductRelations(ctx, createdProduct); err != nil {
		s.logger.Error("Failed to populate product relations", zap.Error(err))
		// Continue even if population fails
	}

	return convertModelToProto(createdProduct), nil
}

func (s *ProductService) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	var product *models.Product
	var err error

	// Try cache first
	if id := req.GetId(); id != "" {
		product, err = s.cacheManager.GetProduct(ctx, id)
		if err == nil {
			s.logger.Debug("Cache hit for product", zap.String("id", id))
			return convertModelToProto(product), nil
		}
	}

	// Cache miss or slug lookup, get from database
	switch identifier := req.Identifier.(type) {
	case *pb.GetProductRequest_Id:
		product, err = s.productRepo.GetByID(ctx, identifier.Id)
	case *pb.GetProductRequest_Slug:
		product, err = s.productRepo.GetBySlug(ctx, identifier.Slug)
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid identifier")
	}

	if err != nil {
		s.logger.Error("Failed to get product", zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "product not found")
	}

	// Populate related entities
	if err := s.populateProductRelations(ctx, product); err != nil {
		s.logger.Error("Failed to populate product relations", zap.Error(err))
		// Continue even if population fails
	}

	// Cache the result
	if err := s.cacheManager.SetProduct(ctx, product); err != nil {
		s.logger.Warn("Failed to cache product", zap.Error(err))
	}

	return convertModelToProto(product), nil
}

func (s *ProductService) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	// Generate cache key from pagination parameters
	cacheKey := fmt.Sprintf("page:%d:limit:%d", req.Page, req.Limit)

	// Try cache first
	products, err := s.cacheManager.GetProductList(ctx, cacheKey)
	if err == nil {
		s.logger.Debug("Cache hit for product list", zap.String("key", cacheKey))
		return &pb.ListProductsResponse{
			Products: convertProductModelsToProtos(products),
			Total:    int32(len(products)),
		}, nil
	}

	// Cache miss, get from database
	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0 // Ensure offset is not negative
	}
	products, total, err := s.productRepo.List(ctx, int(offset), int(req.Limit))
	if err != nil {
		s.logger.Error("Failed to list products", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list products")
	}

	// Cache the result
	if err := s.cacheManager.SetProductList(ctx, cacheKey, products); err != nil {
		s.logger.Warn("Failed to cache product list", zap.Error(err))
	}

	return &pb.ListProductsResponse{
		Products: convertProductModelsToProtos(products),
		Total:    int32(total),
	}, nil
}

// --- Conversion Helper Functions ---

func convertModelToProto(model *models.Product) *pb.Product {
	if model == nil {
		return nil
	}
	protoProduct := &pb.Product{
		Id:                 model.ID,
		Title:              model.Title,
		Slug:               model.Slug,
		Description:        model.Description,
		ShortDescription:   model.ShortDescription,
		Price:              model.Price,
		Sku:                model.SKU,
		InventoryQty:       int32(model.InventoryQty),
		InventoryStatus:    model.InventoryStatus,
		IsPublished:        model.IsPublished,
		CreatedAt:          timestamppb.New(model.CreatedAt),
		UpdatedAt:          timestamppb.New(model.UpdatedAt),
		Brand:              convertBrandModelToProto(model.Brand),                            // Convert Brand
		Images:             convertImageModelsToProtos(model.Images),                         // Convert Images
		Categories:         convertCategorySliceToProtos(model.Categories),                   // Convert Categories
		Variants:           convertVariantModelsToProtos(model.Variants),                     // Convert Variants
		Tags:               convertTagModelsToProtos(model.Tags),                             // Convert Tags
		Attributes:         convertProductAttributeModelsToProtos(model.Attributes),          // Convert Attributes
		Specifications:     convertSpecificationModelsToProtos(model.Specifications),         // Convert Specifications
		Seo:                convertSEOModelToProto(model.SEO),                                // Convert SEO
		Shipping:           convertShippingModelToProto(model.Shipping),                      // Convert Shipping
		Discount:           convertDiscountModelToProto(model.Discount),                      // Convert Discount
		InventoryLocations: convertInventoryLocationModelsToProtos(model.InventoryLocations), // Convert Inventory Locations
	}

	// Handle nullable fields
	if model.DiscountPrice != nil {
		protoProduct.DiscountPrice = wrapperspb.Double(*model.DiscountPrice)
	}
	if model.Weight != nil {
		protoProduct.Weight = wrapperspb.Double(*model.Weight)
	}
	if model.BrandID != nil {
		protoProduct.BrandId = wrapperspb.String(*model.BrandID)
	}
	if model.DefaultVariantID != nil {
		protoProduct.DefaultVariantId = wrapperspb.String(*model.DefaultVariantID)
	}

	return protoProduct
}

func convertBrandModelToProto(model *models.Brand) *pb.Brand {
	if model == nil {
		return nil
	}
	return &pb.Brand{
		Id:          model.ID,
		Name:        model.Name,
		Slug:        model.Slug,
		Description: model.Description,
		CreatedAt:   timestamppb.New(model.CreatedAt),
		UpdatedAt:   timestamppb.New(model.UpdatedAt),
	}
}

func convertSingleCategoryModelToProto(model *models.Category) *pb.Category {
	if model == nil {
		return nil
	}
	protoCategory := &pb.Category{
		Id:          model.ID,
		Name:        model.Name,
		Slug:        model.Slug,
		Description: model.Description,
		CreatedAt:   timestamppb.New(model.CreatedAt),
		UpdatedAt:   timestamppb.New(model.UpdatedAt),
	}
	if model.ParentID != nil {
		protoCategory.ParentId = wrapperspb.String(*model.ParentID)
	}
	return protoCategory
}

func convertImageModelToProto(model *models.ProductImage) *pb.ProductImage {
	if model == nil {
		return nil
	}
	return &pb.ProductImage{
		Id:        model.ID,
		ProductId: model.ProductID,
		Url:       model.URL,
		AltText:   model.AltText,
		Position:  int32(model.Position),
		CreatedAt: timestamppb.New(model.CreatedAt),
		UpdatedAt: timestamppb.New(model.UpdatedAt),
	}
}

func convertImageModelsToProtos(models []models.ProductImage) []*pb.ProductImage {
	protos := make([]*pb.ProductImage, len(models))
	for i, model := range models {
		protos[i] = convertImageModelToProto(&model) // Pass address if model is value type in slice
	}
	return protos
}

func convertCategorySliceToProtos(models []models.Category) []*pb.Category {
	protos := make([]*pb.Category, len(models))
	for i, model := range models {
		protos[i] = convertCategoryModelToProto(&model) // Pass address if model is value type in slice
	}
	return protos
}

func convertProductModelsToProtos(models []*models.Product) []*pb.Product {
	protos := make([]*pb.Product, len(models))
	for i, model := range models {
		protos[i] = convertModelToProto(model)
	}
	return protos
}

func convertVariantModelToProto(model models.ProductVariant) *pb.ProductVariant {
	protoVariant := &pb.ProductVariant{
		Id:           model.ID,
		ProductId:    model.ProductID,
		Sku:          model.SKU,
		Price:        model.Price,
		InventoryQty: int32(model.InventoryQty),
		CreatedAt:    timestamppb.New(model.CreatedAt),
		UpdatedAt:    timestamppb.New(model.UpdatedAt),
	}

	// Handle nullable fields
	if model.Title != nil {
		protoVariant.Title = *model.Title
	}
	if model.DiscountPrice != nil {
		protoVariant.DiscountPrice = wrapperspb.Double(*model.DiscountPrice)
	}

	// Convert attributes
	if len(model.Attributes) > 0 {
		protoVariant.Attributes = make([]*pb.VariantAttributeValue, len(model.Attributes))
		for i, attr := range model.Attributes {
			protoVariant.Attributes[i] = &pb.VariantAttributeValue{
				Name:  attr.Name,
				Value: attr.Value,
			}
		}
	}

	return protoVariant
}

func convertVariantModelsToProtos(models []models.ProductVariant) []*pb.ProductVariant {
	protos := make([]*pb.ProductVariant, len(models))
	for i, model := range models {
		protos[i] = convertVariantModelToProto(model)
	}
	return protos
}

// Convert ProductTag models to protos
func convertTagModelsToProtos(models []models.ProductTag) []*pb.ProductTag {
	if len(models) == 0 {
		return nil
	}
	protos := make([]*pb.ProductTag, len(models))
	for i, model := range models {
		protos[i] = &pb.ProductTag{
			Id:        model.ID,
			ProductId: model.ProductID,
			Tag:       model.Tag,
			CreatedAt: timestamppb.New(model.CreatedAt),
			UpdatedAt: timestamppb.New(model.UpdatedAt),
		}
	}
	return protos
}

// Convert ProductAttribute models to protos
func convertProductAttributeModelsToProtos(models []models.ProductAttribute) []*pb.ProductAttribute {
	if len(models) == 0 {
		return nil
	}
	protos := make([]*pb.ProductAttribute, len(models))
	for i, model := range models {
		protos[i] = &pb.ProductAttribute{
			Id:        model.ID,
			ProductId: model.ProductID,
			Name:      model.Name,
			Value:     model.Value,
			CreatedAt: timestamppb.New(model.CreatedAt),
			UpdatedAt: timestamppb.New(model.UpdatedAt),
		}
	}
	return protos
}

// Convert ProductSpecification models to protos
func convertSpecificationModelsToProtos(models []models.ProductSpecification) []*pb.ProductSpecification {
	if len(models) == 0 {
		return nil
	}
	protos := make([]*pb.ProductSpecification, len(models))
	for i, model := range models {
		protos[i] = &pb.ProductSpecification{
			Id:        model.ID,
			ProductId: model.ProductID,
			Name:      model.Name,
			Value:     model.Value,
			Unit:      model.Unit,
			CreatedAt: timestamppb.New(model.CreatedAt),
			UpdatedAt: timestamppb.New(model.UpdatedAt),
		}
	}
	return protos
}

// Convert ProductSEO model to proto
func convertSEOModelToProto(model *models.ProductSEO) *pb.ProductSEO {
	if model == nil {
		return nil
	}
	return &pb.ProductSEO{
		Id:              model.ID,
		ProductId:       model.ProductID,
		MetaTitle:       model.MetaTitle,
		MetaDescription: model.MetaDescription,
		Keywords:        model.Keywords,
		Tags:            model.Tags,
		CreatedAt:       timestamppb.New(model.CreatedAt),
		UpdatedAt:       timestamppb.New(model.UpdatedAt),
	}
}

// Convert ProductShipping model to proto
func convertShippingModelToProto(model *models.ProductShipping) *pb.ProductShipping {
	if model == nil {
		return nil
	}
	return &pb.ProductShipping{
		Id:               model.ID,
		ProductId:        model.ProductID,
		FreeShipping:     model.FreeShipping,
		EstimatedDays:    int32(model.EstimatedDays),
		ExpressAvailable: model.ExpressAvailable,
		CreatedAt:        timestamppb.New(model.CreatedAt),
		UpdatedAt:        timestamppb.New(model.UpdatedAt),
	}
}

// Convert ProductDiscount model to proto
func convertDiscountModelToProto(model *models.ProductDiscount) *pb.ProductDiscount {
	if model == nil {
		return nil
	}
	proto := &pb.ProductDiscount{
		Id:        model.ID,
		ProductId: model.ProductID,
		Type:      model.Type,
		Value:     model.Value,
		CreatedAt: timestamppb.New(model.CreatedAt),
		UpdatedAt: timestamppb.New(model.UpdatedAt),
	}

	if model.ExpiresAt != nil {
		proto.ExpiresAt = timestamppb.New(*model.ExpiresAt)
	}

	return proto
}

// Convert InventoryLocation models to protos
func convertInventoryLocationModelsToProtos(models []models.InventoryLocation) []*pb.InventoryLocation {
	if len(models) == 0 {
		return nil
	}
	protos := make([]*pb.InventoryLocation, len(models))
	for i, model := range models {
		protos[i] = &pb.InventoryLocation{
			Id:           model.ID,
			ProductId:    model.ProductID,
			WarehouseId:  model.WarehouseID,
			AvailableQty: int32(model.AvailableQty),
			CreatedAt:    timestamppb.New(model.CreatedAt),
			UpdatedAt:    timestamppb.New(model.UpdatedAt),
		}
	}
	return protos
}

// --- Service Methods (Stubs for missing ones) ---

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.Product, error) {
	if req == nil || req.Product == nil || req.Product.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request: product ID is required")
	}

	productID := req.Product.Id
	s.logger.Info("UpdateProduct service method called", zap.String("id", productID))

	// 1. Get existing product
	existingProduct, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product with ID %s not found", productID)
	}

	// 2. Update base product
	updatedProduct := convertProtoToModelForUpdate(req.Product, existingProduct)
	updatedProduct.UpdatedAt = time.Now().UTC()

	if err := s.productRepo.UpdateProduct(ctx, updatedProduct); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update product: %v", err)
	}

	// 3. Handle variants if provided
	if len(req.Product.Variants) > 0 {
		// Start transaction for variant operations
		tx, err := s.productRepo.BeginTx(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
		}
		defer tx.Rollback()

		if err := s.updateProductVariants(ctx, tx, productID, req.Product.Variants); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update variants: %v", err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
		}
	}

	// 4. Invalidate cache
	if err := s.cacheManager.InvalidateProductAndRelated(ctx, productID); err != nil {
		s.logger.Warn("Failed to invalidate caches", zap.String("id", productID), zap.Error(err))
	}

	// 5. Return updated product
	return s.GetProduct(ctx, &pb.GetProductRequest{
		Identifier: &pb.GetProductRequest_Id{Id: productID},
	})
}

func (s *ProductService) updateProductVariants(ctx context.Context, tx *sql.Tx, productID string, variants []*pb.ProductVariant) error {
	// 1. Get existing variants
	existingVariants, err := s.productRepo.GetProductVariants(ctx, productID)
	if err != nil {
		return err
	}

	// Create maps for easier lookup
	existingVariantMap := make(map[string]*models.ProductVariant)
	for _, v := range existingVariants {
		existingVariantMap[v.ID] = v
	}

	// 2. Process each variant
	for _, variant := range variants {
		if variant.Id == "" {
			// New variant
			if err := s.productRepo.CreateVariant(ctx, tx, productID, convertProtoToVariantModel(variant)); err != nil {
				return err
			}
		} else {
			// Update existing variant
			if err := s.productRepo.UpdateVariant(ctx, tx, convertProtoToVariantModel(variant)); err != nil {
				return err
			}
			delete(existingVariantMap, variant.Id)
		}
	}

	// 3. Delete variants that weren't included in the update
	for variantID := range existingVariantMap {
		if err := s.productRepo.DeleteVariant(ctx, tx, variantID); err != nil {
			return err
		}
	}

	return nil
}

// convertProtoToVariantModel converts a proto ProductVariant to a model ProductVariant
func convertProtoToVariantModel(proto *pb.ProductVariant) *models.ProductVariant {
	if proto == nil {
		return nil
	}

	variant := &models.ProductVariant{
		ID:           proto.Id,
		ProductID:    proto.ProductId,
		SKU:          proto.Sku,
		Price:        proto.Price,
		InventoryQty: int(proto.InventoryQty),
	}

	// Handle nullable fields
	if proto.Title != "" {
		variant.Title = &proto.Title
	}
	if proto.DiscountPrice != nil {
		variant.DiscountPrice = &proto.DiscountPrice.Value
	}

	// Convert attributes
	if len(proto.Attributes) > 0 {
		variant.Attributes = make([]models.VariantAttributeValue, len(proto.Attributes))
		for i, attr := range proto.Attributes {
			variant.Attributes[i] = models.VariantAttributeValue{
				Name:  attr.Name,
				Value: attr.Value,
			}
		}
	}

	return variant
}

// DeleteProduct deletes a product by its ID
func (s *ProductService) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	// Delete product (cascade will handle variants and attributes)
	if err := s.productRepo.DeleteProduct(ctx, req.Id); err != nil {
		if err == models.ErrProductNotFound {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete product: %v", err)
	}

	// Invalidate cache
	if err := s.cacheManager.InvalidateProductAndRelated(ctx, req.Id); err != nil {
		s.logger.Warn("Failed to invalidate caches", zap.String("id", req.Id), zap.Error(err))
	}

	return &pb.DeleteProductResponse{Success: true}, nil
}

func (s *ProductService) CreateBrand(ctx context.Context, brand *pb.Brand) (*pb.Brand, error) {
	s.logger.Info("CreateBrand service method called", zap.String("name", brand.Name))

	// Validate request
	if brand.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "brand name is required")
	}
	if brand.Slug == "" {
		return nil, status.Error(codes.InvalidArgument, "brand slug is required")
	}

	// Convert proto to model
	brandModel := &models.Brand{
		Name:        brand.Name,
		Slug:        brand.Slug,
		Description: brand.Description,
	}

	// Save to database
	if err := s.brandRepo.CreateBrand(ctx, brandModel); err != nil {
		s.logger.Error("Failed to create brand", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create brand: %v", err)
	}

	// Convert model back to proto
	return &pb.Brand{
		Id:          brandModel.ID,
		Name:        brandModel.Name,
		Slug:        brandModel.Slug,
		Description: brandModel.Description,
		CreatedAt:   timestamppb.New(brandModel.CreatedAt),
		UpdatedAt:   timestamppb.New(brandModel.UpdatedAt),
	}, nil
}

func (s *ProductService) GetBrand(ctx context.Context, req *pb.GetBrandRequest) (*pb.Brand, error) {
	var brand *models.Brand
	var err error
	var cacheKey string

	// Determine identifier and prepare cache key
	switch identifier := req.Identifier.(type) {
	case *pb.GetBrandRequest_Id:
		if identifier.Id == "" {
			return nil, status.Error(codes.InvalidArgument, "brand ID cannot be empty")
		}
		cacheKey = fmt.Sprintf("brand:%s", identifier.Id)
		s.logger.Info("GetBrand service method called", zap.String("id", identifier.Id))
	case *pb.GetBrandRequest_Slug:
		if identifier.Slug == "" {
			return nil, status.Error(codes.InvalidArgument, "brand slug cannot be empty")
		}
		cacheKey = fmt.Sprintf("brand:slug:%s", identifier.Slug)
		s.logger.Info("GetBrand service method called", zap.String("slug", identifier.Slug))
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid identifier provided for GetBrand")
	}

	// Try cache first
	brand, err = s.cacheManager.GetBrand(ctx, cacheKey)
	if err == nil {
		s.logger.Debug("Cache hit for brand", zap.String("key", cacheKey))
		return convertBrandModelToProto(brand), nil
	}
	s.logger.Debug("Cache miss for brand", zap.String("key", cacheKey), zap.Error(err))

	// Cache miss, get from database
	switch identifier := req.Identifier.(type) {
	case *pb.GetBrandRequest_Id:
		brand, err = s.brandRepo.GetBrandByID(ctx, identifier.Id)
	case *pb.GetBrandRequest_Slug:
		brand, err = s.brandRepo.GetBrandBySlug(ctx, identifier.Slug)
		// No default needed here as it's handled above
	}

	if err != nil {
		// Differentiate between not found and other errors
		if err.Error() == "brand not found" { // Assuming repo returns this specific error string
			s.logger.Warn("Brand not found in DB", zap.Any("identifier", req.Identifier))
			return nil, status.Errorf(codes.NotFound, "brand not found")
		}
		s.logger.Error("Failed to get brand from repository", zap.Any("identifier", req.Identifier), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get brand: %v", err)
	}

	// Cache the result
	if err := s.cacheManager.SetBrand(ctx, cacheKey, brand); err != nil {
		s.logger.Warn("Failed to cache brand", zap.String("key", cacheKey), zap.Error(err))
		// Continue even if caching fails
	}

	return convertBrandModelToProto(brand), nil
}

func (s *ProductService) ListBrands(ctx context.Context, req *pb.ListBrandsRequest) (*pb.ListBrandsResponse, error) {
	s.logger.Info("ListBrands service method called")
	s.logger.Info("ListBrands service method called")

	// Define cache key based on pagination (if any)
	// For now, assuming no pagination for brands, cache all.
	cacheKey := "brands:all" // Adjust if pagination is added

	// Try cache first
	cachedBrands, err := s.cacheManager.GetBrandList(ctx, cacheKey)
	if err == nil {
		s.logger.Debug("Cache hit for brand list", zap.String("key", cacheKey))
		return &pb.ListBrandsResponse{
			Brands: convertBrandModelsToProtos(cachedBrands),
			Total:  int32(len(cachedBrands)), // Assuming total is the count of cached items
		}, nil
	}
	s.logger.Debug("Cache miss for brand list", zap.String("key", cacheKey), zap.Error(err))

	// Cache miss, get from database
	// TODO: Implement pagination in repository if needed. For now, list all.
	brands, total, err := s.brandRepo.ListBrands(ctx, 0, 0) // Assuming ListBrands(ctx, offset, limit) - 0, 0 means list all
	if err != nil {
		s.logger.Error("Failed to list brands from repository", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list brands: %v", err)
	}

	// Cache the result
	if err := s.cacheManager.SetBrandList(ctx, cacheKey, brands); err != nil {
		s.logger.Warn("Failed to cache brand list", zap.String("key", cacheKey), zap.Error(err))
		// Continue even if caching fails
	}

	return &pb.ListBrandsResponse{
		Brands: convertBrandModelsToProtos(brands),
		Total:  int32(total),
	}, nil
}

// Helper function to convert multiple brand models to protos
func convertBrandModelsToProtos(models []*models.Brand) []*pb.Brand {
	if models == nil {
		return nil
	}
	protos := make([]*pb.Brand, len(models))
	for i, model := range models {
		protos[i] = convertBrandModelToProto(model)
	}
	return protos
}

// CreateCategory implements the category creation endpoint
func (s *ProductService) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.Category, error) {
	s.logger.Info("Creating new category", zap.String("name", req.Category.Name))

	category := &models.Category{
		Name:        req.Category.Name,
		Slug:        req.Category.Slug,
		Description: req.Category.Description,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Handle optional parent ID
	if req.Category.ParentId != nil {
		category.ParentID = &req.Category.ParentId.Value

		// Verify parent exists
		parent, err := s.categoryRepo.GetCategoryByID(ctx, *category.ParentID)
		if err != nil {
			s.logger.Error("Parent category not found",
				zap.String("parent_id", *category.ParentID),
				zap.Error(err))
			return nil, status.Errorf(codes.NotFound, "parent category not found")
		}
		category.ParentName = parent.Name
	}

	// Create the category
	err := s.categoryRepo.CreateCategory(ctx, category)
	if err != nil {
		s.logger.Error("Failed to create category", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create category: %v", err)
	}

	// Invalidate category cache
	if err := s.cacheManager.InvalidateCategoryLists(ctx); err != nil {
		s.logger.Warn("Failed to invalidate category cache", zap.Error(err))
	}

	return convertCategoryModelToProto(category), nil
}

// GetCategory implements the category retrieval endpoint
func (s *ProductService) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.Category, error) {
	var category *models.Category
	var err error

	// Try cache first
	if id := req.GetId(); id != "" {
		category, err = s.cacheManager.GetCategory(ctx, id)
		if err == nil {
			s.logger.Debug("Cache hit for category", zap.String("id", id))
			return convertCategoryModelToProto(category), nil
		}
	}

	// Cache miss or slug lookup, get from database
	switch identifier := req.Identifier.(type) {
	case *pb.GetCategoryRequest_Id:
		category, err = s.categoryRepo.GetCategoryByID(ctx, identifier.Id)
	case *pb.GetCategoryRequest_Slug:
		category, err = s.categoryRepo.GetCategoryBySlug(ctx, identifier.Slug)
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid identifier")
	}

	if err != nil {
		s.logger.Error("Failed to get category", zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "category not found")
	}

	// Cache the result
	if err := s.cacheManager.SetCategory(ctx, category); err != nil {
		s.logger.Warn("Failed to cache category", zap.Error(err))
	}

	return convertCategoryModelToProto(category), nil
}

// ListCategories implements the category listing endpoint
func (s *ProductService) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	// Generate cache key from pagination parameters
	cacheKey := fmt.Sprintf("categories:page:%d:limit:%d", req.Page, req.Limit)

	// Try cache first
	categories, err := s.cacheManager.GetCategoryList(ctx, cacheKey)
	if err == nil {
		s.logger.Debug("Cache hit for category list", zap.String("key", cacheKey))
		return &pb.ListCategoriesResponse{
			Categories: convertCategoryModelsToProtos(categories),
			Total:      int32(len(categories)),
		}, nil
	}

	// Cache miss, get from database
	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0 // Ensure offset is not negative
	}

	categories, total, err := s.categoryRepo.ListCategories(ctx, int(offset), int(req.Limit))
	if err != nil {
		s.logger.Error("Failed to list categories", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list categories")
	}

	// Cache the result
	if err := s.cacheManager.SetCategoryList(ctx, cacheKey, categories); err != nil {
		s.logger.Warn("Failed to cache category list", zap.Error(err))
	}

	return &pb.ListCategoriesResponse{
		Categories: convertCategoryModelsToProtos(categories),
		Total:      int32(total),
	}, nil
}

// Helper function to convert multiple categories
func convertCategoryModelsToProtos(categories []*models.Category) []*pb.Category {
	if categories == nil {
		return nil
	}

	protos := make([]*pb.Category, len(categories))
	for i, category := range categories {
		protos[i] = convertCategoryModelToProto(category)
	}
	return protos
}

// Helper function to convert a single category
func convertCategoryModelToProto(model *models.Category) *pb.Category {
	if model == nil {
		return nil
	}

	protoCategory := &pb.Category{
		Id:          model.ID,
		Name:        model.Name,
		Slug:        model.Slug,
		Description: model.Description,
		ParentName:  model.ParentName,
		CreatedAt:   timestamppb.New(model.CreatedAt),
		UpdatedAt:   timestamppb.New(model.UpdatedAt),
	}

	if model.ParentID != nil {
		protoCategory.ParentId = wrapperspb.String(*model.ParentID)
	}

	return protoCategory
}

// populateProductRelations populates related entities for a product
func (s *ProductService) populateProductRelations(ctx context.Context, product *models.Product) error {
	if product == nil {
		return fmt.Errorf("cannot populate relations for nil product")
	}

	// Get variants
	variants, err := s.productRepo.GetProductVariants(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get product variants", zap.Error(err), zap.String("product_id", product.ID))
		return fmt.Errorf("failed to get product variants: %w", err)
	}

	// Convert variants to the expected format
	product.Variants = make([]models.ProductVariant, len(variants))
	for i, v := range variants {
		product.Variants[i] = *v
	}

	// Get tags
	tags, err := s.productRepo.GetProductTags(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get product tags", zap.Error(err), zap.String("product_id", product.ID))
		// Continue even if tags fail to load
	} else {
		product.Tags = tags
	}

	// Get attributes
	attributes, err := s.productRepo.GetProductAttributes(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get product attributes", zap.Error(err), zap.String("product_id", product.ID))
		// Continue even if attributes fail to load
	} else {
		product.Attributes = attributes
	}

	// Get specifications
	specs, err := s.productRepo.GetProductSpecifications(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get product specifications", zap.Error(err), zap.String("product_id", product.ID))
		// Continue even if specifications fail to load
	} else {
		product.Specifications = specs
	}

	// Get SEO
	seo, err := s.productRepo.GetProductSEO(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get product SEO", zap.Error(err), zap.String("product_id", product.ID))
		// Continue even if SEO fails to load
	} else {
		product.SEO = seo
	}

	// Get shipping
	shipping, err := s.productRepo.GetProductShipping(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get product shipping", zap.Error(err), zap.String("product_id", product.ID))
		// Continue even if shipping fails to load
	} else {
		product.Shipping = shipping
	}

	// Get discounts
	discounts, err := s.productRepo.GetProductDiscounts(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get product discounts", zap.Error(err), zap.String("product_id", product.ID))
		// Continue even if discounts fail to load
	} else if len(discounts) > 0 {
		// Use the first active discount
		now := time.Now()
		for _, discount := range discounts {
			if discount.ExpiresAt == nil || discount.ExpiresAt.After(now) {
				product.Discount = &discount
				break
			}
		}
	}

	// Get inventory locations
	locations, err := s.productRepo.GetInventoryLocations(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get inventory locations", zap.Error(err), zap.String("product_id", product.ID))
		// Continue even if inventory locations fail to load
	} else {
		product.InventoryLocations = locations
	}

	return nil
}

// convertProtoToModelForUpdate applies updates from a proto Product to an existing model Product.
// It only updates fields that are explicitly provided in the proto message.
func convertProtoToModelForUpdate(proto *pb.Product, model *models.Product) *models.Product {
	if proto == nil || model == nil {
		return model // Should not happen if called correctly
	}

	// Update only if fields are provided in the proto
	if proto.Title != "" {
		model.Title = proto.Title
	}
	if proto.Slug != "" {
		model.Slug = proto.Slug
	}
	if proto.Description != "" {
		model.Description = proto.Description
	}
	if proto.ShortDescription != "" {
		model.ShortDescription = proto.ShortDescription
	}
	// Price is not nullable in proto, assume 0 is a valid value if intended
	model.Price = proto.Price

	if proto.Sku != "" {
		model.SKU = proto.Sku // Corrected field name
	}
	// InventoryQty is not nullable in proto
	model.InventoryQty = int(proto.InventoryQty)

	// Update inventory status if provided
	if proto.InventoryStatus != "" {
		model.InventoryStatus = proto.InventoryStatus
	}

	// IsPublished is not nullable in proto
	model.IsPublished = proto.IsPublished

	// Handle nullable fields from proto
	if proto.DiscountPrice != nil {
		model.DiscountPrice = &proto.DiscountPrice.Value
	} else {
		model.DiscountPrice = nil // Explicitly set to nil if not provided
	}

	if proto.Weight != nil {
		model.Weight = &proto.Weight.Value
	} else {
		model.Weight = nil // Explicitly set to nil if not provided
	}

	if proto.BrandId != nil {
		model.BrandID = &proto.BrandId.Value
	} else {
		model.BrandID = nil // Explicitly set to nil if not provided
	}

	// TODO: Handle updates for Categories and Images if needed
	// This would involve comparing the incoming IDs/data with existing ones
	// and calling appropriate repository methods (e.g., AddProductCategory, RemoveProductCategory)

	return model
}
