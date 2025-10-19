package utils

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func BadRequest(c *gin.Context, err error) {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func NotFound(c *gin.Context) {
    c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

func ServerError(c *gin.Context, err error) {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
