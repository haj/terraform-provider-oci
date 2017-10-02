// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package main

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	baremetal "github.com/oracle/bmcs-go-sdk"

	"github.com/stretchr/testify/suite"
)

type DatasourceCoreInstanceTestSuite struct {
	suite.Suite
	Client 			*baremetal.Client
	Config			string
	Provider		terraform.ResourceProvider
	Providers 		map[string]terraform.ResourceProvider
	ResourceName	string
}

func (s *DatasourceCoreInstanceTestSuite) SetupTest() {
	s.Client = testAccClient
	s.Provider = testAccProvider
	s.Providers = testAccProviders
	s.Config = testProviderConfig() + `
		data "oci_identity_availability_domains" "ADs" {
		compartment_id = "${var.compartment_id}"
	}

	resource "oci_core_virtual_network" "vn" {
		compartment_id = "${var.compartment_id}"
		cidr_block = "10.0.0.0/16"
		display_name = "-tf-vcn"
	}

	resource "oci_core_subnet" "sb" {
		compartment_id      = "${var.compartment_id}"
		vcn_id              = "${oci_core_virtual_network.vn.id}"
		availability_domain = "${lookup(data.oci_identity_availability_domains.ADs.availability_domains[0],"name")}"
		route_table_id      = "${oci_core_virtual_network.vn.default_route_table_id}"
		security_list_ids = ["${oci_core_virtual_network.vn.default_security_list_id}"]
		dhcp_options_id     = "${oci_core_virtual_network.vn.default_dhcp_options_id}"
		cidr_block          = "10.0.1.0/24"
		display_name        = "-tf-subnet"
	}

	data "oci_core_images" "img" {
		compartment_id = "${var.compartment_id}"
		operating_system = "Oracle Linux"
		operating_system_version = "7.3"
		limit = 1
	}

	resource "oci_core_instance" "inst_create" {
		availability_domain = "${data.oci_identity_availability_domains.ADs.availability_domains.0.name}"
		compartment_id = "${var.compartment_id}"
		subnet_id = "${oci_core_subnet.sb.id}"
		image = "${data.oci_core_images.img.images.0.id}"
		shape = "VM.Standard1.1"
		metadata {
			ssh_authorized_keys = "${var.ssh_public_key}"
		}
		timeouts {
			create = "15m"
		}
	}

	data "oci_core_instances" "inst_read" {
		compartment_id = "${var.compartment_id}"
		limit = 1
	}	`

	s.ResourceName = "oci_core_instances.inst_read"
}

func (s *DatasourceCoreInstanceTestSuite) TestAccDatasourceCoreInstance_basic() {
	resource.Test(s.T(), resource.TestCase {
		PreventPostDestroyRefresh: 	true,
		Providers:					s.Providers,
		Steps:	[]resource.TestStep{
			{
				ImportState: true,
				ImportStateVerify: true,
				Config: s.Config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(s.ResourceName,"availability_domain"),
					resource.TestCheckResourceAttr(s.ResourceName, "instances.#", "1"),
					resource.TestCheckResourceAttrSet(s.ResourceName,"instances.0.id"),
					resource.TestCheckResourceAttrSet(s.ResourceName, "instances.0.display_name"),
					resource.TestCheckResourceAttrSet(s.ResourceName, "instances.0.region"),
					resource.TestCheckResourceAttrSet(s.ResourceName, "instances.0.state"),
					resource.TestCheckResourceAttrSet(s.ResourceName, "instances.0.shape"),
					resource.TestCheckResourceAttrSet(s.ResourceName, "instances.0.image"),
					resource.TestCheckResourceAttrSet(s.ResourceName, "instances.0.ipxe_script"),
					resource.TestCheckResourceAttrSet(s.ResourceName, "instances.0.metadata"),
				),
			},
		},
	},
	)
}

func TestDatasourceCoreInstanceTestSuite(t *testing.T) {
	suite.Run(t, new(DatasourceCoreInstanceTestSuite))
}