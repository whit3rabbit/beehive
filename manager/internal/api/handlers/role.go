package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"manager/models"
	"manager/mongodb"
)

// ListRoles handles GET /roles.
// It returns all defined roles.
func ListRoles(c echo.Context) error {
	collection := mongodb.Client.Database("manager_db").Collection("roles")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to retrieve roles"})
	}
	defer cursor.Close(ctx)

	var roles []models.Role
	if err = cursor.All(ctx, &roles); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to parse roles"})
	}
	return c.JSON(http.StatusOK, roles)
}

// CreateRole handles POST /roles.
// It creates a new role.
func CreateRole(c echo.Context) error {
	var role models.Role
	if err := c.Bind(&role); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	role.CreatedAt = time.Now()
	// Generate a unique ID for the role using MongoDB's ObjectID.
	role.ID = primitive.NewObjectID().Hex()

	collection := mongodb.Client.Database("manager_db").Collection("roles")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := collection.InsertOne(ctx, role); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create role"})
	}
	return c.JSON(http.StatusCreated, role)
}

// GetRole handles GET /roles/:role_id.
// It retrieves details of a specific role.
func GetRole(c echo.Context) error {
	roleID := c.Param("role_id")
	if roleID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Missing role ID"})
	}

	collection := mongodb.Client.Database("manager_db").Collection("roles")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var role models.Role
	if err := collection.FindOne(ctx, bson.M{"id": roleID}).Decode(&role); err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Role not found"})
	}
	return c.JSON(http.StatusOK, role)
}
