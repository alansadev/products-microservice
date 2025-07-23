package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"os"
	"products/database"
	"products/models"
	"regexp"
)

type BatchRequest struct {
	IDs []uuid.UUID `json:"ids"`
}

type UpdateStockRequest struct {
	QuantityChange int64 `json:"quantity_change"`
}

// CreateProduct godoc
// @Summary     Create a new Product
// @Description Add a product to database
// @Tags        products
// @Accept      json
// @Produce     json
// @Param       product body models.Product true "Data of Product (ID, CreatedAt, UpdatedAt are ignored)"
// @Success     201 {object} models.Product
// @Router      /products [post]
func CreateProduct(c *fiber.Ctx) error {
	db := database.DB
	product := new(models.Product)

	if err := c.BodyParser(product); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if err := db.Create(&product).Error; err != nil {
		log.Printf("Error creating product in database: %s", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create product"})
	}

	return c.Status(fiber.StatusCreated).JSON(product)
}

// GetProducts godoc
// @Summary      List all Products
// @Description  Return an array with all products
// @Tags         products
// @Produce      json
// @Success      200  {array}   models.Product
// @Router       /products [get]
func GetProducts(c *fiber.Ctx) error {
	db := database.DB
	var products []models.Product

	if err := db.Find(&products).Error; err != nil {
		log.Printf("Error getting products in database: %s", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch products"})
	}
	return c.JSON(products)
}

// GetProductsByIDs godoc
// @Summary      Search multiple products by a list of IDs
// @Description  Returns an array with the products corresponding to the submitted IDs.
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        request body      BatchRequest  true  "List of product IDs"
// @Success      200     {array}   models.Product
// @Router       /products/batch [post]

func GetProductsByIDs(c *fiber.Ctx) error {
	db := database.DB
	payload := new(BatchRequest)

	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if len(payload.IDs) == 0 {
		return c.JSON([]models.Product{})
	}

	var products []models.Product
	if err := db.Where("id IN (?)", payload.IDs).Find(&products).Error; err != nil {
		log.Printf("Error getting products in database: %s", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch products"})
	}
	return c.JSON(products)
}

// GetProductByID godoc
// @Summary     Find product by id
// @Description Return data only unique product
// @Tags        products
// @Produce     json
// @Param       id path string true "Product ID (UUID)"
// @Success     200 {object} models.Product
// @Failure     404 {object} map[string]string
// @Router      /products/{id} [get]
func GetProductByID(c *fiber.Ctx) error {
	db := database.DB
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID format"})
	}

	var product models.Product
	if err := db.First(&product, id).Error; err != nil {
		log.Printf("Error getting product in database: %s", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}

	return c.JSON(product)
}

// PatchProduct godoc
// @Summary      Update a Product
// @Description  Update data of product by exists ID
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      string          true  "Product ID (UUID)"
// @Param        product  body      models.Product  true  "New Product Data"
// @Success      200      {object}  models.Product
// @Failure      404      {object}  map[string]string
// @Router       /products/{id} [patch]
func PatchProduct(c *fiber.Ctx) error {
	db := database.DB
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID format"})
	}

	var product models.Product
	if err := db.First(&product, id).Error; err != nil {
		log.Printf("Error getting product in database: %s", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}

	updateData := make(map[string]interface{})
	if err := c.BodyParser(&updateData); err != nil {
		log.Printf("Error parsing patch request body: %s", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	delete(updateData, "image_url")

	if err := db.Model(&product).Updates(updateData).Error; err != nil {
		log.Printf("Error updating product in database: %s", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update product"})
	}
	return c.JSON(product)
}

// DeleteProduct godoc
// @Summary      Delete a Product
// @Description  Remove a product to database by your ID
// @Tags         products
// @Param        id   path      string  true  "Product ID (UUID)"
// @Success      204  {object}  nil
// @Failure      404  {object}  map[string]string
// @Router       /products/{id} [delete]
func DeleteProduct(c *fiber.Ctx) error {
	db := database.DB
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID format"})
	}

	var product models.Product
	if err := db.First(&product, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}

	if product.ImageURL != nil && *product.ImageURL != "" {
		cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
		if err != nil {
			log.Printf("Error iniatializing Cloudinary for deletion: %s", err)
		} else {
			publicID := extractPublicIDFromURL(*product.ImageURL)
			if publicID != "" {
				_, err := cld.Upload.Destroy(context.Background(), uploader.DestroyParams{PublicID: publicID})
				if err != nil {
					log.Printf("Failed to delete image from Cloudinary with public_id %s: %v", publicID, err)
				} else {
					log.Printf("Successfully deleted image from Cloudinary with public_id: %s", publicID)
				}
			}
		}
	}

	result := db.Delete(&product, id)
	if result.Error != nil {
		log.Printf("Error deleting product: %s", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not delete product"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UploadProductImage godoc
// @Summary     Upload image from product
// @Description Received image and associate url to product
// @Tags        products
// @Accept      multipart/form-data
// @Produce     json
// @Param       id path string true "Product ID (UUID)"
// @Param       image formData file true "Product Image"
// @Success     200 {object} models.Product
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /products/{id}/upload [post]
func UploadProductImage(c *fiber.Ctx) error {
	db := database.DB
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID format"})
	}

	var product models.Product
	if err := db.First(&product, id).Error; err != nil {
		log.Printf("Error getting product in database: %s", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}

	file, err := c.FormFile("image")
	if err != nil {
		log.Printf("Error uploading product image: %s", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Image upload failed"})
	}

	fileReader, err := file.Open()
	if err != nil {
		log.Printf("Error uploading product image: %s", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer func() {
		if err := fileReader.Close(); err != nil {
			log.Printf("Failed to close file reader: %v", err)
		}
	}()

	cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to initialize Cloudinary"})
	}

	uploadParams := uploader.UploadParams{
		Folder: "sabordarondonia",
	}

	uploadResult, err := cld.Upload.Upload(context.Background(), fileReader, uploadParams)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to upload file"})
	}

	imageURL := uploadResult.URL
	product.ImageURL = &imageURL

	if err := db.Save(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update product with image URL"})
	}

	return c.JSON(product)
}

// UpdateStock godoc
// @Summary      Updates the stock of a product
// @Description  Adjusts a product's inventory atomically. Use a negative value to decrease inventory.
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      string              true  "Product ID (UUID)"
// @Param        request  body      UpdateStockRequest  true  "Change in stock quantity"
// @Success      200      {object}  models.Product
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Router       /products/{id}/stock [post]
func UpdateStock(c *fiber.Ctx) error {
	db := database.DB
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID format"})
	}

	payload := new(UpdateStockRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	var updatedProduct models.Product
	err = db.Transaction(func(tx *gorm.DB) error {
		var product models.Product

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, id).Error; err != nil {
			return errors.New("product not found")
		}

		newStock := product.Stock + payload.QuantityChange
		if newStock < 0 {
			return fmt.Errorf("product stock out of stock: %s", product.Name)
		}

		product.Stock = newStock
		if err := tx.Save(&product).Error; err != nil {
			return err
		}

		updatedProduct = product
		return nil
	})
	if err != nil {
		if err.Error() == "Product not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		if errors.Is(err, gorm.ErrInvalidField) || err.Error()[:18] == "insufficient stock" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		log.Printf("Error updating stock in transaction: %s", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update stock"})
	}
	return c.Status(fiber.StatusOK).JSON(updatedProduct)
}

func extractPublicIDFromURL(url string) string {
	re := regexp.MustCompile(`/sabordarondonia/([^.]+)\.`)
	matches := re.FindStringSubmatch(url)

	if len(matches) > 1 {
		return "sabordarondonia/" + matches[1]
	}
	return ""
}
