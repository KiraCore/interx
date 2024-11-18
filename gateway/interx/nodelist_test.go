package interx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KiraCore/interx/tasks"
	"github.com/KiraCore/interx/test"
	"github.com/KiraCore/interx/types"
	"github.com/stretchr/testify/suite"
)

type NodeListTestSuite struct {
	suite.Suite
}

func (suite *NodeListTestSuite) SetupTest() {
}

func (suite *NodeListTestSuite) TestSnapListQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	q := r.URL.Query()
	q.Add("ip_only", "true")
	q.Add("order", "random")
	r.URL.RawQuery = q.Encode()

	tasks.SnapNodeListResponse = types.SnapNodeListResponse{
		NodeList: []types.SnapNode{
			{
				IP: "0.0.0.0", Port: 11111, Size: 10000, Checksum: "test_checksum", Alive: true, Synced: true,
			},
		},
	}

	response, _, statusCode := querySnapList(r, "")
	suite.Require().EqualValues(statusCode, http.StatusOK)
	suite.Require().EqualValues(response, "0.0.0.0")
}

func (suite *NodeListTestSuite) TestInterxListQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	q := r.URL.Query()
	q.Add("ip_only", "true")
	q.Add("order", "random")
	r.URL.RawQuery = q.Encode()

	tasks.InterxP2PNodeListResponse = types.InterxNodeListResponse{
		NodeList: []types.InterxNode{
			{
				IP: "0.0.0.0", ID: "test_id", Ping: 1111, Moniker: "test_moniker", Faucet: "test_faucet", Type: "test_type", InterxVersion: "0.1", SekaiVersion: "0.2", Alive: true, Synced: true, Safe: true, BlockHeightAtSync: 10, BlockDiff: 0,
			},
		},
	}

	response, _, statusCode := queryInterxList(r, "")
	suite.Require().EqualValues(statusCode, http.StatusOK)
	suite.Require().EqualValues(response, "0.0.0.0")
}

func (suite *NodeListTestSuite) TestPubP2PNodeListQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	q := r.URL.Query()
	q.Add("peers_only", "true")
	q.Add("order", "random")
	q.Add("format", "simple")
	q.Add("connected", "true")
	r.URL.RawQuery = q.Encode()

	tasks.PubP2PNodeListResponse = types.P2PNodeListResponse{
		NodeList: []types.P2PNode{
			{
				IP: "0.0.0.0", Port: 1111, Ping: 1111, Connected: true, Peers: []string{"0.0.0.1", "0.0.0.2"}, Alive: true, Synced: true, Safe: true, BlockHeightAtSync: 10, BlockDiff: 0,
			},
		},
	}

	response, _, statusCode := queryPubP2PNodeList(r, "")
	suite.Require().EqualValues(statusCode, http.StatusOK)
	suite.Require().EqualValues(response, "@0.0.0.0:1111")

	r = httptest.NewRequest("GET", test.INTERX_RPC, nil)
	q = r.URL.Query()
	q.Add("ip_only", "true")
	q.Add("order", "random")
	q.Add("format", "simple")
	q.Add("connected", "true")

	r.URL.RawQuery = q.Encode()
	response, _, statusCode = queryPubP2PNodeList(r, "")
	suite.Require().EqualValues(statusCode, http.StatusOK)
	suite.Require().EqualValues(response, "0.0.0.0")
}

func (suite *NodeListTestSuite) TestPrivP2PNodeListQuery() {
	r := httptest.NewRequest("GET", test.INTERX_RPC, nil)
	q := r.URL.Query()
	q.Add("peers_only", "true")
	q.Add("order", "random")
	q.Add("format", "simple")
	q.Add("behind", "0")
	q.Add("connected", "true")
	r.URL.RawQuery = q.Encode()

	tasks.PrivP2PNodeListResponse = types.P2PNodeListResponse{
		NodeList: []types.P2PNode{
			{
				IP: "0.0.0.0", Port: 1111, Ping: 1111, Connected: true, Peers: []string{"0.0.0.1", "0.0.0.2"}, Alive: true, Synced: true, Safe: true, BlockHeightAtSync: 10, BlockDiff: 0,
			},
		},
	}

	response, _, statusCode := queryPrivP2PNodeList(r, "")
	suite.Require().EqualValues(statusCode, http.StatusOK)
	suite.Require().EqualValues(response, "@0.0.0.0:1111")

	r = httptest.NewRequest("GET", test.INTERX_RPC, nil)
	q = r.URL.Query()
	q.Add("ip_only", "true")
	q.Add("order", "random")
	q.Add("format", "simple")
	q.Add("behind", "0")
	q.Add("connected", "true")

	r.URL.RawQuery = q.Encode()
	response, _, statusCode = queryPrivP2PNodeList(r, "")
	suite.Require().EqualValues(statusCode, http.StatusOK)
	suite.Require().EqualValues(response, "0.0.0.0")
}

func TestNodeListTestSuite(t *testing.T) {
	testSuite := new(NodeListTestSuite)
	suite.Run(t, testSuite)
}
