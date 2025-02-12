package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/whit3rabbit/beehive/manager/models"
	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
)

var validate = validator.New()

// CreateTask handles POST /task/create.
// It accepts a task creation request and inserts a new task.
func CreateTask(c echo.Context) error {
	var req struct {
		Task models.Task `json:"task" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}
	// Validate the request structure.
	if err := validate.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Validation failed", "details": err.Error()})
	}

	task := req.Task
	now := time.Now()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	if task.UpdatedAt.IsZero() {
		task.UpdatedAt = now
	}
	// Set default task status if not provided
	if task.Status == "" {
		task.Status = "queued"
	}

	var validStatuses = map[string]bool{
		"queued":    true,
		"running":   true,
		"completed": true,
		"failed":    true,
		"cancelled": true,
	}

	if !validStatuses[task.Status] {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid status"})
	}

	collection := mongodb.Client.Database(os.Getenv("MONGODB_DATABASE")).Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := collection.InsertOne(ctx, task); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create task"})
	}

	response := echo.Map{
		"task_id":  task.TaskID,
		"status":   "queued", // initial status
		"timestamp": now,
	}
	return c.JSON(http.StatusOK, response)
}

// GetTaskStatus handles GET /task/status/:task_id.
// It retrieves the status and output for a specific task.
func GetTaskStatus(c echo.Context) error {
	taskID := c.Param("task_id")
	if taskID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Missing task ID"})
	}

	collection := mongodb.Client.Database(os.Getenv("MONGODB_DATABASE")).Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var task models.Task
	if err := collection.FindOne(ctx, bson.M{"task_id": taskID}).Decode(&task); err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Task not found"})
	}
	return c.JSON(http.StatusOK, task)
}

// CancelTask handles POST /task/cancel/:task_id.
// It updates a task's status to cancelled.
func CancelTask(c echo.Context) error {
	taskID := c.Param("task_id")
	if taskID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Missing task ID"})
	}

	collection := mongodb.Client.Database(os.Getenv("MONGODB_DATABASE")).Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"status":     "cancelled",
			"updated_at": time.Now(),
		},
	}
	res, err := collection.UpdateOne(ctx, bson.M{"task_id": taskID}, update)
	if err != nil || res.MatchedCount == 0 {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to cancel task"})
	}

	response := echo.Map{
		"task_id":  taskID,
		"status":   "cancelled",
		"timestamp": time.Now(),
	}
	return c.JSON(http.StatusOK, response)
}
