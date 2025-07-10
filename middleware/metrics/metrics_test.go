package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestInitMetrics(t *testing.T) {
	// Verify HTTPDuration collector
	count := testutil.CollectAndCount(HTTPDuration)
	assert.Greater(t, count, 0, "HTTPDuration should be registered and collectable")

	// Verify GRPCDuration collector
	count = testutil.CollectAndCount(GRPCDuration)
	assert.Greater(t, count, 0, "GRPCDuration should be registered and collectable")
}
