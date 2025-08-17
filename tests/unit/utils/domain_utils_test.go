package utils_test

import (
	"gn-indexer/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewU64(t *testing.T) {
	// Test creating U64 from int64
	value := int64(12345)
	u64 := domain.NewU64(value)

	assert.Equal(t, value, u64.Int64())
	assert.Equal(t, "12345", u64.String())
}

func TestNewU64FromString(t *testing.T) {
	// Test creating U64 from string
	u64, err := domain.NewU64FromString("12345")
	assert.NoError(t, err)
	assert.Equal(t, int64(12345), u64.Int64())

	// Test invalid string
	_, err = domain.NewU64FromString("invalid")
	assert.Error(t, err)
}

func TestU64_Int64(t *testing.T) {
	// Test Int64 conversion
	u64 := domain.NewU64(100)
	assert.Equal(t, int64(100), u64.Int64())

	// Test zero value
	zero := &domain.U64{}
	assert.Equal(t, int64(0), zero.Int64())
}

func TestU64_String(t *testing.T) {
	// Test U64 string representation
	u64 := domain.NewU64(12345)
	assert.Equal(t, "12345", u64.String())

	// Test zero value
	zero := &domain.U64{}
	assert.Equal(t, "0", zero.String())
}

func TestU64_Value(t *testing.T) {
	// Test Value method for database serialization
	u64 := domain.NewU64(12345)
	value, err := u64.Value()
	assert.NoError(t, err)
	assert.Equal(t, "12345", value)

	// Test nil value
	zero := &domain.U64{}
	value, err = zero.Value()
	assert.NoError(t, err)
	assert.Nil(t, value)
}

func TestU64_Scan(t *testing.T) {
	// Test Scan method for database deserialization
	u64 := &domain.U64{}

	// Test scanning from string
	err := u64.Scan("12345")
	assert.NoError(t, err)
	assert.Equal(t, int64(12345), u64.Int64())

	// Test scanning from bytes
	err = u64.Scan([]byte("67890"))
	assert.NoError(t, err)
	assert.Equal(t, int64(67890), u64.Int64())

	// Test scanning nil
	err = u64.Scan(nil)
	assert.NoError(t, err)
	assert.Nil(t, u64.Int)
}
