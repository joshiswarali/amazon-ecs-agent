// +build windows,unit

// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package ecscni

import (
	"encoding/json"
	"net"
	"testing"

	apieni "github.com/aws/amazon-ecs-agent/agent/api/eni"
	"github.com/stretchr/testify/assert"
)

const (
	linkName                = "Ethernet 4"
	validVPCGatewayCIDR     = "10.0.0.1/24"
	validVPCGatewayIPv4Addr = "10.0.0.1"
	invalidVPCGatewayCIDR   = "10.0.0.300/24"
	invalidVPCGatewayAddr   = "10.0.0.300"
	validDNSServer          = "10.0.0.2"
	ipv4                    = "10.0.0.120"
	ipv4CIDR                = "10.0.0.120/24"
	vpcCIDR                 = "10.0.0.0/16"
	mac                     = "02:7b:64:49:b1:40"
	cniMinSupportedVersion  = "1.0.0"
)

func getTaskENI() *apieni.ENI {
	return &apieni.ENI{
		ID:                           "TestENI",
		LinkName:                     linkName,
		MacAddress:                   mac,
		InterfaceAssociationProtocol: apieni.DefaultInterfaceAssociationProtocol,
		SubnetGatewayIPV4Address:     validVPCGatewayCIDR,
		IPV4Addresses: []*apieni.ENIIPV4Address{
			{
				Primary: true,
				Address: ipv4,
			},
		},
	}
}

func getCNIConfig() *Config {
	_, cidr, _ := net.ParseCIDR(vpcCIDR)
	return &Config{
		MinSupportedCNIVersion: cniMinSupportedVersion,
		PrimaryIPv4VPCCIDR:     cidr,
	}
}

// TestNewVPCENIPluginConfigForTaskNSSetup tests the generated configuration when all parameters are valid.
func TestNewVPCENIPluginConfigForTaskNSSetup(t *testing.T) {
	taskENI := getTaskENI()
	cniConfig := getCNIConfig()
	config, err := NewVPCENIPluginConfigForTaskNSSetup(taskENI, cniConfig)

	netConfig := &VPCENIPluginConfig{}
	json.Unmarshal(config.Bytes, netConfig)

	assert.NoError(t, err)
	assert.EqualValues(t, ECSVPCENIPluginExecutable, config.Network.Type)
	assert.EqualValues(t, cniMinSupportedVersion, config.Network.CNIVersion)
	assert.EqualValues(t, []string{validDNSServer}, netConfig.DNS.Nameservers)
	assert.EqualValues(t, ipv4CIDR, netConfig.ENIIPAddress)
	assert.EqualValues(t, mac, netConfig.ENIMACAddress)
	assert.EqualValues(t, validVPCGatewayIPv4Addr, netConfig.GatewayIPAddress)
	assert.EqualValues(t, linkName, netConfig.ENIName)
	assert.False(t, netConfig.UseExistingNetwork)
	assert.False(t, netConfig.NoInfraContainer)
}

// TestConstructDNSFromVPCCIDRSuccess tests if the dns is constructed properly from the given vpc's primary cidr
func TestConstructDNSFromVPCCIDRSuccess(t *testing.T) {
	_, cidr, _ := net.ParseCIDR(vpcCIDR)
	result, err := constructDNSFromVPCCIDR(cidr)

	assert.NoError(t, err)
	assert.EqualValues(t, []string{validDNSServer}, result)
}

// TestConstructDNSFromVPCCIDRError tests if an error is thrown if the vpc cidr is invalid
func TestConstructDNSFromVPCCIDRError(t *testing.T) {
	result, err := constructDNSFromVPCCIDR(nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}