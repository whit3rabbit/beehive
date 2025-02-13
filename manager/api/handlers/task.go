package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/whit3rabbit/beehive/manager/internal/logger"
	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/whit3rabbit/beehive/manager/models"
	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
)

// TaskRequest defines the structure for task creation requests.
type TaskRequest struct {
	Task models.Task `json:"task" validate:"required"`
}

// MaxTaskOutputSize defines the maximum size for task output (in bytes).
const MaxTaskOutputSize = 1024 * 1024 // 1MB

// CreateTask handles POST /task/create.
// @Summary Creates a new task
// @Description Adds a new task to the database.
// @Tags task
// @Accept json
// @Produce json
// @Param task body TaskRequest true "Task object to be created"
// @Success 200 {object} models.TaskCreationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /task/create [post]
func CreateTask(c echo.Context) error {
	var req TaskRequest
	if err := c.Bind(&req); err != nil {
		logger.Error("Invalid request payload", zap.Error(err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request payload"})
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
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid status"})
	}

	if !validTaskTypes[task.Type] {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid task type"})
	}

	// Validate task output size
	if task.Output != nil {
		// Calculate combined size of logs and error
		outputSize := len(task.Output.Logs)
		if task.Output.Error != "" {
			outputSize += len(task.Output.Error)
		}
		
		if outputSize > MaxTaskOutputSize {
			logger.Error("Task output exceeds size limit", 
				zap.Int("output_size", outputSize), 
				zap.Int("max_size", MaxTaskOutputSize))
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Task output exceeds size limit"})
		}
	}

	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := collection.InsertOne(ctx, task); err != nil {
		logger.Error("Failed to create task", zap.Error(err), zap.String("task_type", task.Type), zap.String("agent_id", task.AgentID))
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create task"})
	}

	response := models.TaskCreationResponse{
		TaskID:    task.ID.Hex(),
		Status:    "queued", // initial status
		Timestamp: now,
	}
	return c.JSON(http.StatusOK, response)
}

// GetTaskStatus handles GET /task/status/:task_id.
// @Summary Retrieves the status of a specific task
// @Description Gets the status and output of a task based on its ID.
// @Tags task
// @Accept json
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {object} models.Task
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /task/status/{task_id} [get]
func GetTaskStatus(c echo.Context) error {
	taskID := c.Param("task_id")
	if taskID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Missing task ID"})
	}

	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid task ID format"})
	}

	var task models.Task
	if err := collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&task); err != nil {
		logger.Error("Task not found", zap.Error(err), zap.String("task_id", taskID))
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "Task not found"})
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
				return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve task status"})
			}
			task.Status = "timeout" // Update local task status
		}
	}

	return c.JSON(http.StatusOK, task)
}

// CancelTask handles POST /task/cancel/:task_id.
// @Summary Cancels a specific task
// @Description Updates the status of a task to "cancelled".
// @Tags task
// @Accept json
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {object} models.TaskCancelResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /task/cancel/{task_id} [post]
func CancelTask(c echo.Context) error {
	taskID := c.Param("task_id")
	if taskID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Missing task ID"})
	}

	dbName := c.Get("mongodb_database").(string)
	collection := mongodb.Client.Database(dbName).Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid task ID format"})
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
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to cancel task"})
	}

	response := models.TaskCancelResponse{
		TaskID:    taskID,
		Status:    "cancelled",
		Timestamp: time.Now(),
	}
	return c.JSON(http.StatusOK, response)
}
