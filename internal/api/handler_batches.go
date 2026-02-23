package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleListBatches(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	result, err := s.batchSvc.ListBatches(ctx, status, page, pageSize)
	if err != nil {
		errJSON(c, http.StatusBadGateway, err.Error())
		return
	}
	okJSON(c, result)
}

func (s *Server) handleGetBatch(c *gin.Context) {
	batchID := c.Param("id")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	batch, err := s.batchSvc.GetBatch(ctx, batchID)
	if err != nil {
		errJSON(c, http.StatusNotFound, err.Error())
		return
	}
	okJSON(c, batch)
}

func (s *Server) handleGetBatchRUIDs(c *gin.Context) {
	batchID := c.Param("id")
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "100")

	offset, _ := strconv.ParseUint(offsetStr, 10, 32)
	limit, _ := strconv.ParseUint(limitStr, 10, 32)
	if limit == 0 || limit > 1000 {
		limit = 100
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	ruids, err := s.batchSvc.GetBatchRUIDs(ctx, batchID, int(offset), int(limit))
	if err != nil {
		errJSON(c, http.StatusNotFound, err.Error())
		return
	}
	okJSON(c, ruids)
}
