package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

type Client struct {
	httpClient *http.Client
	idCounter  atomic.Int64
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:    30 * time.Second,
			},
		},
	}
}

func (c *Client) Call(ctx context.Context, url, method string, params ...interface{}) (json.RawMessage, error) {
	if params == nil {
		params = []interface{}{}
	}

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      int(c.idCounter.Add(1)),
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// Typed convenience methods

func (c *Client) EthBlockNumber(ctx context.Context, url string) (uint64, error) {
	result, err := c.Call(ctx, url, "eth_blockNumber")
	if err != nil {
		return 0, err
	}
	var hex string
	if err := json.Unmarshal(result, &hex); err != nil {
		return 0, err
	}
	var num uint64
	fmt.Sscanf(hex, "0x%x", &num)
	return num, nil
}

func (c *Client) EthCoinbase(ctx context.Context, url string) (string, error) {
	result, err := c.Call(ctx, url, "eth_coinbase")
	if err != nil {
		return "", err
	}
	var addr string
	if err := json.Unmarshal(result, &addr); err != nil {
		return "", err
	}
	return addr, nil
}

func (c *Client) OTSStatus(ctx context.Context, url string) (*StatusResult, error) {
	result, err := c.Call(ctx, url, "ots_status")
	if err != nil {
		return nil, err
	}
	var status StatusResult
	if err := json.Unmarshal(result, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (c *Client) OTSHealth(ctx context.Context, url string) (*HealthResult, error) {
	result, err := c.Call(ctx, url, "ots_health")
	if err != nil {
		return nil, err
	}
	var health HealthResult
	if err := json.Unmarshal(result, &health); err != nil {
		return nil, err
	}
	return &health, nil
}

func (c *Client) GetBatch(ctx context.Context, url, batchID string) (*BatchResult, error) {
	result, err := c.Call(ctx, url, "ots_getBatch", batchID)
	if err != nil {
		return nil, err
	}
	var batch BatchResult
	if err := json.Unmarshal(result, &batch); err != nil {
		return nil, err
	}
	return &batch, nil
}

func (c *Client) GetPendingBatches(ctx context.Context, url string) ([]*BatchSummary, error) {
	result, err := c.Call(ctx, url, "ots_getPendingBatches")
	if err != nil {
		return nil, err
	}
	var batches []*BatchSummary
	if err := json.Unmarshal(result, &batches); err != nil {
		return nil, err
	}
	return batches, nil
}

type CalendarURLStatus struct {
	URL          string `json:"url"`
	Priority     int    `json:"priority"`
	SuccessCount uint32 `json:"successCount"`
	FailureCount uint32 `json:"failureCount"`
	LastAttemptAt string `json:"lastAttemptAt,omitempty"`
	LastError    string `json:"lastError,omitempty"`
	LastStatus   string `json:"lastStatus"`
}

func (c *Client) GetCalendarURLStatus(ctx context.Context, url string) ([]*CalendarURLStatus, error) {
	result, err := c.Call(ctx, url, "ots_getCalendarURLStatus")
	if err != nil {
		return nil, err
	}
	var statuses []*CalendarURLStatus
	if err := json.Unmarshal(result, &statuses); err != nil {
		return nil, err
	}
	return statuses, nil
}

func (c *Client) GetConfirmedBatches(ctx context.Context, url string, limit uint32) ([]*BatchSummary, error) {
	result, err := c.Call(ctx, url, "ots_getConfirmedBatches", limit)
	if err != nil {
		return nil, err
	}
	var batches []*BatchSummary
	if err := json.Unmarshal(result, &batches); err != nil {
		return nil, err
	}
	return batches, nil
}

func (c *Client) GetRUIDs(ctx context.Context, url, batchID string, offset, limit uint32) (*RUIDsResult, error) {
	result, err := c.Call(ctx, url, "ots_getRUIDs", batchID, offset, limit)
	if err != nil {
		return nil, err
	}
	var ruids RUIDsResult
	if err := json.Unmarshal(result, &ruids); err != nil {
		return nil, err
	}
	return &ruids, nil
}

func (c *Client) VerifyRUID(ctx context.Context, url, ruidHex string) (*VerifyResult, error) {
	result, err := c.Call(ctx, url, "ots_verifyRUID", ruidHex)
	if err != nil {
		return nil, err
	}
	var verify VerifyResult
	if err := json.Unmarshal(result, &verify); err != nil {
		return nil, err
	}
	return &verify, nil
}

func (c *Client) GetProof(ctx context.Context, url, ruidHex, batchID string) (*ProofResult, error) {
	result, err := c.Call(ctx, url, "ots_getProof", ruidHex, batchID)
	if err != nil {
		return nil, err
	}
	var proof ProofResult
	if err := json.Unmarshal(result, &proof); err != nil {
		return nil, err
	}
	return &proof, nil
}

func (c *Client) GetOTSProof(ctx context.Context, url, batchID string) (*OTSProofResult, error) {
	result, err := c.Call(ctx, url, "ots_getOTSProof", batchID)
	if err != nil {
		return nil, err
	}
	var proof OTSProofResult
	if err := json.Unmarshal(result, &proof); err != nil {
		return nil, err
	}
	return &proof, nil
}

func (c *Client) ListByPuid(ctx context.Context, url, puidHex string, offset, limit uint32) (*ListResult, error) {
	result, err := c.Call(ctx, url, "ots_listByPuid", puidHex, offset, limit)
	if err != nil {
		return nil, err
	}
	var list ListResult
	if err := json.Unmarshal(result, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) ListByAuid(ctx context.Context, url, auidHex string, offset, limit uint32) (*ListResult, error) {
	result, err := c.Call(ctx, url, "ots_listByAuid", auidHex, offset, limit)
	if err != nil {
		return nil, err
	}
	var list ListResult
	if err := json.Unmarshal(result, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) ListByClaimant(ctx context.Context, url, claimantHex string, offset, limit uint32) (*ListResult, error) {
	result, err := c.Call(ctx, url, "ots_listByClaimant", claimantHex, offset, limit)
	if err != nil {
		return nil, err
	}
	var list ListResult
	if err := json.Unmarshal(result, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) CheckConflict(ctx context.Context, url, auidHex string) (*ConflictResult, error) {
	result, err := c.Call(ctx, url, "ots_checkConflict", auidHex)
	if err != nil {
		return nil, err
	}
	var conflict ConflictResult
	if err := json.Unmarshal(result, &conflict); err != nil {
		return nil, err
	}
	return &conflict, nil
}

// GetLogs calls eth_getLogs with filter parameters.
func (c *Client) GetLogs(ctx context.Context, url string, fromBlock, toBlock uint64, address string, topics []string) ([]LogEntry, error) {
	filter := map[string]interface{}{
		"fromBlock": fmt.Sprintf("0x%x", fromBlock),
		"toBlock":   fmt.Sprintf("0x%x", toBlock),
		"address":   address,
	}
	if len(topics) > 0 {
		filter["topics"] = []interface{}{topics}
	}

	result, err := c.Call(ctx, url, "eth_getLogs", filter)
	if err != nil {
		return nil, err
	}
	var logs []LogEntry
	if err := json.Unmarshal(result, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}
