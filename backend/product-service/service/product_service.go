package service

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
	"github.com/louai60/e-commerce_project/backend/product-service/cache"
	"github.com/louai60/e-commerce_project/backend/product-service/clients"
	"github.com/louai60/e-commerce_project/backend/product-service/models"
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	"github.com/louai60/e-commerce_project/backend/product-service/repository"
	"github.com/louai60/e-commerce_project/backend/product-service/storage"
	"github.com/louai60/e-commerce_project/backend/product-service/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ProductService handles business logic for products, brands, and categories
type ProductService struct {
	productRepo     repository.ProductRepository
	brandRepo       repository.BrandRepository
	categoryRepo    repository.CategoryRepository
	cacheManager    cache.CacheInterface
	logger          *zap.Logger
	cld             *cloudinary.Cloudinary
	inventoryClient *clients.InventoryClient
}

// NewProductService creates a new product service
func NewProductService(
	productRepo repository.ProductRepository,
	brandRepo repository.BrandRepository,
	categoryRepo repository.CategoryRepository,
	cacheManager cache.CacheInterface,
	logger *zap.Logger,
	inventoryClient *clients.InventoryClient,
) *ProductService {
	// Initialize Cloudinary
	var cld *cloudinary.Cloudinary
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName != "" && apiKey != "" && apiSecret != "" {
		var err error
		cld, err = cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
		if err != nil {
			logger.Error("Failed to initialize Cloudinary", zap.Error(err))
		}
	}

	return &ProductService{
		productRepo:     productRepo,
		brandRepo:       brandRepo,
		categoryRepo:    categoryRepo,
		cacheManager:    cacheManager,
		logger:          logger,
		cld:             cld,
		inventoryClient: inventoryClient,
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

	// Generate UUID for the product
	productID := uuid.New().String()

	// Log the start of product creation
	s.logger.Info("Creating new product", zap.String("title", req.Product.Title), zap.String("id", productID))

	product := &models.Product{
		ID:               productID,
		Title:            req.Product.Title,
		Slug:             req.Product.Slug, // Consider generating slug if empty
		Description:      req.Product.Description,
		ShortDescription: req.Product.ShortDescription,
		Price:            models.Price{Amount: req.Product.Price, Currency: "USD"}, // Default to USD
		SKU:              req.Product.Sku,
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
	if req.Product.DiscountPrice != nil {
		product.DiscountPrice = &models.Price{
			Amount:   req.Product.DiscountPrice.Value,
			Currency: "USD", // Changed from TND to USD for consistency
		}
	}

	// Process product images
	if len(req.Product.Images) > 0 {
		product.Images = make([]models.ProductImage, len(req.Product.Images))
		for i, img := range req.Product.Images {
			product.Images[i] = models.ProductImage{
				ProductID: productID,
				URL:       img.Url,
				AltText:   img.AltText,
				Position:  int(img.Position),
			}
		}
	}

	// Process categories if provided
	if len(req.Product.Categories) > 0 {
		product.Categories = make([]models.Category, len(req.Product.Categories))
		for i, cat := range req.Product.Categories {
			// Log the category being processed
			s.logger.Info("Processing category for product",
				zap.String("category_id", cat.Id),
				zap.String("product_id", productID))

			product.Categories[i] = models.Category{
				ID: cat.Id,
			}

			// If we have more category details, add them
			if cat.Name != "" {
				product.Categories[i].Name = cat.Name
			}
			if cat.Slug != "" {
				product.Categories[i].Slug = cat.Slug
			}
			if cat.Description != "" {
				product.Categories[i].Description = cat.Description
			}
		}
	}

	// Create the product
	if err := s.productRepo.CreateProduct(ctx, product); err != nil {
		s.logger.Error("Failed to create product", zap.Error(err))
		if err == models.ErrProductSlugExists {
			return nil, status.Errorf(codes.AlreadyExists, "product with this slug already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	// NOTE: We're no longer creating inventory items directly from the product service
	// This is now handled by the API gateway to avoid conflicts with inventory quantities
	// The API gateway will create inventory items with the correct initial quantity
	// If this product is being created directly through the product service (not via API gateway),
	// you'll need to create the inventory item separately through the inventory service

	// Create default variant if no variants are provided
	if len(req.Product.Variants) == 0 {
		defaultVariant := createDefaultVariant(product)
		defaultVariant.ProductID = product.ID

		// Ensure SKU is set for the default variant
		if defaultVariant.SKU == "" {
			// Generate a proper SKU if none provided
			// First try to get brand and category information
			var brandName string
			if product.BrandID != nil {
				brand, err := s.brandRepo.GetBrandByID(ctx, *product.BrandID)
				if err == nil && brand != nil {
					brandName = brand.Name
				}
			}

			// For categories, we can't get them yet since the product was just created
			// and categories are associated after creation
			var categoryName string
			// Try to extract category from request if available
			if len(req.Product.Categories) > 0 {
				categoryName = req.Product.Categories[0].Name
			}

			// Generate a unique SKU using our utility
			uniqueSKU, err := utils.GenerateUniqueSKU(ctx, s.productRepo, brandName, categoryName, "", "", 5)
			if err != nil {
				s.logger.Warn("Failed to generate unique SKU, falling back to basic SKU generation",
					zap.Error(err),
					zap.String("product_id", product.ID))
				// Fallback to basic SKU generation
				uniqueSKU = utils.GenerateSKU(brandName, categoryName, "", "")
			}

			defaultVariant.SKU = uniqueSKU

			// Also update the product's SKU
			product.SKU = defaultVariant.SKU

			s.logger.Info("Generated unique SKU for product",
				zap.String("product_id", product.ID),
				zap.String("sku", product.SKU))
		}

		if err := s.productRepo.CreateVariant(ctx, nil, product.ID, &defaultVariant); err != nil {
			s.logger.Error("Failed to create default variant", zap.Error(err))
			// Continue even if default variant creation fails
		}
	}

	// Handle variants if provided
	if len(req.Product.Variants) > 0 {
		for _, variantProto := range req.Product.Variants {
			variant := &models.ProductVariant{
				ProductID: product.ID,
				SKU:       variantProto.Sku,
				Price:     variantProto.Price,
			}

			if variantProto.Title != "" {
				variant.Title = &variantProto.Title
			}
			if variantProto.DiscountPrice != nil {
				discountPrice := variantProto.DiscountPrice.Value
				variant.DiscountPrice = &discountPrice
			}

			// Process variant attributes
			if len(variantProto.Attributes) > 0 {
				variant.Attributes = make([]models.VariantAttributeValue, len(variantProto.Attributes))
				for i, attr := range variantProto.Attributes {
					variant.Attributes[i] = models.VariantAttributeValue{
						Name:  attr.Name,
						Value: attr.Value,
					}
				}
			}

			// Process variant images
			if len(variantProto.Images) > 0 {
				variant.Images = make([]models.VariantImage, len(variantProto.Images))
				for i, img := range variantProto.Images {
					variant.Images[i] = models.VariantImage{
						VariantID: variant.ID, // This will be populated after variant creation
						URL:       img.Url,
						AltText:   img.AltText,
						Position:  int(img.Position),
					}
				}
			}

			// Inherit fields from parent product
			variant.InheritFromProduct(product)

			if err := s.productRepo.CreateVariant(ctx, nil, product.ID, variant); err != nil {
				s.logger.Error("Failed to create variant", zap.Error(err))
				// Continue with other variants even if one fails
			}
		}
	}

	// Process shipping data if provided
	if req.Product.Shipping != nil {
		shipping := &models.ProductShipping{
			ProductID:        product.ID,
			FreeShipping:     req.Product.Shipping.FreeShipping,
			EstimatedDays:    int(req.Product.Shipping.EstimatedDays),
			ExpressAvailable: req.Product.Shipping.ExpressAvailable,
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		}

		if err := s.productRepo.UpsertProductShipping(ctx, shipping); err != nil {
			s.logger.Error("Failed to save shipping information", zap.Error(err))
			// Continue even if shipping data fails to save
		}
	}

	// Process SEO data if provided
	if req.Product.Seo != nil {
		seo := &models.ProductSEO{
			ProductID:       product.ID,
			MetaTitle:       req.Product.Seo.MetaTitle,
			MetaDescription: req.Product.Seo.MetaDescription,
			Keywords:        req.Product.Seo.Keywords,
			Tags:            req.Product.Seo.Tags,
			CreatedAt:       time.Now().UTC(),
			UpdatedAt:       time.Now().UTC(),
		}

		if err := s.productRepo.UpsertProductSEO(ctx, seo); err != nil {
			s.logger.Error("Failed to save SEO information", zap.Error(err))
			// Continue even if SEO data fails to save
		}
	}

	// Process specifications if provided
	if len(req.Product.Specifications) > 0 {
		for _, specProto := range req.Product.Specifications {
			spec := &models.ProductSpecification{
				ProductID: product.ID,
				Name:      specProto.Name,
				Value:     specProto.Value,
				Unit:      specProto.Unit,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}

			if err := s.productRepo.AddProductSpecification(ctx, spec); err != nil {
				s.logger.Error("Failed to add product specification", zap.Error(err))
				// Continue with other specifications even if one fails
			}
		}
	}

	// Process tags if provided
	if len(req.Product.Tags) > 0 {
		for _, tagProto := range req.Product.Tags {
			tag := &models.ProductTag{
				ProductID: product.ID,
				Tag:       tagProto.Tag,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}

			if err := s.productRepo.AddProductTag(ctx, tag); err != nil {
				s.logger.Error("Failed to add product tag", zap.Error(err))
				// Continue with other tags even if one fails
			}
		}
	}

	// Process attributes if provided
	if len(req.Product.Attributes) > 0 {
		for _, attrProto := range req.Product.Attributes {
			attr := &models.ProductAttribute{
				ProductID: product.ID,
				Name:      attrProto.Name,
				Value:     attrProto.Value,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}

			if err := s.productRepo.AddProductAttribute(ctx, attr); err != nil {
				s.logger.Error("Failed to add product attribute", zap.Error(err))
				// Continue with other attributes even if one fails
			}
		}
	}

	// Process discount if provided
	if req.Product.Discount != nil {
		discountProto := req.Product.Discount
		var expiresAt *time.Time
		if discountProto.ExpiresAt != nil {
			expTime := discountProto.ExpiresAt.AsTime()
			expiresAt = &expTime
		}

		discount := &models.ProductDiscount{
			ProductID: product.ID,
			Type:      discountProto.Type,
			Value:     discountProto.Value,
			ExpiresAt: expiresAt,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		if err := s.productRepo.AddProductDiscount(ctx, discount); err != nil {
			s.logger.Error("Failed to add product discount", zap.Error(err))
			// Continue even if discount fails to save
		}
	}

	// Invalidate all product list caches to ensure new product appears in lists
	if err := s.cacheManager.InvalidateProductLists(ctx); err != nil {
		s.logger.Warn("Failed to invalidate product lists after creation",
			zap.String("product_id", product.ID),
			zap.Error(err))
		// Continue even if cache invalidation fails
	} else {
		s.logger.Info("Successfully invalidated product list caches", zap.String("product_id", product.ID))
	}

	// Return the created product
	return s.GetProduct(ctx, &pb.GetProductRequest{
		Identifier: &pb.GetProductRequest_Id{Id: product.ID},
	})
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

		// Get the total count from the database to ensure accurate pagination
		_, total, err := s.productRepo.List(ctx, 0, 1)
		if err != nil {
			s.logger.Error("Failed to get total product count", zap.Error(err))
			// Fall back to using the cached products length
			return &pb.ListProductsResponse{
				Products: convertProductModelsToProtos(products),
				Total:    int32(len(products)),
			}, nil
		}

		return &pb.ListProductsResponse{
			Products: convertProductModelsToProtos(products),
			Total:    int32(total),
		}, nil
	}

	// Cache miss, get from database
	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0 // Ensure offset is not negative
	}

	// Get basic product list
	products, total, err := s.productRepo.List(ctx, int(offset), int(req.Limit))
	if err != nil {
		s.logger.Error("Failed to list products", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list products")
	}

	// Log the total count for debugging
	s.logger.Info("Product count from database",
		zap.Int("total", total),
		zap.Int("page", int(req.Page)),
		zap.Int("limit", int(req.Limit)),
		zap.Int("offset", int(offset)),
		zap.Int("products_returned", len(products)))

	// Enhance each product with complete data
	enhancedProducts := make([]*models.Product, 0, len(products))
	for _, product := range products {
		// Get complete product data
		enhancedProduct, err := s.productRepo.GetByID(ctx, product.ID)
		if err != nil {
			s.logger.Error("Failed to get enhanced product data",
				zap.Error(err),
				zap.String("product_id", product.ID))
			// Continue with the basic product data
			enhancedProducts = append(enhancedProducts, product)
		} else {
			// Populate related entities
			if err := s.populateProductRelations(ctx, enhancedProduct); err != nil {
				s.logger.Error("Failed to populate product relations",
					zap.Error(err),
					zap.String("product_id", product.ID))
				// Continue with the basic product data
			}
			enhancedProducts = append(enhancedProducts, enhancedProduct)
		}
	}

	// Cache the enhanced result
	if err := s.cacheManager.SetProductList(ctx, cacheKey, enhancedProducts); err != nil {
		s.logger.Warn("Failed to cache product list", zap.Error(err))
	}

	// Log the number of products and their data completeness
	for i, product := range enhancedProducts {
		s.logger.Debug("Product in list",
			zap.Int("index", i),
			zap.String("id", product.ID),
			zap.Int("image_count", len(product.Images)),
			zap.Int("spec_count", len(product.Specifications)),
			zap.Int("tag_count", len(product.Tags)),
			zap.Bool("has_seo", product.SEO != nil),
			zap.Bool("has_shipping", product.Shipping != nil),
			zap.Float64("price", product.Price.Amount))
	}

	return &pb.ListProductsResponse{
		Products: convertProductModelsToProtos(enhancedProducts),
		Total:    int32(total),
	}, nil
}

// --- Conversion Helper Functions ---

func convertModelToProto(model *models.Product) *pb.Product {
	if model == nil {
		return nil
	}
	protoProduct := &pb.Product{
		Id:               model.ID,
		Title:            model.Title,
		Slug:             model.Slug,
		Description:      model.Description,
		ShortDescription: model.ShortDescription,
		Price:            model.Price.Amount,
		Sku:              model.SKU,
		IsPublished:      model.IsPublished,
		CreatedAt:        timestamppb.New(model.CreatedAt),
		UpdatedAt:        timestamppb.New(model.UpdatedAt),
		Brand:            convertBrandModelToProto(model.Brand),                    // Convert Brand
		Images:           convertImageModelsToProtos(model.Images),                 // Convert Images
		Categories:       convertCategorySliceToProtos(model.Categories),           // Convert Categories
		Variants:         convertVariantModelsToProtos(model.Variants),             // Convert Variants
		Tags:             convertTagModelsToProtos(model.Tags),                     // Convert Tags
		Attributes:       convertProductAttributeModelsToProtos(model.Attributes),  // Convert Attributes
		Specifications:   convertSpecificationModelsToProtos(model.Specifications), // Convert Specifications
		Seo:              convertSEOModelToProto(model.SEO),                        // Convert SEO
		Shipping:         convertShippingModelToProto(model.Shipping),              // Convert Shipping
		Discount:         convertDiscountModelToProto(model.Discount),              // Convert Discount
	}

	// Handle nullable fields
	if model.DiscountPrice != nil {
		protoProduct.DiscountPrice = wrapperspb.Double(model.DiscountPrice.Amount)
	}
	if model.Weight != nil {
		protoProduct.Weight = wrapperspb.Double(*model.Weight)
	}
	if model.BrandID != nil {
		protoProduct.BrandId = wrapperspb.String(*model.BrandID)
	}

	return protoProduct
}

func convertBrandModelToProto(model *models.Brand) *pb.Brand {
	if model == nil {
		return nil
	}
	proto := &pb.Brand{
		Id:          model.ID,
		Name:        model.Name,
		Slug:        model.Slug,
		Description: model.Description,
		CreatedAt:   timestamppb.New(model.CreatedAt),
		UpdatedAt:   timestamppb.New(model.UpdatedAt),
	}

	if model.DeletedAt != nil {
		proto.DeletedAt = timestamppb.New(*model.DeletedAt)
	}

	return proto
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
		Id:        model.ID,
		ProductId: model.ProductID,
		Sku:       model.SKU,
		Price:     model.Price,
		CreatedAt: timestamppb.New(model.CreatedAt),
		UpdatedAt: timestamppb.New(model.UpdatedAt),
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

	// Convert images
	if len(model.Images) > 0 {
		protoVariant.Images = make([]*pb.VariantImage, len(model.Images))
		for i, img := range model.Images {
			protoVariant.Images[i] = &pb.VariantImage{
				Id:        img.ID,
				VariantId: img.VariantID,
				Url:       img.URL,
				AltText:   img.AltText,
				Position:  int32(img.Position),
				CreatedAt: timestamppb.New(img.CreatedAt),
				UpdatedAt: timestamppb.New(img.UpdatedAt),
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

// UploadImage handles image upload to Cloudinary or local storage
func (s *ProductService) UploadImage(ctx context.Context, req *pb.UploadImageRequest) (*pb.UploadImageResponse, error) {
	// Set default folder if not provided
	folder := req.Folder
	if folder == "" {
		folder = "products"
	}

	// Try to initialize Cloudinary if not already initialized
	if s.cld == nil {
		// Get Cloudinary configuration from environment variables
		cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
		apiKey := os.Getenv("CLOUDINARY_API_KEY")
		apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

		if cloudName != "" && apiKey != "" && apiSecret != "" {
			cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
			if err == nil {
				s.cld = cld
			} else {
				s.logger.Warn("Failed to initialize Cloudinary, will use local storage", zap.Error(err))
			}
		} else {
			s.logger.Warn("Cloudinary configuration is missing, will use local storage")
		}
	}

	// Create a reader from the file bytes
	reader := bytes.NewReader(req.File)

	// Try to upload to Cloudinary if available
	if s.cld != nil {
		// Upload the file to Cloudinary
		result, err := s.cld.Upload.Upload(ctx, reader, uploader.UploadParams{
			Folder:   folder,
			PublicID: req.Filename,
		})
		if err == nil {
			s.logger.Info("Image uploaded to Cloudinary successfully",
				zap.String("public_id", result.PublicID),
				zap.String("url", result.SecureURL),
				zap.String("alt_text", req.AltText),
				zap.Int32("position", req.Position),
			)

			return &pb.UploadImageResponse{
				Url:      result.SecureURL,
				PublicId: result.PublicID,
				AltText:  req.AltText,
				Position: req.Position,
			}, nil
		}
		s.logger.Warn("Failed to upload to Cloudinary, falling back to local storage", zap.Error(err))
	}

	// Fallback to local storage
	// Reset reader position
	reader.Seek(0, 0)

	// Initialize local storage
	localStoragePath := os.Getenv("LOCAL_STORAGE_PATH")
	if localStoragePath == "" {
		localStoragePath = "./uploads" // Default path
	}

	localStorage, err := storage.NewLocalStorage(localStoragePath)
	if err != nil {
		s.logger.Error("Failed to initialize local storage", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to initialize storage")
	}

	// Upload to local storage
	result, err := localStorage.SaveFromReader(reader, folder, req.Filename)
	if err != nil {
		s.logger.Error("Failed to save image to local storage", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to save image")
	}

	// Generate a full URL for the image
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default base URL
	}

	imageURL := baseURL + result.URL

	s.logger.Info("Image saved to local storage successfully",
		zap.String("public_id", result.PublicID),
		zap.String("url", imageURL),
		zap.String("alt_text", req.AltText),
		zap.Int32("position", req.Position),
	)

	return &pb.UploadImageResponse{
		Url:      imageURL,
		PublicId: result.PublicID,
		AltText:  req.AltText,
		Position: req.Position,
	}, nil
}

// DeleteImage handles image deletion from Cloudinary
func (s *ProductService) DeleteImage(ctx context.Context, req *pb.DeleteImageRequest) (*pb.DeleteImageResponse, error) {
	// Initialize Cloudinary if not already initialized
	if s.cld == nil {
		// Get Cloudinary configuration from environment variables
		cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
		apiKey := os.Getenv("CLOUDINARY_API_KEY")
		apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

		if cloudName == "" || apiKey == "" || apiSecret == "" {
			return nil, status.Error(codes.FailedPrecondition, "Cloudinary configuration is missing")
		}

		cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
		if err != nil {
			s.logger.Error("Failed to initialize Cloudinary", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to initialize image upload service")
		}
		s.cld = cld
	}

	// Delete the image from Cloudinary
	result, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: req.PublicId,
	})
	if err != nil {
		s.logger.Error("Failed to delete image from Cloudinary", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to delete image")
	}

	s.logger.Info("Image deleted successfully",
		zap.String("public_id", req.PublicId),
		zap.String("result", result.Result),
	)

	return &pb.DeleteImageResponse{
		Success: result.Result == "ok",
	}, nil
}

// GenerateSKUPreview generates a preview of a SKU based on the provided parameters
// This is used by the admin UI to show a preview of the SKU that would be generated
func (s *ProductService) GenerateSKUPreview(ctx context.Context, req *pb.GenerateSKUPreviewRequest) (*pb.GenerateSKUPreviewResponse, error) {
	s.logger.Info("Generating SKU preview",
		zap.String("brand", req.BrandName),
		zap.String("category", req.CategoryName),
		zap.String("color", req.Color),
		zap.String("size", req.Size))

	// Generate a SKU using the provided parameters
	sku := utils.GenerateSKU(req.BrandName, req.CategoryName, req.Color, req.Size)

	// Check if the SKU already exists
	exists, err := s.productRepo.IsSKUExists(ctx, sku)
	if err != nil {
		s.logger.Error("Failed to check if SKU exists", zap.Error(err))
		// Continue with the generated SKU even if the check fails
	}

	// If the SKU exists, add a note to the response
	if exists {
		// Generate a new unique SKU
		uniqueSKU, err := utils.GenerateUniqueSKU(ctx, s.productRepo, req.BrandName, req.CategoryName, req.Color, req.Size, 5)
		if err != nil {
			s.logger.Warn("Failed to generate unique SKU, using original SKU", zap.Error(err))
			// Return the original SKU with a note that it's not unique
			return &pb.GenerateSKUPreviewResponse{
				Sku: sku + " (not unique)",
			}, nil
		}

		// Return the unique SKU
		return &pb.GenerateSKUPreviewResponse{
			Sku: uniqueSKU,
		}, nil
	}

	// Return the generated SKU
	return &pb.GenerateSKUPreviewResponse{
		Sku: sku,
	}, nil
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
		ID:        proto.Id,
		ProductID: proto.ProductId,
		SKU:       proto.Sku,
		Price:     proto.Price,
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

	// Invalidate brand cache
	if err := s.cacheManager.InvalidateBrandLists(ctx); err != nil {
		s.logger.Warn("Failed to invalidate brand cache", zap.Error(err))
		// Continue even if cache invalidation fails
	}

	// Convert model back to proto using the helper function
	return convertBrandModelToProto(brandModel), nil
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
	s.logger.Info("ListBrands service method called", zap.Int32("page", req.Page), zap.Int32("limit", req.Limit))

	// Set default pagination values if not provided
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Define cache key based on pagination
	cacheKey := fmt.Sprintf("brands:page:%d:limit:%d", req.Page, req.Limit)

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

	// Calculate offset from page and limit
	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0 // Ensure offset is not negative
	}

	// Cache miss, get from database with pagination
	brands, total, err := s.brandRepo.ListBrands(ctx, int(offset), int(req.Limit))
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

	// Fetch the complete category with parent name to ensure it's properly populated
	if category.ParentID != nil {
		completeCategory, err := s.categoryRepo.GetCategoryByID(ctx, category.ID)
		if err == nil {
			category = completeCategory
		} else {
			s.logger.Warn("Failed to fetch complete category after creation", zap.Error(err))
		}
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

	if model.DeletedAt != nil {
		protoCategory.DeletedAt = timestamppb.New(*model.DeletedAt)
	}

	return protoCategory
}

// populateProductRelations populates related entities for a product
func (s *ProductService) populateProductRelations(ctx context.Context, product *models.Product) error {
	if product == nil {
		return fmt.Errorf("cannot populate relations for nil product")
	}

	// Get brand if brandID is set but brand is nil
	if product.BrandID != nil && product.Brand == nil {
		brand, err := s.brandRepo.GetBrandByID(ctx, *product.BrandID)
		if err != nil {
			s.logger.Error("Failed to get brand for product", zap.Error(err), zap.String("brand_id", *product.BrandID))
			// Continue even if brand fails to load
		} else {
			product.Brand = brand
		}
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

		// Get variant images
		variantImages, err := s.productRepo.GetVariantImages(ctx, v.ID)
		if err != nil {
			s.logger.Error("Failed to get variant images", zap.Error(err), zap.String("variant_id", v.ID))
			// Continue even if variant images fail to load
		} else {
			product.Variants[i].Images = variantImages
		}

		// Get variant attributes
		variantAttributes, err := s.productRepo.GetVariantAttributes(ctx, v.ID)
		if err != nil {
			s.logger.Error("Failed to get variant attributes", zap.Error(err), zap.String("variant_id", v.ID))
			// Continue even if variant attributes fail to load
		} else {
			product.Variants[i].Attributes = variantAttributes
		}

		// Inherit fields from parent product
		product.Variants[i].InheritFromProduct(product)
	}

	// Use the first variant's data for backward compatibility
	if len(variants) > 0 {
		defaultVariant := variants[0] // Use first variant

		// Copy default variant's values to product's transient fields
		product.Price = models.Price{
			Amount:   defaultVariant.Price,
			Currency: "USD", // Default currency
		}
		if defaultVariant.DiscountPrice != nil {
			product.DiscountPrice = &models.Price{
				Amount:   *defaultVariant.DiscountPrice,
				Currency: "USD", // Default currency
			}
		} else {
			product.DiscountPrice = nil
		}
		product.SKU = defaultVariant.SKU
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

	// Get product images
	images, err := s.productRepo.GetProductImages(ctx, product.ID)
	if err != nil {
		s.logger.Error("Failed to get product images", zap.Error(err), zap.String("product_id", product.ID))
		// Continue even if product images fail to load
	} else {
		product.Images = images
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
	model.Price = models.Price{Amount: proto.Price, Currency: "USD"}

	if proto.Sku != "" {
		model.SKU = proto.Sku // Corrected field name
	}

	// IsPublished is not nullable in proto
	model.IsPublished = proto.IsPublished

	// Handle nullable fields from proto
	if proto.DiscountPrice != nil {
		model.DiscountPrice = &models.Price{Amount: proto.DiscountPrice.Value, Currency: "USD"}
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

	// Handle categories if provided in the update
	if len(proto.Categories) > 0 {
		// Replace existing categories with the new ones
		model.Categories = make([]models.Category, len(proto.Categories))
		for i, cat := range proto.Categories {
			// We don't need to log here as this is a helper function

			model.Categories[i] = models.Category{
				ID: cat.Id,
			}

			// If we have more category details, add them
			if cat.Name != "" {
				model.Categories[i].Name = cat.Name
			}
			if cat.Slug != "" {
				model.Categories[i].Slug = cat.Slug
			}
			if cat.Description != "" {
				model.Categories[i].Description = cat.Description
			}
		}
	}

	return model
}

func createDefaultVariant(product *models.Product) models.ProductVariant {
	now := time.Now().UTC()

	var discountPrice *float64
	if product.DiscountPrice != nil {
		discountPrice = &product.DiscountPrice.Amount
	}

	// Ensure we have a valid price
	price := product.Price.Amount
	if price <= 0 {
		price = 9.99 // Default price if none provided
	}

	// Ensure we have a valid SKU
	sku := product.SKU
	if sku == "" {
		// Just use a simple SKU format here
		// The main unique SKU generation happens in the CreateProduct method
		// where we have more context and can check for uniqueness
		sku = utils.GenerateSKU(
			"", // No brand info in this context
			"", // No category info in this context
			"", // No color info
			"", // No size info
		)
	}

	return models.ProductVariant{
		Title:         &product.Title,
		SKU:           sku,
		Price:         price,
		DiscountPrice: discountPrice,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}
