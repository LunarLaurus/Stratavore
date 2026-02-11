# Stratavore TODO - Remaining Features & Improvements

**Last Updated:** February 11, 2026  
**Current Version:** v1.4.0 (98% complete)  
**Target Version:** v2.0 (100%)

---

## 🎯 High Priority (Required for 100%)

### 1. Redis Caching Layer ⏳
**Status:** Ready to implement  
**Effort:** 2-3 hours  
**Impact:** HIGH - 10x faster queries

**Tasks:**
- [x] Create Redis client wrapper
- [x] Implement cache-aside pattern for projects
- [x] Cache runner lists with TTL
- [x] Cache-through for frequently accessed data
- [x] Add cache invalidation on updates
- [x] Implement cache warming on startup

**Files to create:**
- ✅ `internal/cache/redis.go`
- ✅ `internal/cache/manager.go`

**Benefits:**
- Sub-millisecond project queries
- Reduced database load
- Better scalability

---

### 2. Load Testing & Benchmarks ⏳
**Status:** Framework ready, needs execution  
**Effort:** 3-4 hours  
**Impact:** HIGH - Validates production readiness

**Tasks:**
- [x] Create load test scenarios (`test/load/scenarios_test.go`)
- [x] Test 100 concurrent launches — `TestConcurrentLaunches`
- [x] Test heartbeat throughput — `TestHeartbeatThroughput`
- [x] Test status under load — `TestStatusUnderLoad`
- [x] Mixed workload scenario — `TestMixedWorkload`
- [ ] Test 1000 active runners (requires live daemon)
- [ ] Test 10k events/second (requires live daemon)
- [ ] Measure database connection pool
- [ ] Profile memory usage
- [ ] Profile CPU usage
- [ ] Document bottlenecks

**Files to create:**
- `test/load/scenarios.go`
- `test/load/runner_launch_test.go`
- `test/load/heartbeat_test.go`
- `docs/PERFORMANCE.md`

---

### 3. Security Hardening ⏳
**Status:** Foundation ready  
**Effort:** 4-5 hours  
**Impact:** CRITICAL - Production security

**Tasks:**
- [x] Implement API authentication (JWT/API keys) — `internal/auth/jwt.go`
- [x] Add rate limiting per client — `internal/auth/ratelimit.go`
- [x] Wire auth middleware into HTTP server (`internal/daemon/http_server.go`)
- [ ] Enable mTLS for internal communication
- [ ] Implement agent token validation
- [x] Add request signing (HMAC) — `internal/auth/hmac.go`
- [ ] SQL injection prevention audit
- [ ] Input sanitization review
- [ ] Secrets management (Vault integration)
- [ ] Security audit documentation

**Files to create:**
- ✅ `internal/auth/jwt.go`
- ✅ `internal/auth/ratelimit.go`
- ✅ `internal/auth/hmac.go`
- [ ] `docs/SECURITY.md`

---

### 4. Grafana Dashboards 🎯 EASY WIN
**Status:** Infrastructure ready  
**Effort:** 1-2 hours  
**Impact:** MEDIUM - Better visibility

**Tasks:**
- [x] Create main dashboard JSON (overview — modernised)
- [x] Add runner metrics panel (runner-metrics dashboard)
- [x] Add token usage panel
- [x] Add event rate panel
- [x] Add error rate panel
- [x] Add latency histograms (p50/p95/p99)
- [x] Add project breakdown
- [x] Configure auto-provisioning

**Files to create:**
- ✅ `configs/grafana/dashboards/stratavore-overview.json`
- ✅ `configs/grafana/dashboards/runner-metrics.json`
- ✅ `configs/grafana/datasources/prometheus.yml`

---

## 🚀 Medium Priority (Nice to Have)

### 5. S3 Transcript Storage ⏳
**Status:** Metadata ready, needs implementation  
**Effort:** 3-4 hours  
**Impact:** MEDIUM - Session persistence

**Tasks:**
- [ ] Add AWS S3 client
- [ ] Implement transcript upload on session end
- [ ] Add transcript download for resumption
- [ ] Implement lifecycle policies
- [ ] Add compression before upload
- [ ] Handle upload failures gracefully
- [ ] Add presigned URL generation
- [ ] Test with MinIO (S3-compatible)

**Files to create:**
- `internal/storage/s3.go`
- `internal/storage/transcript.go`

---

### 6. Vector Embeddings (Qdrant) ⏳
**Status:** Qdrant in docker-compose  
**Effort:** 4-5 hours  
**Impact:** MEDIUM - Session similarity

**Tasks:**
- [ ] Add Qdrant client
- [ ] Generate session embeddings
- [ ] Store embeddings on session end
- [ ] Implement similarity search
- [ ] Add "find similar sessions" CLI command
- [ ] Create embedding pipeline
- [ ] Test with actual sessions

**Files to create:**
- `internal/embeddings/qdrant.go`
- `internal/embeddings/generator.go`

---

### 7. Advanced CLI Commands ⏳
**Status:** Foundation ready  
**Effort:** 2-3 hours  
**Impact:** MEDIUM - Better UX

**Tasks:**
- [ ] Add `stratavore logs <runner-id>` - Stream runner logs
- [ ] Add `stratavore attach <runner-id>` - Attach to runner (PTY)
- [ ] Add `stratavore budget create` - Create budgets via CLI
- [ ] Add `stratavore budget show` - Display budget status
- [ ] Add `stratavore sessions` - List sessions
- [ ] Add `stratavore resume <session-id>` - Resume session
- [ ] Add `stratavore export <runner-id>` - Export runner data
- [ ] Add autocomplete support

**Files to update:**
- `cmd/stratavore/main.go`

---

### 8. Web UI (React Dashboard) ⏳
**Status:** API ready for consumption  
**Effort:** 8-12 hours  
**Impact:** MEDIUM - Alternative interface

**Tasks:**
- [ ] Create React app structure
- [ ] Add dashboard page
- [ ] Add project management
- [ ] Add runner management
- [ ] Add real-time updates (WebSocket)
- [ ] Add budget visualization
- [ ] Add metrics charts
- [ ] Add session browser

**Files to create:**
- `web/` directory structure
- API integration layer

---

## 📊 Low Priority (Future Enhancements)

### 9. Multi-Node Support (Remote Runners) ⏳
**Status:** Architecture supports it  
**Effort:** 10-15 hours  
**Impact:** LOW - Multi-datacenter

**Tasks:**
- [ ] Add node registration
- [ ] Implement node discovery
- [ ] Add cross-node communication
- [ ] Implement runner migration
- [ ] Add node health monitoring
- [ ] Implement load balancing

---

### 10. Auto-Scaling ⏳
**Status:** Metrics available  
**Effort:** 6-8 hours  
**Impact:** LOW - Dynamic capacity

**Tasks:**
- [ ] Add scaling policies
- [ ] Implement autoscaler
- [ ] Add CPU-based scaling
- [ ] Add memory-based scaling
- [ ] Add token-budget-based scaling

---

### 11. Plugin System ⏳
**Status:** Not started  
**Effort:** 8-10 hours  
**Impact:** LOW - Extensibility

**Tasks:**
- [ ] Define plugin interface
- [ ] Add plugin loader
- [ ] Create example plugins
- [ ] Add plugin marketplace

---

## 🐛 Bug Fixes & Improvements

### 12. Known Issues
- [x] Agent collects real CPU/memory via `internal/procmetrics` (Linux `/proc`, macOS `ps`)
- [ ] Session transcript download not implemented
- [ ] No actual Claude Code token parsing
- [ ] Missing cleanup on daemon crash recovery

### 13. Code Quality Improvements
- [ ] Add unit tests (currently only integration)
- [ ] Increase test coverage to 80%+
- [ ] Add golangci-lint to CI/CD
- [ ] Add pre-commit hooks
- [ ] Improve error messages
- [ ] Add more detailed logging

---

## 📚 Documentation Needs

### 14. Additional Documentation
- [ ] API Reference (OpenAPI/Swagger)
- [ ] Deployment guide for Kubernetes
- [ ] Troubleshooting guide
- [ ] Performance tuning guide
- [ ] Migration guide (upgrades)
- [ ] Backup and recovery guide
- [ ] Disaster recovery plan

---

## 🎯 Quick Wins (Can Do Now)

### Priority Quick Wins
1. ✅ **Redis Caching** - 2 hours, 10x speedup
2. ✅ **Grafana Dashboards** - 1 hour, immediate visibility
3. ✅ **CLI autocomplete** - bash/zsh/fish/powershell via `stratavore completion`
4. ⏳ **API documentation** - 1 hour, better developer experience
5. ✅ **Process metrics collection** - real CPU/memory via `/proc` + `ps` fallback

---

## 📈 Effort vs Impact Matrix

```
High Impact, Low Effort (DO FIRST):
✅ Redis caching (2h)
✅ Grafana dashboards (1h)
- CLI autocomplete (30m)
- Process metrics (30m)

High Impact, High Effort:
- Load testing (4h)
- Security hardening (5h)
- S3 storage (4h)

Medium Impact, Low Effort:
- Advanced CLI commands (3h)
- API docs (1h)

Low Priority:
- Web UI (12h)
- Multi-node (15h)
- Auto-scaling (8h)
```

---

## 🎯 Roadmap to 100%

### Phase 6 (Target: 97%)
- [ ] Redis caching
- [ ] Grafana dashboards
- [ ] Process metrics collection
**Estimated:** 3 hours

### Phase 7 (Target: 99%)
- [ ] Load testing
- [ ] Security hardening
- [ ] API documentation
**Estimated:** 8 hours

### Phase 8 (Target: 100%)
- [ ] S3 transcript storage
- [ ] Advanced CLI commands
- [ ] Production deployment validation
**Estimated:** 6 hours

**Total to 100%:** ~17 hours of focused work

---

## ✅ Completed (95%)

**Phase 1-5 Complete:**
- ✅ Core infrastructure
- ✅ HTTP API
- ✅ CLI integration
- ✅ Telegram notifications
- ✅ Token budgets
- ✅ Session management
- ✅ Integration tests
- ✅ Docker Compose
- ✅ Prometheus metrics
- ✅ Agent heartbeats

---

## 📞 Next Steps

**Immediate (Today):**
1. Implement Redis caching
2. Create Grafana dashboards
3. Release v1.1

**This Week:**
1. Load testing
2. Security hardening
3. S3 storage

**This Month:**
1. Advanced CLI
2. Vector embeddings
3. Web UI (optional)

---

**Current Status:** 95% → Target: 100%  
**Remaining Work:** ~17 hours  
**Priority:** Quick wins first (Redis + Grafana)
