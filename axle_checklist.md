# Axle Improvement Checklist

## Critical Fixes (Do First)

- [ ] Fix memory leak: Add cleanup routine for `lastEventTime` map in watcher.go
- [ ] Add patch validation: Check for path traversal attacks before applying patches
- [x] Fix race conditions: Add sequence numbers to SyncMetadata for patch ordering
- [ ] Add Redis connection retry logic with exponential backoff
- [ ] Implement graceful shutdown: Flush pending changes before exit

## Performance Improvements

- [x] Batch git operations: Collect changes for 2 seconds before committing
- [ ] Add Redis connection pooling
- [ ] Optimize file watching: Skip temporary/swap files (.tmp, .swp, ~)
- [ ] Add file size limits to prevent syncing huge files
- [ ] Implement delta sync for large files

## Security Enhancements

- [ ] Add authentication: JWT tokens or API keys for team access
- [ ] Validate incoming patches: Check format and dangerous operations
- [ ] Add rate limiting: Prevent spam/DOS from malicious peers
- [ ] Encrypt sensitive data in Redis
- [ ] Add peer verification: Digital signatures for patches

## Reliability Features

- [ ] Add health checks: Monitor Redis, Git, and file system
- [ ] Implement conflict resolution: Handle concurrent file modifications
- [ ] Add rollback capability: Undo problematic patches
- [ ] Create backup mechanism: Periodic snapshots of sync state
- [ ] Add network partition handling: Queue changes during disconnection

## User Experience

- [ ] Add progress indicators for large sync operations
- [x] Implement file exclusion patterns (like .gitignore)
- [ ] Add sync status dashboard/CLI command
- [x] Create better logging with levels (DEBUG, INFO, ERROR)
- [ ] Add configuration validation on startup

## Testing & Monitoring

- [ ] Write unit tests for core functions (patch apply, git operations)
- [ ] Add integration tests for multi-peer scenarios
- [ ] Implement metrics collection (files synced, errors, performance)
- [ ] Add performance benchmarks
- [ ] Create end-to-end testing suite

## Documentation & DevOps

- [ ] Add comprehensive README with setup instructions
- [ ] Create Docker containers for easy deployment
- [ ] Add CI/CD pipeline with automated testing
- [ ] Write API documentation
- [ ] Add troubleshooting guide
