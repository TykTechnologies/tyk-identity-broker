package identityHandlers

import (
	"testing"

	"github.com/markbates/goth"
)

const (
	TestEmail      = "test@tyk.io"
	TestId         = "user-id"
	DefaultGroupId = "default-group-id"
)

var UserGroupMapping = map[string]string{
	"devs":   "devs-group",
	"admins": "admins-group",
	"CN=tyk_admin,OU=Security Groups,OU=GenericOrg,DC=GenericOrg,DC=COM,DC=GEN": "tyk-admin",
}

func TestGetEmail(t *testing.T) {
	cases := []struct {
		TestName         string
		CustomEmailField string
		user             goth.User
		ExpectedEmail    string
	}{
		{
			TestName:         "Custom email field empty & goth.User email not empty",
			CustomEmailField: "",
			user: goth.User{
				Email: TestEmail,
			},
			ExpectedEmail: TestEmail,
		},
		{
			TestName:         "Custom email empty & goth.User email empty",
			CustomEmailField: "",
			user: goth.User{
				Email: "",
			},
			ExpectedEmail: DefaultSSOEmail,
		},
		{
			TestName:         "Custom email not empty but field doesn't exist",
			CustomEmailField: "myEmailField",
			user:             goth.User{},
			ExpectedEmail:    DefaultSSOEmail,
		},
		{
			TestName:         "Custom email not empty and is a valid field",
			CustomEmailField: "myEmailField",
			user: goth.User{
				RawData: map[string]interface{}{
					"myEmailField": TestEmail,
				},
			},
			ExpectedEmail: TestEmail,
		},
	}

	for _, tc := range cases {
		t.Run(tc.TestName, func(t *testing.T) {
			email := GetEmail(tc.user, tc.CustomEmailField)
			if email != tc.ExpectedEmail {
				t.Errorf("Email for SSO incorrect. Expected:%v got:%v", tc.ExpectedEmail, email)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	cases := []struct {
		TestName      string
		CustomIDField string
		user          goth.User
		ExpectedID    string
	}{
		{
			TestName:      "Custom id field empty",
			CustomIDField: "",
			user: goth.User{
				UserID: TestId,
			},
			ExpectedID: TestId,
		},
		{
			TestName:      "Custom id not empty but field doesn't exist",
			CustomIDField: "myIdField",
			user: goth.User{
				UserID: TestId,
			},
			ExpectedID: TestId,
		},
		{
			TestName:      "Custom id not empty and is a valid field",
			CustomIDField: "myIdField",
			user: goth.User{
				UserID: TestId,
				RawData: map[string]interface{}{
					"myIdField": "customId",
				},
			},
			ExpectedID: "customId",
		},
	}

	for _, tc := range cases {
		t.Run(tc.TestName, func(t *testing.T) {
			id := GetUserID(tc.user, tc.CustomIDField)
			if id != tc.ExpectedID {
				t.Errorf("User id incorrect. Expected:%v got:%v", tc.ExpectedID, id)
			}
		})
	}
}

func TestGetGroupId(t *testing.T) {
	cases := []struct {
		TestName           string
		CustomGroupIDField string
		user               goth.User
		ExpectedGroupID    string
		DefaultGroupID     string
		UserGroupMapping   map[string]string
		UserGroupSeparator string
	}{
		{
			TestName:           "Custom group id field empty",
			CustomGroupIDField: "",
			user:               goth.User{},
			ExpectedGroupID:    "",
			DefaultGroupID:     "",
			UserGroupMapping:   UserGroupMapping,
		},
		{
			TestName:           "Custom group id field empty & default group set",
			CustomGroupIDField: "",
			user:               goth.User{},
			ExpectedGroupID:    DefaultGroupId,
			DefaultGroupID:     DefaultGroupId,
			UserGroupMapping:   UserGroupMapping,
		},
		{
			TestName:           "Custom group id field not empty but invalid & default group set",
			CustomGroupIDField: "my-custom-group-id-field",
			user:               goth.User{},
			ExpectedGroupID:    DefaultGroupId,
			DefaultGroupID:     DefaultGroupId,
			UserGroupMapping:   UserGroupMapping,
		},
		{
			TestName:           "Custom group id field not empty but invalid & default group not set",
			CustomGroupIDField: "my-custom-group-id-field",
			user:               goth.User{},
			ExpectedGroupID:    "",
			DefaultGroupID:     "",
			UserGroupMapping:   UserGroupMapping,
		},
		{
			TestName:           "Custom group id field not empty & valid. With default group not set",
			CustomGroupIDField: "my-custom-group-id-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-group-id-field": "admins",
				},
			},
			ExpectedGroupID:  "admins-group",
			DefaultGroupID:   "",
			UserGroupMapping: UserGroupMapping,
		},
		{
			TestName:           "Receive many groups from idp with blank space separated",
			CustomGroupIDField: "my-custom-group-id-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-group-id-field": "devs admins",
				},
			},
			ExpectedGroupID:  "admins-group",
			DefaultGroupID:   "",
			UserGroupMapping: UserGroupMapping,
		},
		{
			TestName:           "Receive many groups from idp with comma separated",
			CustomGroupIDField: "my-custom-group-id-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-group-id-field": "devs,admins",
				},
			},
			ExpectedGroupID:    "admins-group",
			DefaultGroupID:     "",
			UserGroupMapping:   UserGroupMapping,
			UserGroupSeparator: ",",
		},
		{
			TestName:           "Custom group id field not empty & valid. With default group set",
			CustomGroupIDField: "my-custom-group-id-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-group-id-field": "admins",
				},
			},
			ExpectedGroupID:  "admins-group",
			DefaultGroupID:   "devs",
			UserGroupMapping: UserGroupMapping,
		},
		{
			TestName:           "Custom group id field not empty, and the claim being an array",
			CustomGroupIDField: "memberOf",
			user: goth.User{RawData: map[string]interface{}{
				"memberOf": []string{
					"CN=tyk_admin,OU=Security Groups,OU=GenericOrg,DC=GenericOrg,DC=COM,DC=GEN",
					"CN=openshift-uat-users,OU=Security Groups,OU=GenericOrg,DC=GenericOrg,DC=COM,DC=GEN",
					"CN=Generic Contract Employees,OU=Email_Group,OU=GenericOrg,DC=GenericOrg,DC=COM,DC=GEN",
					"CN=VPN-Group-Outsourced,OU=Security Groups,OU=GenericOrg,DC=GenericOrg,DC=COM,DC=GEN",
					"CN=Normal Group,OU=Security Groups,OU=GenericOrg,DC=GenericOrg,DC=COM,DC=GEN",
				},
			}},
			ExpectedGroupID:  "tyk-admin",
			DefaultGroupID:   "devs",
			UserGroupMapping: UserGroupMapping,
		},
	}

	for _, tc := range cases {
		t.Run(tc.TestName, func(t *testing.T) {
			id := GetGroupId(tc.user, tc.CustomGroupIDField, tc.DefaultGroupID, tc.UserGroupMapping, tc.UserGroupSeparator)
			if id != tc.ExpectedGroupID {
				t.Errorf("group id incorrect. Expected:%v got:%v", tc.ExpectedGroupID, id)
			}
		})
	}
}
