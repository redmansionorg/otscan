package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleListNodes(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()

	nodes, err := s.nodeSvc.GetAllNodes(ctx)
	if err != nil {
		errJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	okJSON(c, nodes)
}

func (s *Server) handleGetNode(c *gin.Context) {
	name := c.Param("name")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	node, err := s.nodeSvc.GetNode(ctx, name)
	if err != nil {
		errJSON(c, http.StatusNotFound, err.Error())
		return
	}
	okJSON(c, node)
}

func (s *Server) handleNodeCalendar(c *gin.Context) {
	name := c.Param("name")
	url := s.findNodeURL(name)
	if url == "" {
		errJSON(c, http.StatusNotFound, "node not found")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()

	batches, err := s.rpcClient.GetPendingBatches(ctx, url)
	if err != nil {
		errJSON(c, http.StatusBadGateway, "failed to query node: "+err.Error())
		return
	}

	type CalendarBatchInfo struct {
		BatchID        string `json:"batchID"`
		Status         string `json:"status"`
		CalendarServer string `json:"calendarServer"`
		AttemptCount   uint32 `json:"attemptCount"`
		LastAttemptAt  string `json:"lastAttemptAt,omitempty"`
		LastError      string `json:"lastError,omitempty"`
		StartBlock     uint64 `json:"startBlock"`
		EndBlock       uint64 `json:"endBlock"`
		RUIDCount      uint32 `json:"ruidCount"`
	}

	results := make([]CalendarBatchInfo, 0, len(batches))
	for _, b := range batches {
		results = append(results, CalendarBatchInfo{
			BatchID:        b.BatchID,
			Status:         b.Status,
			CalendarServer: b.CalendarServer,
			AttemptCount:   b.AttemptCount,
			LastAttemptAt:  b.LastAttemptAt,
			LastError:      b.LastError,
			StartBlock:     b.StartBlock,
			EndBlock:       b.EndBlock,
			RUIDCount:      b.RUIDCount,
		})
	}

	okJSON(c, results)
}

func (s *Server) handleNodeCalendarURLStatus(c *gin.Context) {
	name := c.Param("name")
	url := s.findNodeURL(name)
	if url == "" {
		errJSON(c, http.StatusNotFound, "node not found")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()

	statuses, err := s.rpcClient.GetCalendarURLStatus(ctx, url)
	if err != nil {
		errJSON(c, http.StatusBadGateway, "failed to query node: "+err.Error())
		return
	}

	okJSON(c, statuses)
}

func (s *Server) handleNodeHistory(c *gin.Context) {
	name := c.Param("name")
	limitStr := c.DefaultQuery("limit", "360")
	limit, _ := strconv.Atoi(limitStr)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	history, err := s.db.GetNodeHistory(ctx, name, limit)
	if err != nil {
		errJSON(c, http.StatusNotFound, err.Error())
		return
	}
	okJSON(c, history)
}
