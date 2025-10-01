# Axle Improvement Checklist

## Critical Fixes (Do First)

- [x] Fix memory leak: Add cleanup routine for `lastEventTime` map in watcher.go
- [x] Add patch validation: Check for path traversal attacks before applying patches
- [x] Fix race conditions: Add sequence numbers to SyncMetadata for patch ordering
- [x] Add Redis connection retry logic with exponential backoff
- [x] Implement graceful shutdown: Flush pending changes before exit

## Performance Improvements

- [x] Batch git operations: Collect changes for 2 seconds before committing
- [ ] Add Redis connection pooling
- [x] Optimize file watching: Skip temporary/swap files (.tmp, .swp, ~)
- [x] Add file size limits to prevent syncing huge files
- [x] Add dynamic batch window based on activity (1-5 seconds)
- [ ] Implement delta sync for large files

## Security Enhancements

- [x] Add team password protection with `init` and `join` commands
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

- [x] Persist Node-ID to prevent duplicate users in team status
- [ ] Add progress indicators for large sync operations
- [x] Implement file exclusion patterns (like .gitignore)
- [x] Add sync status dashboard/CLI command (stats command)
- [x] Smart .gitignore auto-detection based on tech stack
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

## Future Features

- [ ] Implement "force sync" for Git-ignored files: Add an `includePatterns` option to sync files that are in `.gitignore` (e.g., `.env` files). This would require a custom, non-Git patching mechanism.
