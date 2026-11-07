package integration

import (
	"os"
	"testing"

	"github.com/crcaniullan-commits/Tally/tests/integration/testutil"
)

func TestMain(m *testing.M) {
	os.Exit(testutil.SetupSuite(m))
}
