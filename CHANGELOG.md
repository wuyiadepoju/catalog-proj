# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased] - 2026-02-25

### Improvements

**Error Handling**
- Fixed silent failures in event-to-outbox mutation functions; errors are now properly propagated
- Applied across all use cases: create_product, update_product, apply_discount, activate_product, deactivate_product, archive_product, remove_discount

**Test Reliability**
- Added context timeouts in test setup to prevent hanging when Spanner emulator isn't running
- Fixed context cancellation issues that caused "context canceled" errors
- Better error messages with actionable guidance

**Input Validation**
- Prices: Must be positive, non-zero
- Strings: Trimmed and validated for length limits (name: 255, description: 1000, category: 100)
- Discounts: ID validation, amount range 0-100%, date range validation
- Updates: At least one field must be provided

**Precision**
- Fixed money conversion precision issues using exact arithmetic when possible
- Prevents precision loss in money calculations

**Domain Validation**
- Added domain-level validation in Product and Discount entities
- New error types: `ErrInvalidProductName`, `ErrInvalidProductDescription`, `ErrInvalidProductCategory`, `ErrInvalidDiscountID`, `ErrInvalidDiscountAmount`, `ErrInvalidDiscountDateRange`
- Updated gRPC error mapping for all new domain errors

**Documentation**
- Updated discount examples: dates from 2026-02-25T00:00:00Z onward, amounts 0-100 range
- Clarified product activation requirement before applying discounts
- Improved API usage examples
