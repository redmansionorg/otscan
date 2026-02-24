package api

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type VerifyRequest struct {
	RUID string `json:"ruid" binding:"required"`
}

// OTSOperation represents a parsed operation from the OTS proof.
type OTSOperation struct {
	Op       string `json:"op"`                 // "append", "prepend", "sha256", "ripemd160", "keccak256", "reverse"
	Argument string `json:"argument,omitempty"` // hex-encoded argument for append/prepend
}

// OTSAttestation represents a parsed attestation from the OTS proof.
type OTSAttestation struct {
	Type           string `json:"type"`                     // "bitcoin" or "pending"
	BTCBlockHeight uint64 `json:"btcBlockHeight,omitempty"` // for bitcoin
	CalendarURL    string `json:"calendarUrl,omitempty"`    // for pending
}

// ParsedOTSProof is the structured representation of an OTS proof.
type ParsedOTSProof struct {
	Digest       string           `json:"digest"`                 // initial digest (hex)
	HashType     string           `json:"hashType"`               // "sha256"
	Operations   []OTSOperation   `json:"operations"`
	Attestations []OTSAttestation `json:"attestations"`
}

type VerifyResponse struct {
	RUID           string `json:"ruid"`
	Verified       bool   `json:"verified"`
	BatchID        string `json:"batchID,omitempty"`
	BTCBlockHeight uint64 `json:"btcBlockHeight,omitempty"`
	BTCTimestamp   uint64 `json:"btcTimestamp,omitempty"`
	Message        string `json:"message,omitempty"`
	// Claim info from database (if published)
	Published  bool   `json:"published,omitempty"`
	AUID       string `json:"auid,omitempty"`
	PUID       string `json:"puid,omitempty"`
	Claimant   string `json:"claimant,omitempty"`
	SubmitBlock uint64 `json:"submitBlock,omitempty"`
	// Proof data (populated when verified)
	RootHash       string          `json:"rootHash,omitempty"`
	OTSDigest      string          `json:"otsDigest,omitempty"`
	MerkleProof    string          `json:"merkleProof,omitempty"`
	OTSProof       string          `json:"otsProof,omitempty"`
	ParsedOTSProof *ParsedOTSProof `json:"parsedOTSProof,omitempty"`
	LeafIndex      *uint32         `json:"leafIndex,omitempty"`
	LeafCount      uint32          `json:"leafCount,omitempty"`
}

func (s *Server) handleVerify(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errJSON(c, http.StatusBadRequest, "ruid is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	url := s.pickNode()
	result, err := s.rpcClient.VerifyRUID(ctx, url, req.RUID)
	if err != nil {
		errJSON(c, http.StatusBadGateway, err.Error())
		return
	}

	resp := VerifyResponse{
		RUID:           result.RUID,
		Verified:       result.Verified,
		BatchID:        result.BatchID,
		BTCBlockHeight: result.BTCBlockHeight,
		BTCTimestamp:   result.BTCTimestamp,
		Message:        result.Message,
	}

	// Query claim info from database (to get AUID/PUID if published)
	if claim, err := s.db.GetClaim(ctx, req.RUID); err == nil && claim != nil {
		resp.Published = claim.Published
		resp.Claimant = claim.Claimant
		resp.SubmitBlock = claim.SubmitBlock
		if claim.AUID != "" {
			resp.AUID = claim.AUID
		}
		if claim.PUID != "" {
			resp.PUID = claim.PUID
		}
	}

	// If verified, fetch proof data
	if result.Verified && result.BatchID != "" {
		proof, err := s.rpcClient.GetProof(ctx, url, req.RUID, result.BatchID)
		if err == nil {
			resp.RootHash = proof.RootHash
			resp.MerkleProof = proof.MerkleProof
			resp.OTSProof = proof.OTSProof
			leafIdx := proof.LeafIndex
			resp.LeafIndex = &leafIdx
			resp.LeafCount = proof.LeafCount

			// Parse OTS proof into structured form
			if proof.OTSProof != "" {
				parsed := parseOTSProofHex(proof.OTSProof)
				if parsed != nil {
					resp.ParsedOTSProof = parsed
				}
			}
		}
		// Also get OTS digest
		otsProof, err := s.rpcClient.GetOTSProof(ctx, url, result.BatchID)
		if err == nil {
			resp.OTSDigest = otsProof.OTSDigest
		}
	}

	okJSON(c, resp)
}

func (s *Server) handleGetProof(c *gin.Context) {
	batchID := c.Param("batchId")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	result, err := s.rpcClient.GetOTSProof(ctx, s.pickNode(), batchID)
	if err != nil {
		errJSON(c, http.StatusNotFound, err.Error())
		return
	}

	okJSON(c, result)
}

// --- OTS proof parser (mirrors ots/opentimestamps/parser.go) ---

var otsMagicHeader = []byte{
	0x00, 0x4f, 0x70, 0x65, 0x6e, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x73, 0x00,
	0x00, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x00, 0xbf,
	0x89, 0xe2, 0xe8, 0x84, 0xe8, 0x92, 0x94,
}

var btcAttestMagic = []byte{0x05, 0x88, 0x96, 0x0d, 0x73, 0xd7, 0x19, 0x01}
var pendingAttestMagic = []byte{0x83, 0xdf, 0xe3, 0x0d, 0x2e, 0xf9, 0x0c, 0x8e}

func parseOTSProofHex(proofHex string) *ParsedOTSProof {
	if len(proofHex) < 2 {
		return nil
	}
	raw := proofHex
	if raw[:2] == "0x" || raw[:2] == "0X" {
		raw = raw[2:]
	}
	data, err := hex.DecodeString(raw)
	if err != nil || len(data) < len(otsMagicHeader)+3 {
		return nil
	}

	// Verify magic header
	for i, b := range otsMagicHeader {
		if data[i] != b {
			return nil
		}
	}
	offset := len(otsMagicHeader)

	// Version
	_ = data[offset] // version
	offset++

	// Hash type
	hashType := data[offset]
	offset++
	var digestSize int
	var hashName string
	switch hashType {
	case 0x08:
		digestSize = 32
		hashName = "sha256"
	case 0x02:
		digestSize = 20
		hashName = "sha1"
	case 0x03:
		digestSize = 20
		hashName = "ripemd160"
	default:
		return nil
	}

	if offset+digestSize > len(data) {
		return nil
	}
	digest := data[offset : offset+digestSize]
	offset += digestSize

	parsed := &ParsedOTSProof{
		Digest:       "0x" + hex.EncodeToString(digest),
		HashType:     hashName,
		Operations:   make([]OTSOperation, 0),
		Attestations: make([]OTSAttestation, 0),
	}

	// Parse operations and attestations
	for offset < len(data) {
		tag := data[offset]
		offset++

		switch tag {
		case 0xf0, 0xf1: // append, prepend
			argLen, n := readVarInt(data, offset)
			if n == 0 {
				return parsed
			}
			offset += n
			if offset+int(argLen) > len(data) {
				return parsed
			}
			arg := data[offset : offset+int(argLen)]
			offset += int(argLen)
			opName := "append"
			if tag == 0xf1 {
				opName = "prepend"
			}
			parsed.Operations = append(parsed.Operations, OTSOperation{
				Op:       opName,
				Argument: "0x" + hex.EncodeToString(arg),
			})

		case 0x08: // SHA256
			parsed.Operations = append(parsed.Operations, OTSOperation{Op: "sha256"})
		case 0x03: // RIPEMD160
			parsed.Operations = append(parsed.Operations, OTSOperation{Op: "ripemd160"})
		case 0x67: // KECCAK256
			parsed.Operations = append(parsed.Operations, OTSOperation{Op: "keccak256"})
		case 0xf2: // reverse
			parsed.Operations = append(parsed.Operations, OTSOperation{Op: "reverse"})
		case 0x00: // fork
			continue

		default:
			// Check for attestation
			offset-- // back up
			if offset+8 <= len(data) && matchBytes(data[offset:offset+8], btcAttestMagic) {
				offset += 8
				payloadLen, n := readVarInt(data, offset)
				if n == 0 {
					return parsed
				}
				offset += n
				blockHeight := uint64(0)
				if payloadLen > 0 && offset+int(payloadLen) <= len(data) {
					blockHeight, _ = readVarInt(data, offset)
				}
				offset += int(payloadLen)
				parsed.Attestations = append(parsed.Attestations, OTSAttestation{
					Type:           "bitcoin",
					BTCBlockHeight: blockHeight,
				})
			} else if offset+8 <= len(data) && matchBytes(data[offset:offset+8], pendingAttestMagic) {
				offset += 8
				urlLen, n := readVarInt(data, offset)
				if n == 0 {
					return parsed
				}
				offset += n
				calURL := ""
				if offset+int(urlLen) <= len(data) {
					calURL = string(data[offset : offset+int(urlLen)])
				}
				offset += int(urlLen)
				parsed.Attestations = append(parsed.Attestations, OTSAttestation{
					Type:        "pending",
					CalendarURL: calURL,
				})
			} else {
				return parsed
			}
		}
	}

	return parsed
}

func readVarInt(data []byte, offset int) (uint64, int) {
	if offset >= len(data) {
		return 0, 0
	}
	var value uint64
	var shift uint
	consumed := 0
	for {
		if offset+consumed >= len(data) {
			return 0, 0
		}
		b := data[offset+consumed]
		consumed++
		value |= uint64(b&0x7f) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
		if shift >= 64 {
			return 0, 0
		}
	}
	return value, consumed
}

func matchBytes(a, b []byte) bool {
	if len(a) < len(b) {
		return false
	}
	for i := range b {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func init() {
	// suppress unused import
	_ = fmt.Sprintf
}
