// Package credential implements the denormalized dual-write pattern for credentials.
//
// Unlike the simple dual-write pattern (used by todo and user_profiles) where primary
// and secondary repositories implement the same interface, this package writes a
// denormalized document to MongoDB as the secondary store. The denormalized document
// embeds related data (vault, category, tags, user profile) alongside the credential
// fields, enabling efficient reads without joins.
//
// Write semantics follow fire-and-forget for the secondary:
//   - Primary write (PostgreSQL) must succeed for the operation to succeed.
//   - Secondary write (MongoDB denormalized) failures are logged but do not fail the operation.
//   - If the primary write fails, the secondary write is never attempted.
package credential
