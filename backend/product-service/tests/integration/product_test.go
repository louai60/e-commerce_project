package integration

// import (
//     "testing" 

//     "github.com/stretchr/testify/suite"
//     "google.golang.org/grpc/codes"
//     "google.golang.org/grpc/status"
//     pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
//     "github.com/louai60/e-commerce_project/backend/product-service/tests/helper"
// )

// type ProductServiceTestSuite struct {
//     helper.TestSuite
// }

// func TestProductService(t *testing.T) {
//     suite.Run(t, new(ProductServiceTestSuite))
// }

// func (s *ProductServiceTestSuite) TestProductCRUD() {
//     s.Run("Create Product", func() {
//         product := s.CreateTestProduct()
//         s.NotEmpty(product.Id)
//         s.Equal("Test Product", product.Title)
//     })

//     s.Run("Get Product - Success", func() {
//         // Create a product specifically for this test
//         created := s.CreateTestProduct()
//         s.NotEmpty(created.Id, "Failed to create product for Get test")

//         // Test getting the product
//         got, err := s.Client.GetProduct(s.Ctx, &pb.GetProductRequest{
//             Identifier: &pb.ProductIdentifier{
//                 Id: created.Id,
//             },
//         })
//         s.NoError(err, "GetProduct failed unexpectedly")
//         s.Require().NotNil(got, "Received nil product")
//         s.Equal(created.Id, got.Id)
//         s.Equal(created.Title, got.Title)
//         s.Equal(created.Price, got.Price)
//         s.Equal(created.Sku, got.Sku)
//         s.Equal(created.Description, got.Description)
//     })

//     s.Run("Get Product - Not Found", func() {
//          _, err := s.Client.GetProduct(s.Ctx, &pb.GetProductRequest{
//             Identifier: &pb.ProductIdentifier{
//                 Id: "non-existent-id", // Use an ID that is unlikely to exist
//             },
//         })
//         s.Error(err, "Expected an error for non-existent product")
//         st, ok := status.FromError(err)
//         s.True(ok, "Error should be a gRPC status error")
//         s.Equal(codes.NotFound, st.Code(), "Expected NotFound error code")
//     })


//     s.Run("List Products - Basic", func() {
//         // Create a known number of products for this list test
//         p1 := s.CreateTestProduct()
//         p2 := s.CreateTestProduct()
//         s.NotEmpty(p1.Id)
//         s.NotEmpty(p2.Id)

//         limit := int32(10)
//         response, err := s.Client.ListProducts(s.Ctx, &pb.ListProductsRequest{
//             Page:  1,
//             Limit: limit,
//         })
//         s.NoError(err, "ListProducts failed unexpectedly")
//         s.Require().NotNil(response, "Received nil response")
//         s.NotEmpty(response.Products, "Expected products in the list")
//         s.LessOrEqual(int32(len(response.Products)), limit, "Returned more products than limit")
//         s.GreaterOrEqual(response.Total, int32(2), "Total should reflect at least the created products") // Use >= 2 as other tests might add products

//         // Basic check if one of the created products is in the list
//         foundP1 := false
//         for _, p := range response.Products {
//             if p.Id == p1.Id {
//                 foundP1 = true
//                 break
//             }
//         }
//         s.True(foundP1, "Created product P1 not found in the list")
//     })

//      s.Run("List Products - Pagination", func() {
//         // Create enough products to test pagination
//         for i := 0; i < 3; i++ {
//             s.CreateTestProduct()
//         }

//         // Get first page
//         limit := int32(2)
//         page1Resp, err := s.Client.ListProducts(s.Ctx, &pb.ListProductsRequest{Page: 1, Limit: limit})
//         s.NoError(err)
//         s.Require().NotNil(page1Resp)
//         s.Len(page1Resp.Products, int(limit), "Page 1 should have 'limit' products")
//         s.GreaterOrEqual(page1Resp.Total, int32(3), "Total should be at least 3")

//         // Get second page
//         page2Resp, err := s.Client.ListProducts(s.Ctx, &pb.ListProductsRequest{Page: 2, Limit: limit})
//         s.NoError(err)
//         s.Require().NotNil(page2Resp)
//         s.NotEmpty(page2Resp.Products, "Page 2 should have products")
//         s.LessOrEqual(int32(len(page2Resp.Products)), limit, "Page 2 should have at most 'limit' products")
//         s.Equal(page1Resp.Total, page2Resp.Total, "Total should be consistent across pages")

//         // Ensure products on page 1 and page 2 are different
//         if len(page1Resp.Products) > 0 && len(page2Resp.Products) > 0 {
//              s.NotEqual(page1Resp.Products[0].Id, page2Resp.Products[0].Id, "First product on Page 1 and Page 2 should be different")
//         }
//     })

//     s.Run("Update Product - Success", func() {
//         // Create a product specifically for this test
//         created := s.CreateTestProduct()
//         s.NotEmpty(created.Id, "Failed to create product for Update test")

//         // Update the product with more fields
//         newTitle := "Updated Test Product Title"
//         newPrice := 123.45
//         newDesc := "Updated description."
//         // Note: SKU update might not be allowed/intended, depends on service logic. Assuming it's allowed here.
//         newSku := created.Sku + "-UPDATED"

//         updated, err := s.Client.UpdateProduct(s.Ctx, &pb.UpdateProductRequest{
//             Product: &pb.Product{
//                 Id:          created.Id,
//                 Title:       newTitle,
//                 Price:       newPrice,
//                 Sku:         newSku,
//                 Description: newDesc,
//                 // Assuming BrandId and CategoryId are not updated here,
//                 // or require separate tests if update logic is complex.
//             },
//         })
//         s.NoError(err, "UpdateProduct failed unexpectedly")
//         s.Require().NotNil(updated, "Received nil product after update")
//         s.Equal(created.Id, updated.Id, "Product ID should not change on update")
//         s.Equal(newTitle, updated.Title)
//         s.Equal(newPrice, updated.Price)
//         s.Equal(newSku, updated.Sku)
//         s.Equal(newDesc, updated.Description)

//         // Verify by getting the product again
//         got, err := s.Client.GetProduct(s.Ctx, &pb.GetProductRequest{Identifier: &pb.ProductIdentifier{Id: created.Id}})
//         s.NoError(err)
//         s.Require().NotNil(got)
//         s.Equal(newTitle, got.Title)
//         s.Equal(newPrice, got.Price)
//         s.Equal(newSku, got.Sku)
//         s.Equal(newDesc, got.Description)
//     })

//      s.Run("Update Product - Not Found", func() {
//         _, err := s.Client.UpdateProduct(s.Ctx, &pb.UpdateProductRequest{
//             Product: &pb.Product{
//                 Id:    "non-existent-id",
//                 Title: "Trying to update non-existent",
//             },
//         })
//         s.Error(err, "Expected an error when updating non-existent product")
//         st, ok := status.FromError(err)
//         s.True(ok, "Error should be a gRPC status error")
//         s.Equal(codes.NotFound, st.Code(), "Expected NotFound error code")
//     })


//     s.Run("Delete Product - Success", func() {
//         // Create a product specifically for this test
//         created := s.CreateTestProduct()
//         s.NotEmpty(created.Id, "Failed to create product for Delete test")

//         // Delete the product
//         response, err := s.Client.DeleteProduct(s.Ctx, &pb.DeleteProductRequest{
//             Id: created.Id,
//         })
//         s.NoError(err, "DeleteProduct failed unexpectedly")
//         s.Require().NotNil(response)
//         s.True(response.Success, "Expected success confirmation on delete")

//         // Verify deletion by trying to get it
//         _, err = s.Client.GetProduct(s.Ctx, &pb.GetProductRequest{
//             Identifier: &pb.ProductIdentifier{
//                 Id: created.Id,
//             },
//         })
//         s.Error(err, "Expected an error when getting a deleted product")
//         st, ok := status.FromError(err)
//         s.True(ok, "Error should be a gRPC status error")
//         s.Equal(codes.NotFound, st.Code(), "Expected NotFound after deletion")
//     })

//     s.Run("Delete Product - Not Found", func() {
//         _, err := s.Client.DeleteProduct(s.Ctx, &pb.DeleteProductRequest{
//             Id: "non-existent-id",
//         })
//         // Depending on service implementation, this might return success (idempotent) or NotFound.
//         // Assuming NotFound is the expected behavior for trying to delete something not there.
//         s.Error(err, "Expected an error when deleting non-existent product")
//         st, ok := status.FromError(err)
//         s.True(ok, "Error should be a gRPC status error")
//         s.Equal(codes.NotFound, st.Code(), "Expected NotFound error code")
//     })

//     s.Run("Create Product - Invalid Input", func() {
//         // Example: Missing Title (assuming Title is required by the service)
//         _, err := s.Client.CreateProduct(s.Ctx, &pb.CreateProductRequest{
//             Product: &pb.Product{
//                 // Title:       "Missing Title Test", // Intentionally missing
//                 Price:       10.0,
//                 Sku:         "INVALID-TEST-SKU",
//                 Description: "Testing invalid creation",
//             },
//         })
//         s.Error(err, "Expected an error when creating product with invalid input")
//         st, ok := status.FromError(err)
//         s.True(ok, "Error should be a gRPC status error")
//         // Expecting InvalidArgument, but could be FailedPrecondition depending on validation logic
//         s.Equal(codes.InvalidArgument, st.Code(), "Expected InvalidArgument error code for missing required field")
//     })
// }

// func (s *ProductServiceTestSuite) TestBrandOperations() {
//     s.Run("Create and Get Brand", func() {
//         // Create
//         brand := s.CreateTestBrand()
//         s.NotEmpty(brand.Id)
//         s.Equal("Test Brand", brand.Name) // Assuming helper creates this name

//         // Get
//         got, err := s.Client.GetBrand(s.Ctx, &pb.GetBrandRequest{
//             Identifier: &pb.BrandIdentifier{
//                 Id: brand.Id,
//             },
//         })
//         s.NoError(err)
//         s.Require().NotNil(got)
//         s.Equal(brand.Id, got.Id)
//         s.Equal(brand.Name, got.Name)
//         s.Equal(brand.Description, got.Description)
//     })

//     s.Run("Get Brand - Not Found", func() {
//         _, err := s.Client.GetBrand(s.Ctx, &pb.GetBrandRequest{
//             Identifier: &pb.BrandIdentifier{ Id: "non-existent-brand-id" },
//         })
//         s.Error(err)
//         st, _ := status.FromError(err)
//         s.Equal(codes.NotFound, st.Code())
//     })

//     s.Run("List Brands", func() {
//         // Create at least one brand for the list
//         b1 := s.CreateTestBrand()
//         s.NotEmpty(b1.Id)

//         limit := int32(10)
//         response, err := s.Client.ListBrands(s.Ctx, &pb.ListBrandsRequest{
//             Page:  1,
//             Limit: limit,
//         })
//         s.NoError(err)
//         s.Require().NotNil(response)
//         s.NotEmpty(response.Brands)
//         s.LessOrEqual(int32(len(response.Brands)), limit)
//         s.GreaterOrEqual(response.Total, int32(1)) // At least 1

//         // Check if created brand is in the list
//         foundB1 := false
//         for _, b := range response.Brands {
//             if b.Id == b1.Id {
//                 foundB1 = true
//                 break
//             }
//         }
//         s.True(foundB1, "Created brand not found in list")
//     })

//     s.Run("Update Brand - Success", func() {
//         created := s.CreateTestBrand()
//         s.NotEmpty(created.Id)

//         newName := "Updated Brand Name"
//         newDesc := "Updated Brand Desc"
//         updated, err := s.Client.UpdateBrand(s.Ctx, &pb.UpdateBrandRequest{
//             Brand: &pb.Brand{
//                 Id:          created.Id,
//                 Name:        newName,
//                 Description: newDesc,
//             },
//         })
//         s.NoError(err)
//         s.Require().NotNil(updated)
//         s.Equal(created.Id, updated.Id)
//         s.Equal(newName, updated.Name)
//         s.Equal(newDesc, updated.Description)

//         // Verify with Get
//         got, err := s.Client.GetBrand(s.Ctx, &pb.GetBrandRequest{Identifier: &pb.BrandIdentifier{Id: created.Id}})
//         s.NoError(err)
//         s.Require().NotNil(got)
//         s.Equal(newName, got.Name)
//         s.Equal(newDesc, got.Description)
//     })

//      s.Run("Update Brand - Not Found", func() {
//         _, err := s.Client.UpdateBrand(s.Ctx, &pb.UpdateBrandRequest{
//             Brand: &pb.Brand{ Id: "non-existent-brand-id", Name: "Update Fail" },
//         })
//         s.Error(err)
//         st, _ := status.FromError(err)
//         s.Equal(codes.NotFound, st.Code())
//     })

//     s.Run("Delete Brand - Success", func() {
//         created := s.CreateTestBrand()
//         s.NotEmpty(created.Id)

//         resp, err := s.Client.DeleteBrand(s.Ctx, &pb.DeleteBrandRequest{Id: created.Id})
//         s.NoError(err)
//         s.Require().NotNil(resp)
//         s.True(resp.Success)

//         // Verify with Get
//         _, err = s.Client.GetBrand(s.Ctx, &pb.GetBrandRequest{Identifier: &pb.BrandIdentifier{Id: created.Id}})
//         s.Error(err)
//         st, _ := status.FromError(err)
//         s.Equal(codes.NotFound, st.Code())
//     })

//      s.Run("Delete Brand - Not Found", func() {
//         _, err := s.Client.DeleteBrand(s.Ctx, &pb.DeleteBrandRequest{Id: "non-existent-brand-id"})
//         s.Error(err) // Assuming NotFound is correct
//         st, _ := status.FromError(err)
//         s.Equal(codes.NotFound, st.Code())
//     })
// }

// func (s *ProductServiceTestSuite) TestCategoryOperations() {
//     s.Run("Create and Get Category", func() {
//         // Create
//         category := s.CreateTestCategory()
//         s.NotEmpty(category.Id)
//         s.Equal("Test Category", category.Name) // Assuming helper creates this name

//         // Get
//         got, err := s.Client.GetCategory(s.Ctx, &pb.GetCategoryRequest{
//             Identifier: &pb.CategoryIdentifier{
//                 Id: category.Id,
//             },
//         })
//         s.NoError(err)
//         s.Require().NotNil(got)
//         s.Equal(category.Id, got.Id)
//         s.Equal(category.Name, got.Name)
//         s.Equal(category.Description, got.Description)
//     })

//      s.Run("Get Category - Not Found", func() {
//         _, err := s.Client.GetCategory(s.Ctx, &pb.GetCategoryRequest{
//             Identifier: &pb.CategoryIdentifier{ Id: "non-existent-cat-id" },
//         })
//         s.Error(err)
//         st, _ := status.FromError(err)
//         s.Equal(codes.NotFound, st.Code())
//     })

//     s.Run("List Categories", func() {
//         // Create at least one category
//         c1 := s.CreateTestCategory()
//         s.NotEmpty(c1.Id)

//         limit := int32(10)
//         response, err := s.Client.ListCategories(s.Ctx, &pb.ListCategoriesRequest{
//             Page:  1,
//             Limit: limit,
//         })
//         s.NoError(err)
//         s.Require().NotNil(response)
//         s.NotEmpty(response.Categories)
//         s.LessOrEqual(int32(len(response.Categories)), limit)
//         s.GreaterOrEqual(response.Total, int32(1)) // At least 1

//         // Check if created category is in the list
//         foundC1 := false
//         for _, c := range response.Categories {
//             if c.Id == c1.Id {
//                 foundC1 = true
//                 break
//             }
//         }
//         s.True(foundC1, "Created category not found in list")
//     })

//     s.Run("Update Category - Success", func() {
//         created := s.CreateTestCategory()
//         s.NotEmpty(created.Id)

//         newName := "Updated Category Name"
//         newDesc := "Updated Category Desc"
//         updated, err := s.Client.UpdateCategory(s.Ctx, &pb.UpdateCategoryRequest{
//             Category: &pb.Category{
//                 Id:          created.Id,
//                 Name:        newName,
//                 Description: newDesc,
//             },
//         })
//         s.NoError(err)
//         s.Require().NotNil(updated)
//         s.Equal(created.Id, updated.Id)
//         s.Equal(newName, updated.Name)
//         s.Equal(newDesc, updated.Description)

//         // Verify with Get
//         got, err := s.Client.GetCategory(s.Ctx, &pb.GetCategoryRequest{Identifier: &pb.CategoryIdentifier{Id: created.Id}})
//         s.NoError(err)
//         s.Require().NotNil(got)
//         s.Equal(newName, got.Name)
//         s.Equal(newDesc, got.Description)
//     })

//      s.Run("Update Category - Not Found", func() {
//         _, err := s.Client.UpdateCategory(s.Ctx, &pb.UpdateCategoryRequest{
//             Category: &pb.Category{ Id: "non-existent-cat-id", Name: "Update Fail" },
//         })
//         s.Error(err)
//         st, _ := status.FromError(err)
//         s.Equal(codes.NotFound, st.Code())
//     })

//     s.Run("Delete Category - Success", func() {
//         created := s.CreateTestCategory()
//         s.NotEmpty(created.Id)

//         resp, err := s.Client.DeleteCategory(s.Ctx, &pb.DeleteCategoryRequest{Id: created.Id})
//         s.NoError(err)
//         s.Require().NotNil(resp)
//         s.True(resp.Success)

//         // Verify with Get
//         _, err = s.Client.GetCategory(s.Ctx, &pb.GetCategoryRequest{Identifier: &pb.CategoryIdentifier{Id: created.Id}})
//         s.Error(err)
//         st, _ := status.FromError(err)
//         s.Equal(codes.NotFound, st.Code())
//     })

//      s.Run("Delete Category - Not Found", func() {
//         _, err := s.Client.DeleteCategory(s.Ctx, &pb.DeleteCategoryRequest{Id: "non-existent-cat-id"})
//         s.Error(err) // Assuming NotFound is correct
//         st, _ := status.FromError(err)
//         s.Equal(codes.NotFound, st.Code())
//     })
// }
//         response, err := s.Client.DeleteProduct(s.Ctx, &pb.DeleteProductRequest{
//             Id: created.Id,
//         })
//         s.NoError(err)
//         s.True(response.Success)

//         // Verify deletion
//         _, err = s.Client.GetProduct(s.Ctx, &pb.GetProductRequest{
//             Identifier: &pb.ProductIdentifier{
//                 Id: created.Id,
//             },
//         })
//         s.Error(err)
//         s.Equal(codes.NotFound, status.Code(err))
//     })
// }

// func (s *ProductServiceTestSuite) TestBrandOperations() {
//     s.Run("Create and Get Brand", func() {
//         brand := s.CreateTestBrand()
//         s.NotEmpty(brand.Id)

//         got, err := s.Client.GetBrand(s.Ctx, &pb.GetBrandRequest{
//             Identifier: &pb.BrandIdentifier{
//                 Id: brand.Id,
//             },
//         })
//         s.NoError(err)
//         s.Equal(brand.Id, got.Id)
//     })

//     s.Run("List Brands", func() {
//         response, err := s.Client.ListBrands(s.Ctx, &pb.ListBrandsRequest{
//             Page:  1,
//             Limit: 10,
//         })
//         s.NoError(err)
//         s.NotEmpty(response.Brands)
//     })
// }

// func (s *ProductServiceTestSuite) TestCategoryOperations() {
//     s.Run("Create and Get Category", func() {
//         category := s.CreateTestCategory()
//         s.NotEmpty(category.Id)

//         got, err := s.Client.GetCategory(s.Ctx, &pb.GetCategoryRequest{
//             Identifier: &pb.CategoryIdentifier{
//                 Id: category.Id,
//             },
//         })
//         s.NoError(err)
//         s.Equal(category.Id, got.Id)
//     })

//     s.Run("List Categories", func() {
//         response, err := s.Client.ListCategories(s.Ctx, &pb.ListCategoriesRequest{
//             Page:  1,
//             Limit: 10,
//         })
//         s.NoError(err)
//         s.NotEmpty(response.Categories)
//     })
// }

