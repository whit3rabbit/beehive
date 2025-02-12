package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/whit3rabbit/beehive/manager/models"
	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
)

// ListRoles retrieves all defined roles from the database.
// It returns a JSON array of roles or an error if the operation fails.
// @Summary List all roles
// @Description Get all defined roles
// @Tags roles
// @Accept json
// @Produce json
// @Success 200 {array} models.Role
// @Failure 500 {object} ErrorResponse
// @Router /roles [get]
func ListRoles(c echo.Context) error {
	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("roles")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		logger.Error("Failed to retrieve roles", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve roles"})
	}
	defer cursor.Close(ctx)

	var roles []models.Role
	if err = cursor.All(ctx, &roles); err != nil {
		logger.Error("Failed to parse roles", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to parse roles"})
	}
	return c.JSON(http.StatusOK, roles)
}

// CreateRole handles POST /roles.
// @Summary Creates a new role
// @Description Adds a new role to the database.
// @Tags roles
// @Accept json
// @Produce json
// @Param role body models.Role true "Role object to be created"
// @Success 201 {object} models.Role
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /roles [post]
func CreateRole(c echo.Context) error {
	var role models.Role
	if err := c.Bind(&role); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request payload"})
	}

	role.CreatedAt = time.Now()
	// Generate a unique ID for the role using MongoDB's ObjectID.
	role.ID = primitive.NewObjectID()

	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("roles")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := collection.InsertOne(ctx, role); err != nil {
		logger.Error("Failed to create role", zap.Error(err), zap.String("role_name", role.Name))
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create role"})
	}
	return c.JSON(http.StatusCreated, role)
}

// GetRole handles GET /roles/:role_id.
// @Summary Retrieves a specific role by ID
// @Description Gets the details of a role based on its ID.
// @Tags roles
// @Accept json
// @Produce json
// @Param role_id path string true "Role ID"
// @Success 200 {object} models.Role
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /roles/{role_id} [get]
func GetRole(c echo.Context) error {
	roleID := c.Param("role_id")
	if roleID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Missing role ID"})
	}

	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("roles")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var role models.Role
	objID, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		logger.Error("Invalid role ID format", zap.Error(err), zap.String("role_id", roleID))
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid role ID format"})
	}

	if err := collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&role); err != nil {
		logger.Error("Role not found", zap.Error(err), zap.String("role_id", roleID))
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "Role not found"})
	}
	return c.JSON(http.StatusOK, role)
}
