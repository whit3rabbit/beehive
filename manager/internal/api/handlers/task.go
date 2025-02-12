package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"

	"manager/models"
	"manager/internal/mongodb"
)

// CreateTask handles POST /task/create.
// It accepts a task creation request and inserts a new task.
func CreateTask(c echo.Context) error {
	var req struct {
		Task models.Task `json:"task"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	task := req.Task
	now := time.Now()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	if task.UpdatedAt.IsZero() {
		task.UpdatedAt = now
	}

	collection := mongodb.Client.Database("manager_db").Collection("tasks")
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

	collection := mongodb.Client.Database("manager_db").Collection("tasks")
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

	collection := mongodb.Client.Database("manager_db").Collection("tasks")
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
