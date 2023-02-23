package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type SnapshotServerTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
}

func (suite *SnapshotServerTestSuite) SetupSuite() {
	suite.envName = "snapshot-server-env"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)

	CreateEnv(global.Owner, suite.envName, "server", suite.T())
}

func (suite *SnapshotServerTestSuite) TestSnapshotServerCmd() {
	tests := []cmdTestCase{
		{
			wantError: false,
			name:      "snapshot server works when all required flags and argument are provided",
			cmd:       fmt.Sprintf(`snapshot server --paths testdata/file1 %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "[1] artifacts were reported to environment snapshot-server-env\n",
		},
		{
			wantError: true,
			name:      "snapshot server reports [0] if --paths not given",
			cmd:       fmt.Sprintf(`snapshot server %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag \"paths\" not set\n",
		},
		{
			wantError: true,
			name:      "snapshot server fails if two arguments are provided",
			cmd:       fmt.Sprintf(`snapshot server %s xxx %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
	}

	runTestCmd(suite.T(), tests)

}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSnapshotServerTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotServerTestSuite))
}
