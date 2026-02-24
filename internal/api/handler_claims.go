package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleSearchClaims(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Check for filter-based listing (Phase 2)
	if filter := c.Query("filter"); filter != "" {
		result, err := s.claimSvc.ListClaims(ctx, filter, offset, limit)
		if err != nil {
			errJSON(c, http.StatusInternalServerError, err.Error())
			return
		}
		okJSON(c, result)
		return
	}

	// Sort=latest means list all claims by submit_block desc
	if sort := c.Query("sort"); sort == "latest" {
		result, err := s.claimSvc.ListClaims(ctx, "", offset, limit)
		if err != nil {
			errJSON(c, http.StatusInternalServerError, err.Error())
			return
		}
		okJSON(c, result)
		return
	}

	if claimant := c.Query("claimant"); claimant != "" {
		result, err := s.claimSvc.SearchClaims(ctx, "claimant", claimant, offset, limit)
		if err != nil {
			errJSON(c, http.StatusBadGateway, err.Error())
			return
		}
		okJSON(c, result)
		return
	}
	if auid := c.Query("auid"); auid != "" {
		result, err := s.claimSvc.SearchClaims(ctx, "auid", auid, offset, limit)
		if err != nil {
			errJSON(c, http.StatusBadGateway, err.Error())
			return
		}
		okJSON(c, result)
		return
	}
	if puid := c.Query("puid"); puid != "" {
		result, err := s.claimSvc.SearchClaims(ctx, "puid", puid, offset, limit)
		if err != nil {
			errJSON(c, http.StatusBadGateway, err.Error())
			return
		}
		okJSON(c, result)
		return
	}

	// Default: list all claims (newest first)
	result, err := s.claimSvc.ListClaims(ctx, "", offset, limit)
	if err != nil {
		errJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	okJSON(c, result)
}

func (s *Server) handleGetClaim(c *gin.Context) {
	ruid := c.Param("ruid")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	result, err := s.rpcClient.VerifyRUID(ctx, s.pickNode(), ruid)
	if err != nil {
		errJSON(c, http.StatusNotFound, err.Error())
		return
	}
	okJSON(c, result)
}

func (s *Server) handleCheckConflict(c *gin.Context) {
	auid := c.Param("auid")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	result, err := s.rpcClient.CheckConflict(ctx, s.pickNode(), auid)
	if err != nil {
		errJSON(c, http.StatusBadGateway, err.Error())
		return
	}
	okJSON(c, result)
}

func (s *Server) handleListConflicts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	conflicts, total, err := s.claimSvc.ListConflicts(ctx, offset, limit)
	if err != nil {
		errJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	okJSON(c, map[string]interface{}{
		"items":  conflicts,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}

func (s *Server) handleClaimStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	stats, err := s.claimSvc.GetClaimStats(ctx)
	if err != nil {
		errJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	okJSON(c, stats)
}

func (s *Server) handleListClaimants(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	claimants, total, err := s.claimSvc.ListClaimants(ctx, offset, limit)
	if err != nil {
		errJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	okJSON(c, map[string]interface{}{
		"items":  claimants,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}

func (s *Server) handleListAssets(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	assets, total, err := s.claimSvc.ListAssets(ctx, offset, limit)
	if err != nil {
		errJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	okJSON(c, map[string]interface{}{
		"items":  assets,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}

func (s *Server) handleListPersons(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	persons, total, err := s.claimSvc.ListPersons(ctx, offset, limit)
	if err != nil {
		errJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	okJSON(c, map[string]interface{}{
		"items":  persons,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}
