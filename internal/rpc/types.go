package rpc

import "encoding/json"

// JSON-RPC 2.0 request/response
type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      int             `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// OTS RPC response types (matching node API)

type StatusResult struct {
	Enabled      bool        `json:"enabled"`
	Status       string      `json:"status"`
	Mode         string      `json:"mode,omitempty"`
	PendingCount int         `json:"pendingCount"`
	Stats        *BatchStats `json:"stats,omitempty"`
}

type BatchStats struct {
	Total             int    `json:"total"`
	Pending           int    `json:"pending"`
	Submitted         int    `json:"submitted"`
	Confirmed         int    `json:"confirmed"`
	Anchored          int    `json:"anchored"`
	Failed            int    `json:"failed"`
	TotalBatches      int    `json:"totalBatches"`
	LastBatchEndBlock uint64 `json:"lastBatchEndBlock"`
	TotalClaims       int    `json:"totalClaims"`
}

type HealthResult struct {
	Status             string                     `json:"status"`
	Mode               string                     `json:"mode"`
	State              string                     `json:"state"`
	PendingCount       int                        `json:"pendingCount"`
	TotalCreated       int                        `json:"totalCreated"`
	TotalConfirmed     int                        `json:"totalConfirmed"`
	LastProcessedBlock uint64                     `json:"lastProcessedBlock"`
	Components         map[string]ComponentStatus `json:"components"`
	LastAnchor         *string                    `json:"lastAnchor,omitempty"`
}

type ComponentStatus struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message,omitempty"`
}

type BatchResult struct {
	BatchID        string `json:"batchID"`
	OnChainID      uint64 `json:"onChainID,omitempty"`
	StartBlock     uint64 `json:"startBlock"`
	EndBlock       uint64 `json:"endBlock"`
	RootHash       string `json:"rootHash"`
	OTSDigest      string `json:"otsDigest,omitempty"`
	RUIDCount      uint32 `json:"ruidCount"`
	CreatedAt      int64  `json:"createdAt"`
	TriggerType    string `json:"triggerType"`
	Status         string `json:"status"`
	BTCTxID        string `json:"btcTxID,omitempty"`
	BTCBlockHeight uint64 `json:"btcBlockHeight,omitempty"`
	BTCTimestamp   uint64 `json:"btcTimestamp,omitempty"`
	CalendarServer string `json:"calendarServer,omitempty"`
	AnchoredBy     string `json:"anchoredBy,omitempty"`
	AnchorBlock    uint64 `json:"anchorBlock,omitempty"`
}

type BatchSummary struct {
	BatchID        string `json:"batchID"`
	OnChainID      uint64 `json:"onChainID,omitempty"`
	StartBlock     uint64 `json:"startBlock"`
	EndBlock       uint64 `json:"endBlock"`
	RUIDCount      uint32 `json:"ruidCount"`
	Status         string `json:"status"`
	AttemptCount   uint32 `json:"attemptCount"`
	LastAttemptAt  string `json:"lastAttemptAt,omitempty"`
	LastError      string `json:"lastError,omitempty"`
	CalendarServer string `json:"calendarServer,omitempty"`
	AnchoredBy     string `json:"anchoredBy,omitempty"`
	AnchorBlock    uint64 `json:"anchorBlock,omitempty"`
}

type RUIDsResult struct {
	BatchID string   `json:"batchID"`
	Total   uint32   `json:"total"`
	Offset  uint32   `json:"offset"`
	Limit   uint32   `json:"limit"`
	RUIDs   []string `json:"ruids"`
}

type ProofResult struct {
	RUID        string `json:"ruid"`
	BatchID     string `json:"batchID"`
	RootHash    string `json:"rootHash"`
	MerkleProof string `json:"merkleProof,omitempty"`
	OTSProof    string `json:"otsProof,omitempty"`
}

type VerifyResult struct {
	RUID           string `json:"ruid"`
	Verified       bool   `json:"verified"`
	BatchID        string `json:"batchID,omitempty"`
	BTCBlockHeight uint64 `json:"btcBlockHeight,omitempty"`
	BTCTimestamp   uint64 `json:"btcTimestamp,omitempty"`
	Message        string `json:"message,omitempty"`
}

type OTSProofResult struct {
	BatchID        string `json:"batchID"`
	RootHash       string `json:"rootHash"`
	OTSDigest      string `json:"otsDigest,omitempty"`
	OTSProof       string `json:"otsProof,omitempty"`
	HasProof       bool   `json:"hasProof"`
	BTCConfirmed   bool   `json:"btcConfirmed"`
	BTCTxID        string `json:"btcTxID,omitempty"`
	BTCBlockHeight uint64 `json:"btcBlockHeight,omitempty"`
	BTCTimestamp   uint64 `json:"btcTimestamp,omitempty"`
	Message        string `json:"message,omitempty"`
}

type ClaimRecordResult struct {
	RUID         string `json:"ruid"`
	Claimant     string `json:"claimant,omitempty"`
	SubmitBlock  uint64 `json:"submitBlock,omitempty"`
	SubmitTime   uint64 `json:"submitTime,omitempty"`
	Published    bool   `json:"published"`
	AUID         string `json:"auid,omitempty"`
	PUID         string `json:"puid,omitempty"`
	PublishBlock uint64 `json:"publishBlock,omitempty"`
	PublishTime  uint64 `json:"publishTime,omitempty"`
}

type ListResult struct {
	Items      []ClaimRecordResult `json:"items"`
	TotalCount uint32              `json:"totalCount"`
	Offset     uint32              `json:"offset"`
	Limit      uint32              `json:"limit"`
}

type ConflictResult struct {
	AUID        string             `json:"auid"`
	HasConflict bool               `json:"hasConflict"`
	ClaimCount  uint32             `json:"claimCount"`
	RUIDs       []string           `json:"ruids,omitempty"`
	Earliest    *ClaimRecordResult `json:"earliest,omitempty"`
}

// LogEntry represents an Ethereum log entry from eth_getLogs.
type LogEntry struct {
	Address     string   `json:"address"`
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
	BlockNumber string   `json:"blockNumber"`
	TxHash      string   `json:"transactionHash"`
	TxIndex     string   `json:"transactionIndex"`
	BlockHash   string   `json:"blockHash"`
	LogIndex    string   `json:"logIndex"`
	Removed     bool     `json:"removed"`
}
