package infoblox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ibclient "github.com/infobloxopen/infoblox-go-client/v2"
)

func validateFuncForSetOfRecordPTR(expectedValues map[string]*ibclient.RecordPTR) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for name, value := range expectedValues {
			if err := validateRecordPTR(name, value, s); err != nil {
				return fmt.Errorf("validation failed for the resource '%s': %s", name, err)
			}
		}
		return nil
	}
}

func validateRecordPTR(
	resourceName string,
	expectedValue *ibclient.RecordPTR,
	s *terraform.State) error {
	res, found := s.RootModule().Resources[resourceName]
	if !found {
		return fmt.Errorf("not found: %s", resourceName)
	}

	id := res.Primary.ID
	if id == "" {
		return fmt.Errorf("ID is not set")
	}

	connector := testAccProvider.Meta().(ibclient.IBConnector)
	objMgr := ibclient.NewObjectManager(
		connector,
		"terraform_test",
		"terraform_test_tenant")
	recPtr, err := objMgr.GetPTRRecordByRef(id)
	if err != nil {
		if isNotFoundError(err) {
			if expectedValue == nil {
				return nil
			}
			return fmt.Errorf("object with ID '%s' not found, but expected to exist", id)
		}
	}

	if recPtr.PtrdName == nil {
		return fmt.Errorf("'ptrdname' is expected to be defined but it is not")
	}
	expPtrdName := expectedValue.PtrdName
	if *recPtr.PtrdName != *expPtrdName {
		return fmt.Errorf(
			"the value of 'ptrdname' field is '%s', but expected '%s'",
			*recPtr.PtrdName, *expPtrdName)
	}

	if recPtr.Comment != nil {
		if expectedValue.Comment == nil {
			return fmt.Errorf("'comment' is expected to be undefined but it is not")
		}
		if *recPtr.Comment != *expectedValue.Comment {
			return fmt.Errorf(
				"'comment' does not match: got '%s', expected '%s'",
				*recPtr.Comment, *expectedValue.Comment)
		}
	} else if expectedValue.Comment != nil {
		return fmt.Errorf("'comment' is expected to be defined but it is not")
	}

	if recPtr.Name == nil {
		return fmt.Errorf("'name' is expected to be defined but it is not")
	}
	expName := expectedValue.Name
	if *recPtr.Name != *expName {
		return fmt.Errorf(
			"the value of 'name' field is '%s', but expected '%s'",
			*recPtr.Name, *expName)
	}

	if recPtr.UseTtl != nil {
		if expectedValue.UseTtl == nil {
			return fmt.Errorf("'use_ttl' is expected to be undefined but it is not")
		}
		if *recPtr.UseTtl != *expectedValue.UseTtl {
			return fmt.Errorf(
				"'use_ttl' does not match: got '%t', expected '%t'",
				*recPtr.UseTtl, *expectedValue.UseTtl)
		}
		if *recPtr.UseTtl {
			if *recPtr.Ttl != *expectedValue.Ttl {
				return fmt.Errorf(
					"'TTL' usage does not match: got '%d', expected '%d'",
					recPtr.Ttl, expectedValue.Ttl)
			}
		}
	}

	expView := expectedValue.View
	if recPtr.View != expView {
		return fmt.Errorf(
			"the value of 'view' field is '%s', but expected '%s'",
			recPtr.View, expView)
	}

	expZone := expectedValue.Zone
	if recPtr.Zone != expZone {
		return fmt.Errorf(
			"the value of 'zone' field is '%s', but expected '%s'",
			recPtr.Zone, expZone)
	}

	expIpv4Addr := expectedValue.Ipv4Addr
	if expIpv4Addr != nil {
		if recPtr.Ipv4Addr == nil {
			return fmt.Errorf("'ipv4addr' is expected to be defined but it is not")
		}
		if *recPtr.Ipv4Addr != *expIpv4Addr {
			return fmt.Errorf(
				"the value of 'ipv4addr' field is '%s', but expected '%s'",
				*recPtr.Ipv4Addr, *expIpv4Addr)
		}
	}

	expIpv6Addr := expectedValue.Ipv6Addr
	if expIpv6Addr != nil {
		if recPtr.Ipv6Addr == nil {
			return fmt.Errorf("'ipv6addr' is expected to be defined but it is not")
		}
		if *recPtr.Ipv6Addr != *expIpv6Addr {
			return fmt.Errorf(
				"the value of 'ipv6addr' field is '%s', but expected '%s'",
				*recPtr.Ipv6Addr, *expIpv6Addr)
		}
	}

	// the rest is about extensible attributes
	expectedEAs := expectedValue.Ea
	if expectedEAs == nil && recPtr.Ea != nil {
		return fmt.Errorf(
			"the object with ID '%s' has 'ext_attrs' field, but it is not expected to exist", id)
	}
	if expectedEAs != nil && recPtr.Ea == nil {
		return fmt.Errorf(
			"the object with ID '%s' has no 'ext_attrs' field, but it is expected to exist", id)
	}
	if expectedEAs == nil {
		return nil
	}

	return validateEAs(recPtr.Ea, expectedEAs)
}

func testAccCheckRecordPTRDestroy(s *terraform.State) error {
	connector := testAccProvider.Meta().(ibclient.IBConnector)
	objMgr := ibclient.NewObjectManager(
		connector,
		"terraform_test",
		"terraform_test_tenant")
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "infoblox_ptr_record" {
			continue
		}
		res, err := objMgr.GetPTRRecordByRef(rs.Primary.ID)
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}
		if res != nil {
			return fmt.Errorf("object with ID '%s' remains", rs.Primary.ID)
		}
	}
	return nil
}

func TestAcc_resourceRecordPTR(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRecordPTRDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCasePtrRecordTestData00,
				Check:  validateFuncForSetOfRecordPTR(testCasePtrRecordExpectedData00),
			},
			{
				Config: testCasePtrRecordTestData01,
				Check:  validateFuncForSetOfRecordPTR(testCasePtrRecordExpectedData01),
			},
			{
				Config:      testCasePtrRecordTestErrData01,
				ExpectError: regexp.MustCompile("only one of 'ip_addr', 'cidr' and 'record_name' must be defined"),
			},
			{
				Config:      testCasePtrRecordTestErrData02,
				ExpectError: regexp.MustCompile("only one of 'ip_addr', 'cidr' and 'record_name' must be defined"),
			},
			{
				Config:      testCasePtrRecordTestErrData03,
				ExpectError: regexp.MustCompile("only one of 'ip_addr', 'cidr' and 'record_name' must be defined"),
			},
			{
				Config:      testCasePtrRecordTestErrData04,
				ExpectError: regexp.MustCompile("only one of 'ip_addr', 'cidr' and 'record_name' must be defined"),
			},
			{
				Config: testCasePtrRecordTestErrData05Pre,
			},
			{
				Config:      testCasePtrRecordTestErrData05,
				ExpectError: regexp.MustCompile("only one of 'cidr', 'ip_addr' and 'record_name' is allowed to be non-empty"),
			},
			{
				Config:      testCasePtrRecordTestErrData06,
				ExpectError: regexp.MustCompile("only one of 'cidr', 'ip_addr' and 'record_name' is allowed to be non-empty"),
			},
			{
				Config:      testCasePtrRecordTestErrData07,
				ExpectError: regexp.MustCompile("only one of 'cidr', 'ip_addr' and 'record_name' is allowed to be non-empty"),
			},
			{
				Config:      testCasePtrRecordTestErrData08,
				ExpectError: regexp.MustCompile("only one of 'cidr', 'ip_addr' and 'record_name' is allowed to be non-empty"),
			},
			{
				Config:      testCasePtrRecordTestErrData09,
				ExpectError: regexp.MustCompile("only one of 'cidr', 'ip_addr' and 'record_name' is allowed to be non-empty"),
			},
			{
				Config:      testCasePtrRecordTestErrData10,
				ExpectError: regexp.MustCompile("only one of 'cidr', 'ip_addr' and 'record_name' is allowed to be non-empty"),
			},
		},
	})
}
