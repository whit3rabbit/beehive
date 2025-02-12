package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/whit3rabbit/beehive/manager/models"
	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
)

// taskRequest defines the structure for task creation requests.
type taskRequest struct {
	Task models.Task `json:"task" validate:"required"`
}

// CreateTask handles POST /task/create.
// MaxTaskOutputSize defines the maximum size for task output (in bytes).
const MaxTaskOutputSize = 1024 * 1024 // 1MB

// CreateTask handles POST /task/create.
// It accepts a task creation request and inserts a new task.
func CreateTask(c echo.Context) error {
	var req taskRequest
	if err := c.Bind(&req); err != nil {
		logger.Error("Invalid request payload", zap.Error(err))
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	c.Set("body", req)

	var validTaskTypes = map[string]bool{
		"scan":    true,
		"execute": true,
		// add other valid types
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

	task.ID = primitive.NewObjectID()

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

	if !validTaskTypes[task.Type] {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid task type"})
	}

	// Validate task output size
	if len(task.Output) > MaxTaskOutputSize {
		logger.Error("Task output exceeds size limit", zap.Int("output_size", len(task.Output)), zap.Int("max_size", MaxTaskOutputSize))
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Task output exceeds size limit"})
	}

	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := collection.InsertOne(ctx, task); err != nil {
		logger.Error("Failed to create task", zap.Error(err), zap.String("task_type", task.Type), zap.String("agent_id", task.AgentID))
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create task"})
	}

	response := echo.Map{
		"task_id":   task.ID.Hex(),
		"status":    "queued", // initial status
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

	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid task ID format"})
	}

	var task models.Task
	if err := collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&task); err != nil {
		logger.Error("Task not found", zap.Error(err), zap.String("task_id", taskID))
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Task not found"})
	}

	// Check for timeout
	if task.Status == "running" && task.Timeout > 0 && !task.StartedAt.IsZero() {
		if time.Since(task.StartedAt) > time.Duration(task.Timeout)*time.Second {
			// Update task status to "timeout"
			update := bson.M{
				"$set": bson.M{
					"status":     "timeout",
					"updated_at": time.Now(),
				},
			}
			_, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
			if err != nil {
				logger.Error("Failed to update task status to timeout", zap.Error(err), zap.String("task_id", taskID))
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to retrieve task status"})
			}
			task.Status = "timeout" // Update local task status
		}
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

	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid task ID format"})
	}

	update := bson.M{
		"$set": bson.M{
			"status":     "cancelled",
			"updated_at": time.Now(),
		},
	}
	res, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil || res.MatchedCount == 0 {
		logger.Error("Failed to cancel task", zap.Error(err), zap.String("task_id", taskID))
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to cancel task"})
	}

	response := echo.Map{
		"task_id": taskID,
		"status":    "cancelled",
		"timestamp": time.Now(),
	}
	return c.JSON(http.StatusOK, response)
}
