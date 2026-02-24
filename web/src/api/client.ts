import axios from 'axios';

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
});

api.interceptors.response.use(
  (resp) => resp.data?.data ?? resp.data,
  (err) => {
    const msg = err.response?.data?.error || err.message;
    return Promise.reject(new Error(msg));
  }
);

export default api;

// Typed API functions

export interface NodeStatus {
  name: string;
  rpcUrl: string;
  status: string;
  blockNumber: number;
  otsMode?: string;
  pendingCount: number;
  totalCreated: number;
  totalConfirmed: number;
  lastProcessedBlock: number;
  components?: Record<string, { healthy: boolean; message?: string }>;
  lastAnchor?: string;
  coinbase?: string;
}

export interface DashboardData {
  chainId: number;
  chainName: string;
  nodeCount: number;
  nodesHealthy: number;
  latestBlock: number;
  totalBatches: number;
  pendingBatches: number;
  anchoredBatches: number;
  totalClaims: number;
  nodes: NodeStatus[];
  recentBatches?: BatchSummary[];
  recentClaims?: ClaimRecord[];
}

export interface BatchSummary {
  batchID: string;
  onChainID?: number;
  startBlock: number;
  endBlock: number;
  ruidCount: number;
  status: string;
  anchoredBy?: string;
  anchorBlock?: number;
  anchoredByName?: string;
}

export interface BatchDetail {
  batchID: string;
  onChainID?: number;
  startBlock: number;
  endBlock: number;
  rootHash: string;
  otsDigest?: string;
  ruidCount: number;
  createdAt: number;
  triggerType: string;
  status: string;
  btcTxID?: string;
  btcBlockHeight?: number;
  btcTimestamp?: number;
  anchoredBy?: string;
  anchorBlock?: number;
  anchoredByName?: string;
}

export interface RUIDsData {
  batchID: string;
  total: number;
  offset: number;
  limit: number;
  ruids: string[];
}

export interface OTSOperation {
  op: string;        // "append" | "prepend" | "sha256" | "ripemd160" | "keccak256" | "reverse"
  argument?: string; // hex for append/prepend
}

export interface OTSAttestation {
  type: string;           // "bitcoin" | "pending"
  btcBlockHeight?: number;
  calendarUrl?: string;
}

export interface ParsedOTSProof {
  digest: string;
  hashType: string;
  operations: OTSOperation[];
  attestations: OTSAttestation[];
}

export interface VerifyData {
  ruid: string;
  verified: boolean;
  batchID?: string;
  btcBlockHeight?: number;
  btcTimestamp?: number;
  message?: string;
  // Claim info from database
  published?: boolean;
  auid?: string;
  puid?: string;
  claimant?: string;
  submitBlock?: number;
  // Proof data
  rootHash?: string;
  otsDigest?: string;
  merkleProof?: string;
  otsProof?: string;
  parsedOTSProof?: ParsedOTSProof;
  leafIndex?: number;
  leafCount?: number;
}

export interface OTSProofData {
  batchID: string;
  rootHash: string;
  otsDigest?: string;
  otsProof?: string;
  hasProof: boolean;
  btcConfirmed: boolean;
  btcTxID?: string;
  btcBlockHeight?: number;
  btcTimestamp?: number;
}

export interface ClaimRecord {
  ruid: string;
  claimant?: string;
  submitBlock?: number;
  submitTime?: number;
  published: boolean;
  auid?: string;
  puid?: string;
  publishBlock?: number;
  publishTime?: number;
}

export interface ListData {
  items: ClaimRecord[];
  totalCount: number;
  offset: number;
  limit: number;
}

export interface ConflictData {
  auid: string;
  hasConflict: boolean;
  claimCount: number;
  ruids?: string[];
  earliest?: ClaimRecord;
}

export interface ConflictSummary {
  auid: string;
  puidCount: number;
  claimCount: number;
  earliestBlock: number;
  latestBlock: number;
}

export interface ConflictListData {
  items: ConflictSummary[] | null;
  total: number;
  offset: number;
  limit: number;
}

export interface ClaimStats {
  totalClaims: number;
  publishedCount: number;
  uniqueAuids: number;
  uniquePuids: number;
  conflictAuids: number;
}

export interface ChainConfig {
  chainId: number;
  chainName: string;
  breatheBlockInterval: number;
  nodeCount: number;
  nodes: string[];
}

export const fetchDashboard = (): Promise<DashboardData> => api.get('/dashboard');
export const fetchNodes = (): Promise<NodeStatus[]> => api.get('/nodes');
export const fetchNode = (name: string): Promise<NodeStatus> => api.get(`/nodes/${name}`);
export interface BatchListResponse {
  batches: BatchSummary[];
  total: number;
  page: number;
  pageSize: number;
}

export const fetchBatches = (params?: Record<string, string>): Promise<BatchListResponse> => api.get('/batches', { params });
export const fetchBatch = (id: string): Promise<BatchDetail> => api.get(`/batches/${id}`);
export const fetchBatchRUIDs = (id: string, offset = 0, limit = 100): Promise<RUIDsData> => api.get(`/batches/${id}/ruids`, { params: { offset, limit } });
export const fetchClaims = (params: Record<string, string>): Promise<ListData> => api.get('/claims', { params });
export const fetchConflict = (auid: string): Promise<ConflictData> => api.get(`/claims/conflicts/${auid}`);
export const verifyRUID = (ruid: string): Promise<VerifyData> => api.post('/verify', { ruid });
export const fetchProof = (batchId: string): Promise<OTSProofData> => api.get(`/proof/${batchId}`);
export const fetchConfig = (): Promise<ChainConfig> => api.get('/config');
export const fetchConflicts = (offset = 0, limit = 20): Promise<ConflictListData> => api.get('/conflicts', { params: { offset, limit } });
export const fetchClaimStats = (): Promise<ClaimStats> => api.get('/stats/claims');

export interface CalendarBatchInfo {
  batchID: string;
  status: string;
  calendarServer: string;
  attemptCount: number;
  lastAttemptAt?: string;
  lastError?: string;
  startBlock: number;
  endBlock: number;
  ruidCount: number;
}

export const fetchNodeCalendar = (name: string): Promise<CalendarBatchInfo[]> =>
  api.get(`/nodes/${name}/calendar`);

export interface CalendarURLStatus {
  url: string;
  priority: number;
  successCount: number;
  failureCount: number;
  lastAttemptAt?: string;
  lastError?: string;
  lastStatus: string;
}

export const fetchCalendarURLStatus = (name: string): Promise<CalendarURLStatus[]> =>
  api.get(`/nodes/${name}/calendar-url-status`);

export interface NodeHistoryPoint {
  blockNumber: number;
  pendingCount: number;
  status: string;
  recordedAt: string;
}

export const fetchNodeHistory = (name: string, limit = 360): Promise<NodeHistoryPoint[]> =>
  api.get(`/nodes/${name}/history`, { params: { limit } });

// Phase 2: New list types and API functions

export interface ClaimantSummary {
  claimant: string;
  claimCount: number;
  publishedCount: number;
  latestBlock: number;
}

export interface ClaimantListData {
  items: ClaimantSummary[] | null;
  total: number;
  offset: number;
  limit: number;
}

export interface AssetSummary {
  auid: string;
  claimCount: number;
  puidCount: number;
}

export interface AssetListData {
  items: AssetSummary[] | null;
  total: number;
  offset: number;
  limit: number;
}

export interface PersonSummary {
  puid: string;
  assetCount: number;
  claimCount: number;
}

export interface PersonListData {
  items: PersonSummary[] | null;
  total: number;
  offset: number;
  limit: number;
}

export const fetchClaimants = (offset = 0, limit = 20): Promise<ClaimantListData> =>
  api.get('/claimants', { params: { offset, limit } });

export const fetchAssets = (offset = 0, limit = 20): Promise<AssetListData> =>
  api.get('/assets', { params: { offset, limit } });

export const fetchPersons = (offset = 0, limit = 20): Promise<PersonListData> =>
  api.get('/persons', { params: { offset, limit } });
