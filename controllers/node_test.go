package controller

import (
	"testing"

	"github.com/gravitl/netmaker/database"
	"github.com/gravitl/netmaker/logic"
	"github.com/gravitl/netmaker/logic/acls"
	"github.com/gravitl/netmaker/logic/acls/nodeacls"
	"github.com/gravitl/netmaker/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateEgressGateway(t *testing.T) {
	var gateway models.EgressGatewayRequest
	gateway.Interface = "eth0"
	gateway.Ranges = []string{"10.100.100.0/24"}
	gateway.NetID = "skynet"
	database.InitializeDatabase()
	deleteAllNetworks()
	createNet()
	t.Run("NoNodes", func(t *testing.T) {
		node, err := logic.CreateEgressGateway(gateway)
		assert.Equal(t, models.Node{}, node)
		assert.EqualError(t, err, "could not find any records")
	})
	t.Run("Non-linux node", func(t *testing.T) {
		createnode := models.Node{PublicKey: "DM5qhLAE20PG9BbfBCger+Ac9D2NDOwCtY1rbYDLf34=", Name: "testnode", Endpoint: "10.0.0.1", MacAddress: "01:02:03:04:05:06", Password: "password", Network: "skynet", OS: "windows"}
		err := logic.CreateNode(&createnode)
		assert.Nil(t, err)
		gateway.NodeID = createnode.ID
		node, err := logic.CreateEgressGateway(gateway)
		assert.Equal(t, models.Node{}, node)
		assert.EqualError(t, err, "windows is unsupported for egress gateways")
	})
	t.Run("Success-Nat-Enabled", func(t *testing.T) {
		deleteAllNodes()
		testnode := createTestNode()
		gateway.NodeID = testnode.ID
		gateway.NatEnabled = "yes"

		node, err := logic.CreateEgressGateway(gateway)
		t.Log(node.EgressGatewayNatEnabled)
		t.Log(node.PostUp)
		t.Log(node.PostDown)
		assert.Nil(t, err)
		assert.Contains(t, node.PostUp, "-j MASQUERADE")
		assert.Contains(t, node.PostDown, "-j MASQUERADE")
	})
	t.Run("Success-Nat-Disabled", func(t *testing.T) {
		deleteAllNodes()
		testnode := createTestNode()
		gateway.NodeID = testnode.ID
		gateway.NatEnabled = "no"

		node, err := logic.CreateEgressGateway(gateway)
		t.Log(node.EgressGatewayNatEnabled)
		t.Log(node.PostUp)
		t.Log(node.PostDown)
		assert.Nil(t, err)
		assert.NotContains(t, node.PostUp, "-j MASUERADE")
		assert.NotContains(t, node.PostDown, "-j MASUERADE")
	})
	t.Run("Success", func(t *testing.T) {
		var gateway models.EgressGatewayRequest
		gateway.Interface = "eth0"
		gateway.Ranges = []string{"10.100.100.0/24"}
		gateway.NetID = "skynet"
		deleteAllNodes()
		testnode := createTestNode()
		gateway.NodeID = testnode.ID

		node, err := logic.CreateEgressGateway(gateway)
		t.Log(node)
		assert.Nil(t, err)
		assert.Contains(t, node.PostUp, "-j MASQUERADE")
		assert.Contains(t, node.PostDown, "-j MASQUERADE")
		assert.Equal(t, "yes", node.IsEgressGateway)
		assert.Equal(t, gateway.Ranges, node.EgressGatewayRanges)
	})

}
func TestDeleteEgressGateway(t *testing.T) {
	var gateway models.EgressGatewayRequest
	database.InitializeDatabase()
	deleteAllNetworks()
	createNet()
	testnode := createTestNode()
	gateway.Interface = "eth0"
	gateway.Ranges = []string{"10.100.100.0/24"}
	gateway.NetID = "skynet"
	gateway.NodeID = testnode.ID
	t.Run("Success", func(t *testing.T) {
		node, err := logic.CreateEgressGateway(gateway)
		assert.Nil(t, err)
		assert.Equal(t, "yes", node.IsEgressGateway)
		assert.Equal(t, []string{"10.100.100.0/24"}, node.EgressGatewayRanges)
		node, err = logic.DeleteEgressGateway(gateway.NetID, gateway.NodeID)
		assert.Nil(t, err)
		assert.Equal(t, "no", node.IsEgressGateway)
		assert.Equal(t, []string([]string{}), node.EgressGatewayRanges)
		assert.Equal(t, "", node.PostUp)
		assert.Equal(t, "", node.PostDown)
	})
	t.Run("NotGateway", func(t *testing.T) {
		node, err := logic.DeleteEgressGateway(gateway.NetID, gateway.NodeID)
		assert.Nil(t, err)
		assert.Equal(t, "no", node.IsEgressGateway)
		assert.Equal(t, []string([]string{}), node.EgressGatewayRanges)
		assert.Equal(t, "", node.PostUp)
		assert.Equal(t, "", node.PostDown)
	})
	t.Run("BadNode", func(t *testing.T) {
		node, err := logic.DeleteEgressGateway(gateway.NetID, "01:02:03")
		assert.EqualError(t, err, "no result found")
		assert.Equal(t, models.Node{}, node)
		deleteAllNodes()
	})
}

func TestGetNetworkNodes(t *testing.T) {
	database.InitializeDatabase()
	deleteAllNetworks()
	createNet()
	t.Run("BadNet", func(t *testing.T) {
		node, err := logic.GetNetworkNodes("badnet")
		assert.Nil(t, err)
		assert.Nil(t, node)
	})
	t.Run("NoNodes", func(t *testing.T) {
		node, err := logic.GetNetworkNodes("skynet")
		assert.Nil(t, err)
		assert.Nil(t, node)
	})
	t.Run("Success", func(t *testing.T) {
		createTestNode()
		node, err := logic.GetNetworkNodes("skynet")
		assert.Nil(t, err)
		assert.NotEqual(t, []models.Node(nil), node)
	})

}
func TestUncordonNode(t *testing.T) {
	database.InitializeDatabase()
	deleteAllNetworks()
	createNet()
	node := createTestNode()
	t.Run("BadID", func(t *testing.T) {
		resp, err := logic.UncordonNode("blahblah")
		assert.Equal(t, models.Node{}, resp)
		assert.EqualError(t, err, "no result found")
	})
	t.Run("Success", func(t *testing.T) {
		resp, err := logic.UncordonNode(node.ID)
		assert.Nil(t, err)
		assert.Equal(t, "no", resp.IsPending)
	})

}
func TestValidateEgressGateway(t *testing.T) {
	var gateway models.EgressGatewayRequest
	t.Run("EmptyRange", func(t *testing.T) {
		gateway.Interface = "eth0"
		gateway.Ranges = []string{}
		err := logic.ValidateEgressGateway(gateway)
		assert.EqualError(t, err, "IP Ranges Cannot Be Empty")
	})
	t.Run("EmptyInterface", func(t *testing.T) {
		gateway.Interface = ""
		err := logic.ValidateEgressGateway(gateway)
		assert.NotNil(t, err)
		assert.Equal(t, "interface cannot be empty", err.Error())
	})
	t.Run("Success", func(t *testing.T) {
		gateway.Interface = "eth0"
		gateway.Ranges = []string{"10.100.100.0/24"}
		err := logic.ValidateEgressGateway(gateway)
		assert.Nil(t, err)
	})
}

func TestNodeACLs(t *testing.T) {
	deleteAllNodes()
	node1 := models.Node{PublicKey: "DM5qhLAE20PG9BbfBCger+Ac9D2NDOwCtY1rbYDLf34=", Name: "testnode", Endpoint: "10.0.0.50", MacAddress: "01:02:03:04:05:06", Password: "password", Network: "skynet", OS: "linux"}
	node2 := models.Node{PublicKey: "DM5qhLAE20FG7BbfBCger+Ac9D2NDOwCtY1rbYDXf14=", Name: "testnode", Endpoint: "10.0.0.100", MacAddress: "01:02:03:04:05:07", Password: "password", Network: "skynet", OS: "linux"}
	logic.CreateNode(&node1)
	logic.CreateNode(&node2)
	t.Run("acls not present", func(t *testing.T) {
		currentACL, err := nodeacls.FetchAllACLs(nodeacls.NetworkID(node1.Network))
		assert.Nil(t, err)
		assert.NotNil(t, currentACL)
		node1ACL, err := nodeacls.FetchNodeACL(nodeacls.NetworkID(node1.Network), nodeacls.NodeID(node1.ID))
		assert.Nil(t, err)
		assert.NotNil(t, node1ACL)
		assert.Equal(t, acls.Allowed, node1ACL[acls.AclID(node2.ID)])
	})
	t.Run("node acls exists after creates", func(t *testing.T) {
		node1ACL, err := nodeacls.FetchNodeACL(nodeacls.NetworkID(node1.Network), nodeacls.NodeID(node1.ID))
		assert.Nil(t, err)
		assert.NotNil(t, node1ACL)
		node2ACL, err := nodeacls.FetchNodeACL(nodeacls.NetworkID(node2.Network), nodeacls.NodeID(node2.ID))
		assert.Nil(t, err)
		assert.NotNil(t, node2ACL)
		assert.Equal(t, acls.Allowed, node2ACL[acls.AclID(node1.ID)])
	})
	t.Run("node acls correct after fetch", func(t *testing.T) {
		node1ACL, err := nodeacls.FetchNodeACL(nodeacls.NetworkID(node1.Network), nodeacls.NodeID(node1.ID))
		assert.Nil(t, err)
		assert.Equal(t, acls.Allowed, node1ACL[acls.AclID(node2.ID)])
	})
	t.Run("node acls correct after modify", func(t *testing.T) {
		node1ACL, err := nodeacls.FetchNodeACL(nodeacls.NetworkID(node1.Network), nodeacls.NodeID(node1.ID))
		assert.Nil(t, err)
		assert.NotNil(t, node1ACL)
		node2ACL, err := nodeacls.FetchNodeACL(nodeacls.NetworkID(node2.Network), nodeacls.NodeID(node2.ID))
		assert.Nil(t, err)
		assert.NotNil(t, node2ACL)
		currentACL, err := nodeacls.DisallowNodes(nodeacls.NetworkID(node1.Network), nodeacls.NodeID(node1.ID), nodeacls.NodeID(node2.ID))
		assert.Nil(t, err)
		assert.Equal(t, acls.NotAllowed, currentACL[acls.AclID(node1.ID)][acls.AclID(node2.ID)])
		assert.Equal(t, acls.NotAllowed, currentACL[acls.AclID(node2.ID)][acls.AclID(node1.ID)])
		currentACL.Save(acls.ContainerID(node1.Network))
	})
	t.Run("node acls correct after add new node not allowed", func(t *testing.T) {
		node3 := models.Node{PublicKey: "this-is-not-valid", Name: "testnode3", Endpoint: "10.0.0.100", MacAddress: "01:02:03:04:05:07", Password: "password", Network: "skynet", OS: "linux"}
		logic.CreateNode(&node3)
		var currentACL, err = nodeacls.FetchAllACLs(nodeacls.NetworkID(node3.Network))
		assert.Nil(t, err)
		assert.NotNil(t, currentACL)
		assert.Equal(t, acls.NotPresent, currentACL[acls.AclID(node1.ID)][acls.AclID(node3.ID)])
		nodeACL, err := nodeacls.CreateNodeACL(nodeacls.NetworkID(node3.Network), nodeacls.NodeID(node3.ID), acls.NotAllowed)
		assert.Nil(t, err)
		nodeACL.Save(acls.ContainerID(node3.Network), acls.AclID(node3.ID))
		currentACL, err = nodeacls.FetchAllACLs(nodeacls.NetworkID(node3.Network))
		assert.Nil(t, err)
		assert.Equal(t, acls.NotAllowed, currentACL[acls.AclID(node1.ID)][acls.AclID(node3.ID)])
		assert.Equal(t, acls.NotAllowed, currentACL[acls.AclID(node2.ID)][acls.AclID(node3.ID)])
	})
	t.Run("node acls removed", func(t *testing.T) {
		retNetworkACL, err := nodeacls.RemoveNodeACL(nodeacls.NetworkID(node1.Network), nodeacls.NodeID(node1.ID))
		assert.Nil(t, err)
		assert.NotNil(t, retNetworkACL)
		assert.Equal(t, acls.NotPresent, retNetworkACL[acls.AclID(node2.ID)][acls.AclID(node1.ID)])
	})
	deleteAllNodes()
}

func deleteAllNodes() {
	database.DeleteAllRecords(database.NODES_TABLE_NAME)
}

func createTestNode() *models.Node {
	createnode := models.Node{PublicKey: "DM5qhLAE20PG9BbfBCger+Ac9D2NDOwCtY1rbYDLf34=", Name: "testnode", Endpoint: "10.0.0.1", MacAddress: "01:02:03:04:05:06", Password: "password", Network: "skynet", OS: "linux"}
	logic.CreateNode(&createnode)
	return &createnode
}
