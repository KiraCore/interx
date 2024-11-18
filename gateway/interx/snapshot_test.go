package interx

import (
	"net/http"
	"testing"

	"github.com/KiraCore/interx/tasks"
	"github.com/KiraCore/interx/test"
	"github.com/KiraCore/interx/types"
	"github.com/stretchr/testify/suite"
)

type SnapshotTestSuite struct {
	suite.Suite
}

func (suite *SnapshotTestSuite) SetupTest() {
}

func (suite *SnapshotTestSuite) TestSnapInfoQuery() {
	tasks.SnapshotChecksumAvailable = true
	tasks.SnapshotChecksum = "test_checksum"
	tasks.SnapshotLength = 100
	response, _, statusCode := querySnapShotInfoHandler(test.TENDERMINT_RPC)

	suite.Require().EqualValues(response.(*types.SnapShotChecksumResponse).Checksum, tasks.SnapshotChecksum)
	suite.Require().EqualValues(statusCode, http.StatusOK)
}

func TestSnapshotTestSuite(t *testing.T) {
	testSuite := new(SnapshotTestSuite)
	suite.Run(t, testSuite)
}
