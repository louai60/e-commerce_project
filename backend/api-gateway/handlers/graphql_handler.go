package handlers

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"go.uber.org/zap"

	"github.com/louai60/e-commerce_project/backend/api-gateway/clients"
	inventorypb "github.com/louai60/e-commerce_project/backend/inventory-service/proto"
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
)

// GraphQLHandler handles GraphQL requests
type GraphQLHandler struct {
	schema          *graphql.Schema
	handler         *handler.Handler
	logger          *zap.Logger
	inventoryClient *clients.InventoryClient
	productClient   pb.ProductServiceClient
}

// InventoryItemWithProduct extends the inventory item with product information
type InventoryItemWithProduct struct {
	*inventorypb.InventoryItem
	Product *clients.ProductInfo
}

// NewGraphQLHandler creates a new GraphQL handler
func NewGraphQLHandler(
	logger *zap.Logger,
	inventoryClient *clients.InventoryClient,
	productClient pb.ProductServiceClient,
) (*GraphQLHandler, error) {
	// Load schema from file - we don't actually use the content directly
	_, err := loadSchemaFromFile("schema/schema.graphql")
	if err != nil {
		return nil, err
	}

	// Create root query
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			// Inventory Items Query
			"inventoryItems": &graphql.Field{
				Type: inventoryItemsResponseType,
				Args: graphql.FieldConfigArgument{
					"page": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"limit": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"lowStockOnly": &graphql.ArgumentConfig{
						Type: graphql.Boolean,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					page, _ := p.Args["page"].(int)
					limit, _ := p.Args["limit"].(int)
					lowStockOnly, _ := p.Args["lowStockOnly"].(bool)

					if page <= 0 {
						page = 1
					}
					if limit <= 0 {
						limit = 10
					}

					// Convert to filters map
					filters := make(map[string]string)
					if lowStockOnly {
						filters["low_stock_only"] = "true"
					}

					// Call inventory client
					items, total, err := inventoryClient.ListInventoryItems(
						context.Background(),
						page,
						limit,
						"", // status
						"", // warehouseID
						lowStockOnly,
					)
					if err != nil {
						logger.Error("Failed to get inventory items", zap.Error(err))
						return nil, err
					}

					// Create enhanced inventory items with product info
					enhancedItems := make([]*InventoryItemWithProduct, len(items))
					for i, item := range items {
						enhancedItems[i] = &InventoryItemWithProduct{
							InventoryItem: item,
							Product:       nil,
						}

						// Fetch product details if needed
						if item.ProductId != "" {
							logger.Info("Fetching product details for inventory item",
								zap.String("inventory_item_id", item.Id),
								zap.String("product_id", item.ProductId))

							req := &pb.GetProductRequest{
								Identifier: &pb.GetProductRequest_Id{
									Id: item.ProductId,
								},
							}
							product, err := productClient.GetProduct(context.Background(), req)
							if err != nil {
								logger.Error("Failed to get product for inventory item",
									zap.String("inventory_item_id", item.Id),
									zap.String("product_id", item.ProductId),
									zap.Error(err))
							} else if product == nil {
								logger.Error("Product is nil for inventory item",
									zap.String("inventory_item_id", item.Id),
									zap.String("product_id", item.ProductId))
							} else {
								logger.Info("Successfully fetched product for inventory item",
									zap.String("inventory_item_id", item.Id),
									zap.String("product_id", item.ProductId),
									zap.String("product_title", product.Title))

								// Map product to inventory item
								productInfo := &clients.ProductInfo{
									Id:    product.Id,
									Title: product.Title,
									Slug:  product.Slug,
								}

								// Map images if available
								if len(product.Images) > 0 {
									productInfo.Images = make([]clients.ImageInfo, len(product.Images))
									for j, img := range product.Images {
										productInfo.Images[j] = clients.ImageInfo{
											Url: img.Url,
										}
									}
								}

								enhancedItems[i].Product = productInfo
							}
						}
					}

					// Create response structure
					resp := map[string]interface{}{
						"items": enhancedItems,
						"pagination": map[string]interface{}{
							"current_page": page,
							"total_pages":  (total + limit - 1) / limit,
							"per_page":     limit,
							"total_items":  total,
						},
					}

					return resp, nil
				},
			},
			// Single Inventory Item Query
			"inventoryItem": &graphql.Field{
				Type: inventoryItemType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, _ := p.Args["id"].(string)

					// Call inventory client
					item, err := inventoryClient.GetInventoryItem(context.Background(), id)
					if err != nil {
						logger.Error("Failed to get inventory item", zap.Error(err))
						return nil, err
					}

					// Create enhanced inventory item with product info
					enhancedItem := &InventoryItemWithProduct{
						InventoryItem: item,
						Product:       nil,
					}

					// Fetch product details if needed
					if item.ProductId != "" {
						req := &pb.GetProductRequest{
							Identifier: &pb.GetProductRequest_Id{
								Id: item.ProductId,
							},
						}
						product, err := productClient.GetProduct(context.Background(), req)
						if err == nil && product != nil {
							// Map product to inventory item
							productInfo := &clients.ProductInfo{
								Id:    product.Id,
								Title: product.Title,
								Slug:  product.Slug,
							}

							// Map images if available
							if len(product.Images) > 0 {
								productInfo.Images = make([]clients.ImageInfo, len(product.Images))
								for j, img := range product.Images {
									productInfo.Images[j] = clients.ImageInfo{
										Url: img.Url,
									}
								}
							}

							enhancedItem.Product = productInfo
						}
					}

					return enhancedItem, nil
				},
			},
			// Warehouses Query
			"warehouses": &graphql.Field{
				Type: warehousesResponseType,
				Args: graphql.FieldConfigArgument{
					"page": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"limit": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					page, _ := p.Args["page"].(int)
					limit, _ := p.Args["limit"].(int)

					if page <= 0 {
						page = 1
					}
					if limit <= 0 {
						limit = 10
					}

					// Call inventory client
					warehouses, total, err := inventoryClient.ListWarehouses(context.Background(), page, limit, nil)
					if err != nil {
						logger.Error("Failed to get warehouses", zap.Error(err))
						return nil, err
					}

					// Create response structure
					resp := map[string]interface{}{
						"warehouses": warehouses,
						"pagination": map[string]interface{}{
							"current_page": page,
							"total_pages":  (total + limit - 1) / limit,
							"per_page":     limit,
							"total_items":  total,
						},
					}

					return resp, nil
				},
			},
			// Single Warehouse Query - not implemented in the client yet
			"warehouse": &graphql.Field{
				Type: warehouseType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, _ := p.Args["id"].(string)

					// We don't have a direct GetWarehouse method, so we'll list warehouses and filter
					warehouses, _, err := inventoryClient.ListWarehouses(context.Background(), 1, 100, nil)
					if err != nil {
						logger.Error("Failed to list warehouses", zap.Error(err))
						return nil, err
					}

					// Find the warehouse with the matching ID
					for _, warehouse := range warehouses {
						if warehouse.Id == id {
							return warehouse, nil
						}
					}

					logger.Error("Warehouse not found", zap.String("id", id))
					return nil, fmt.Errorf("warehouse not found")
				},
			},
			// Inventory Transactions Query
			"inventoryTransactions": &graphql.Field{
				Type: inventoryTransactionsResponseType,
				Args: graphql.FieldConfigArgument{
					"page": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"limit": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"transactionType": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"warehouseId": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"dateFrom": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"dateTo": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					page, _ := p.Args["page"].(int)
					limit, _ := p.Args["limit"].(int)
					transactionType, _ := p.Args["transactionType"].(string)
					warehouseId, _ := p.Args["warehouseId"].(string)
					dateFrom, _ := p.Args["dateFrom"].(string)
					dateTo, _ := p.Args["dateTo"].(string)

					if page <= 0 {
						page = 1
					}
					if limit <= 0 {
						limit = 10
					}

					// Call inventory client
					transactions, total, err := inventoryClient.ListInventoryTransactions(
						context.Background(),
						page,
						limit,
						transactionType,
						warehouseId,
						dateFrom,
						dateTo,
					)
					if err != nil {
						logger.Error("Failed to get inventory transactions", zap.Error(err))
						return nil, err
					}

					// Create response structure
					resp := map[string]interface{}{
						"transactions": transactions,
						"pagination": map[string]interface{}{
							"current_page": page,
							"total_pages":  (total + limit - 1) / limit,
							"per_page":     limit,
							"total_items":  total,
						},
					}

					return resp, nil
				},
			},
			// Product Query
			"product": &graphql.Field{
				Type: productType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, _ := p.Args["id"].(string)

					// Create the request
					req := &pb.GetProductRequest{
						Identifier: &pb.GetProductRequest_Id{
							Id: id,
						},
					}

					// Call product client
					product, err := productClient.GetProduct(context.Background(), req)
					if err != nil {
						logger.Error("Failed to get product", zap.Error(err))
						return nil, err
					}

					return product, nil
				},
			},
			// Products Query
			"products": &graphql.Field{
				Type: productsResponseType,
				Args: graphql.FieldConfigArgument{
					"page": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"limit": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					page, _ := p.Args["page"].(int)
					limit, _ := p.Args["limit"].(int)

					if page <= 0 {
						page = 1
					}
					if limit <= 0 {
						limit = 10
					}

					// Create the request
					req := &pb.ListProductsRequest{
						Page:  int32(page),
						Limit: int32(limit),
					}

					// Call product client
					resp, err := productClient.ListProducts(context.Background(), req)
					if err != nil {
						logger.Error("Failed to get products", zap.Error(err))
						return nil, err
					}

					// Create response structure
					result := map[string]interface{}{
						"products": resp.Products,
						"pagination": map[string]interface{}{
							"current_page": page,
							"total_pages":  (int(resp.Total) + limit - 1) / limit,
							"per_page":     limit,
							"total_items":  int(resp.Total),
						},
					}

					return result, nil
				},
			},
		},
	})

	// Create schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: rootQuery,
	})
	if err != nil {
		return nil, err
	}

	// Create handler
	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	return &GraphQLHandler{
		schema:          &schema,
		handler:         h,
		logger:          logger,
		inventoryClient: inventoryClient,
		productClient:   productClient,
	}, nil
}

// Handle handles GraphQL requests
func (h *GraphQLHandler) Handle(c *gin.Context) {
	h.handler.ServeHTTP(c.Writer, c.Request)
}

// Helper function to load schema from file
func loadSchemaFromFile(filePath string) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	file, err := os.Open(absPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// GraphQL types
var imageType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Image",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
		},
		"url": &graphql.Field{
			Type: graphql.String,
		},
		"alt_text": &graphql.Field{
			Type: graphql.String,
		},
		"position": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

var productType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Product",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
		},
		"title": &graphql.Field{
			Type: graphql.String,
		},
		"slug": &graphql.Field{
			Type: graphql.String,
		},
		"description": &graphql.Field{
			Type: graphql.String,
		},
		"price": &graphql.Field{
			Type: graphql.Float,
		},
		"discount_price": &graphql.Field{
			Type: graphql.Float,
		},
		"status": &graphql.Field{
			Type: graphql.String,
		},
		"images": &graphql.Field{
			Type: graphql.NewList(imageType),
		},
		"created_at": &graphql.Field{
			Type: graphql.String,
		},
		"updated_at": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var paginationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Pagination",
	Fields: graphql.Fields{
		"current_page": &graphql.Field{
			Type: graphql.Int,
		},
		"total_pages": &graphql.Field{
			Type: graphql.Int,
		},
		"per_page": &graphql.Field{
			Type: graphql.Int,
		},
		"total_items": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

var productsResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ProductsResponse",
	Fields: graphql.Fields{
		"products": &graphql.Field{
			Type: graphql.NewList(productType),
		},
		"pagination": &graphql.Field{
			Type: paginationType,
		},
	},
})

var warehouseType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Warehouse",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"code": &graphql.Field{
			Type: graphql.String,
		},
		"address": &graphql.Field{
			Type: graphql.String,
		},
		"city": &graphql.Field{
			Type: graphql.String,
		},
		"state": &graphql.Field{
			Type: graphql.String,
		},
		"country": &graphql.Field{
			Type: graphql.String,
		},
		"postal_code": &graphql.Field{
			Type: graphql.String,
		},
		"is_active": &graphql.Field{
			Type: graphql.Boolean,
		},
		"priority": &graphql.Field{
			Type: graphql.Int,
		},
		"item_count": &graphql.Field{
			Type: graphql.Int,
		},
		"total_quantity": &graphql.Field{
			Type: graphql.Int,
		},
		"created_at": &graphql.Field{
			Type: graphql.String,
		},
		"updated_at": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var inventoryLocationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "InventoryLocation",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
		},
		"inventory_item_id": &graphql.Field{
			Type: graphql.String,
		},
		"warehouse_id": &graphql.Field{
			Type: graphql.String,
		},
		"quantity": &graphql.Field{
			Type: graphql.Int,
		},
		"available_quantity": &graphql.Field{
			Type: graphql.Int,
		},
		"reserved_quantity": &graphql.Field{
			Type: graphql.Int,
		},
		"created_at": &graphql.Field{
			Type: graphql.String,
		},
		"updated_at": &graphql.Field{
			Type: graphql.String,
		},
		"warehouse": &graphql.Field{
			Type: warehouseType,
		},
	},
})

// Define a product info type for the embedded product data
var productInfoType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ProductInfo",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
		},
		"title": &graphql.Field{
			Type: graphql.String,
		},
		"slug": &graphql.Field{
			Type: graphql.String,
		},
		"images": &graphql.Field{
			Type: graphql.NewList(imageType),
		},
	},
})

var inventoryItemType = graphql.NewObject(graphql.ObjectConfig{
	Name: "InventoryItem",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok && item.InventoryItem != nil {
					return item.InventoryItem.Id, nil
				}
				return nil, nil
			},
		},
		"product_id": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok && item.InventoryItem != nil {
					return item.InventoryItem.ProductId, nil
				}
				return nil, nil
			},
		},
		"variant_id": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok && item.InventoryItem != nil && item.InventoryItem.VariantId != nil {
					return item.InventoryItem.VariantId.Value, nil
				}
				return nil, nil
			},
		},
		"sku": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok && item.InventoryItem != nil {
					return item.InventoryItem.Sku, nil
				}
				return nil, nil
			},
		},
		"total_quantity": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok && item.InventoryItem != nil {
					return int(item.InventoryItem.TotalQuantity), nil
				}
				return nil, nil
			},
		},
		"available_quantity": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok && item.InventoryItem != nil {
					return int(item.InventoryItem.AvailableQuantity), nil
				}
				return nil, nil
			},
		},
		"reserved_quantity": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok && item.InventoryItem != nil {
					return int(item.InventoryItem.ReservedQuantity), nil
				}
				return nil, nil
			},
		},
		"reorder_point": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok && item.InventoryItem != nil {
					return int(item.InventoryItem.ReorderPoint), nil
				}
				return nil, nil
			},
		},
		"reorder_quantity": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok && item.InventoryItem != nil {
					return int(item.InventoryItem.ReorderQuantity), nil
				}
				return nil, nil
			},
		},
		"status": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok && item.InventoryItem != nil {
					return item.InventoryItem.Status, nil
				}
				return nil, nil
			},
		},
		"last_updated": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok && item.InventoryItem != nil && item.InventoryItem.LastUpdated != nil {
					return item.InventoryItem.LastUpdated.AsTime().Format(time.RFC3339), nil
				}
				return nil, nil
			},
		},
		"created_at": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok && item.InventoryItem != nil && item.InventoryItem.CreatedAt != nil {
					return item.InventoryItem.CreatedAt.AsTime().Format(time.RFC3339), nil
				}
				return nil, nil
			},
		},
		"updated_at": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok && item.InventoryItem != nil && item.InventoryItem.UpdatedAt != nil {
					return item.InventoryItem.UpdatedAt.AsTime().Format(time.RFC3339), nil
				}
				return nil, nil
			},
		},
		"locations": &graphql.Field{
			Type: graphql.NewList(inventoryLocationType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok && item.InventoryItem != nil {
					return item.InventoryItem.Locations, nil
				}
				return nil, nil
			},
		},
		"product": &graphql.Field{
			Type: productInfoType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if item, ok := p.Source.(*InventoryItemWithProduct); ok {
					return item.Product, nil
				}
				return nil, nil
			},
		},
	},
})

var inventoryItemsResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name: "InventoryItemsResponse",
	Fields: graphql.Fields{
		"items": &graphql.Field{
			Type: graphql.NewList(inventoryItemType),
		},
		"pagination": &graphql.Field{
			Type: paginationType,
		},
	},
})

var warehousesResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name: "WarehousesResponse",
	Fields: graphql.Fields{
		"warehouses": &graphql.Field{
			Type: graphql.NewList(warehouseType),
		},
		"pagination": &graphql.Field{
			Type: paginationType,
		},
	},
})

var inventoryTransactionType = graphql.NewObject(graphql.ObjectConfig{
	Name: "InventoryTransaction",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
		},
		"inventory_item_id": &graphql.Field{
			Type: graphql.String,
		},
		"transaction_type": &graphql.Field{
			Type: graphql.String,
		},
		"quantity": &graphql.Field{
			Type: graphql.Int,
		},
		"warehouse_id": &graphql.Field{
			Type: graphql.String,
		},
		"reference_id": &graphql.Field{
			Type: graphql.String,
		},
		"reference_type": &graphql.Field{
			Type: graphql.String,
		},
		"notes": &graphql.Field{
			Type: graphql.String,
		},
		"created_by": &graphql.Field{
			Type: graphql.String,
		},
		"created_at": &graphql.Field{
			Type: graphql.String,
		},
		"inventory_item": &graphql.Field{
			Type: inventoryItemType,
		},
		"warehouse": &graphql.Field{
			Type: warehouseType,
		},
	},
})

var inventoryTransactionsResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name: "InventoryTransactionsResponse",
	Fields: graphql.Fields{
		"transactions": &graphql.Field{
			Type: graphql.NewList(inventoryTransactionType),
		},
		"pagination": &graphql.Field{
			Type: paginationType,
		},
	},
})
