# Order Service Unit Tests

This directory contains comprehensive unit tests for the order-service microservice. The tests follow Go testing best practices with mocking for external dependencies.

## Test Structure

The test suite is organized into the following files:

### 1. **order_service_test.go** (7 tests)

Tests for the business logic layer (service package):

- `TestOrderServiceValidation` - Validates request validation for order creation
- `TestOrderServiceUpdateValidation` - Validates update request validation
- `TestOrderServiceUpdatePaidOrderValidation` - Ensures paid orders cannot be updated
- `TestOrderServiceDeleteValidation` - Tests delete business logic
- `TestOrderServiceFindByIdValidation` - Tests FindById logic
- `TestOrderServiceFindAllValidation` - Tests FindAll logic
- `TestOrderServiceProcessPaymentCallbackLogic` - Tests payment callback processing

**Key Features:**

- Uses mock repository to isolate service logic
- Tests business rules (e.g., paid orders cannot be updated)
- Tests validation requirements
- Uses struct `MockOrderRepository` to implement the `OrderRepository` interface

### 2. **order_controller_test.go** (8 tests)

Tests for the HTTP controller layer:

- `TestOrderControllerCreate` - Tests POST /orders endpoint
- `TestOrderControllerFindById` - Tests GET /orders/:orderId endpoint
- `TestOrderControllerFindAll` - Tests GET /orders endpoint
- `TestOrderControllerUpdate` - Tests PUT /orders/:orderId endpoint
- `TestOrderControllerDelete` - Tests DELETE /orders/:orderId endpoint
- `TestOrderControllerFindByIdInvalidUUID` - Tests invalid UUID handling

**Key Features:**

- Tests HTTP endpoints with mocked service layer
- Uses Fiber test client for request/response testing
- Tests error handling and validation

### 3. **order_repository_test.go** (8 tests)

Tests for the data access layer:

- `TestOrderRepositorySave` - Tests save operation
- `TestOrderRepositoryFindById` - Tests find by ID
- `TestOrderRepositoryFindByIdNotFound` - Tests not found scenarios
- `TestOrderRepositoryFindByAll` - Tests finding all orders
- `TestOrderRepositoryUpdate` - Tests update operations
- `TestOrderRepositoryUpdateWithPaymentId` - Tests updating with payment ID
- `TestOrderRepositoryDelete` - Tests soft delete

**Key Features:**

- Tests core repository operations
- Validates order state transitions
- Verifies PaymentID handling

### 4. **order_helper_test.go** (7 tests)

Tests for utility/helper functions:

- `TestPanicIfError` - Tests error panic helper
- `TestErrorResponse` - Tests error response formatting
- `TestToOrderResponse` - Tests domain to response conversion
- `TestToOrderResponses` - Tests converting multiple orders
- `TestReadFromRequestBody` - Tests request body parsing

**Key Features:**

- Tests utility functions
- Validates type conversions
- Tests response formatting

### 5. **order_config_test.go** (2 tests)

Tests for database configuration:

- `TestNewDBWithMissingEnvVars` - Tests handling of missing env vars
- `TestDBEnvironmentConfiguration` - Tests environment variable setup

### 6. **order_exception_test.go** (3 tests)

Tests for custom exception types:

- `TestNotFoundError` - Tests NotFoundError type
- `TestNotFoundErrorInterface` - Tests error interface implementation
- `TestNotFoundErrorEmpty` - Tests error with empty message

## Running the Tests

### Run all tests in the test folder:

```bash
go test ./test -v
```

### Run tests with coverage:

```bash
go test ./test -v -cover
```

### Run specific test:

```bash
go test ./test -run TestOrderServiceValidation -v
```

### Run with short timeout:

```bash
go test ./test -short -v
```

## Test Results

All tests pass successfully (30 total test cases):

```
PASS    order-service/test      0.010s
```

## Mocking Strategy

The tests use the `github.com/stretchr/testify/mock` package to create mock implementations:

- **MockOrderService** - Used in controller tests to isolate HTTP layer
- **MockOrderRepository** - Used in service tests to isolate business logic

This isolation ensures:

- Tests are fast (no DB calls)
- Tests are independent
- Errors are traceable to specific layers
- Business logic can be tested without infrastructure

## Test Coverage Areas

1. **Validation**: All inputs are validated
2. **Business Logic**: Rules like "paid orders cannot be updated" are enforced
3. **Error Handling**: Invalid UUIDs, missing orders, etc.
4. **State Transitions**: Orders moving through states (pending → paid)
5. **Helper Functions**: Conversion and formatting utilities
6. **Configuration**: Environment variable handling

## Integration Testing Notes

For full end-to-end integration testing:

- Use a test database (PostgreSQL in Docker)
- Create integration tests that use actual repository implementations
- Test payment callback flow across services
- Test transaction handling and rollbacks

## Dependencies

Test dependencies are declared in `go.mod`:

- `github.com/stretchr/testify` - Assertions and mocking
- `github.com/gofiber/fiber/v2` - Web framework (for controller tests)
- Standard library packages: `context`, `testing`, `encoding/json`, etc.

## Best Practices Applied

✓ Each test is independent and can run in any order
✓ Mock objects isolate the layer under test
✓ Descriptive test names indicate what is being tested
✓ Assertions are clear and specific
✓ No external dependencies in unit tests
✓ Tests verify both happy path and error cases
