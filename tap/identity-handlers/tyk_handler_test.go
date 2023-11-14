package identityHandlers

import (
	"github.com/markbates/goth"
	"testing"
	"time"
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

func TestStringer(t *testing.T) {
	d := []string{"CN=tyk_admin,OU=Security Groups,OU=NEOLEAP,DC=NEOLEAP,DC=COM,DC=SA",
		"CN=openshift-uat-users,OU=Security Groups,OU=NEOLEAP,DC=NEOLEAP,DC=COM,DC=SA",
		"CN=Neoleap Contract Employees,OU=Email_Group,OU=NEOLEAP,DC=NEOLEAP,DC=COM,DC=SA",
		"CN=VPN-Group-Outsourced,OU=Security Groups,OU=NEOLEAP,DC=NEOLEAP,DC=COM,DC=SA",
		"CN=Normal Group,OU=Security Groups,OU=NEOLEAP,DC=NEOLEAP,DC=COM,DC=SA",
	}
	group := groupsStringer(d, ",")
	t.Log(group)
}

func TestGroupi(t *testing.T) {
	var gUser goth.User

	gUser = goth.User{
		RawData: map[string]interface{}{
			"accountExpires":        "133483572000000000",
			"badPasswordTime":       "133442641065870630",
			"badPwdCount":           "0",
			"cn":                    "Dejan Petrovic",
			"codePage":              "0",
			"company":               "Nortal",
			"countryCode":           "0",
			"dSCorePropagationData": "20231105085000.0Z",
			"department":            "Development & Application Enablement",
			"description":           "Created by IdentityIQ on 10/01/2023 15:11:20",
			"displayName":           "Dejan Petrovic",
			"distinguishedName":     "CN=Dejan Petrovic,OU=Con,OU=IT,OU=NEOLEAP,DC=NEOLEAP,DC=COM,DC=SA",
			"employeeType":          "Vendor",
			"extensionAttribute1":   "dejan.petrovic@nortal.com",
			"givenName":             "Dejan",
			"homeMDB":               "CN=DB03,CN=Databases,CN=Exchange Administrative Group (FYDIBOHF23SPDLT),CN=Administrative Groups,CN=NEOLEAP,CN=Microsoft Exchange,CN=Services,CN=Configuration,DC=NEOLEAP,DC=COM,DC=SA",
			"instanceType":          "4",
			"lastLogoff":            "0",
			"lastLogon":             "133442642622033640",
			"lastLogonTimestamp":    "133441743216979740",
			"legacyExchangeDN":      "/o=NEOLEAP/ou=Exchange Administrative Group (FYDIBOHF23SPDLT)/cn=Recipients/cn=84f26ab45c544c27a968c48edf045189-Dejan Petrovic",
			"lockoutTime":           "0",
			"logonCount":            "1",
			"mDBUseDefaults":        "TRUE",
			"mS-DS-ConsistencyGuid": "\\x05\\x0f\\x84\\x98\\x9eK\\xaeJ\\xa2\\xc5N\\f\\x88ө\\xed",
			"mail":                  "dpetrovic.c@neoleap.com.sa",
			"mailNickname":          "dpetrovic.c",
			"manager":               "CN=Abdullah Alfaleh,OU=Emp,OU=IT,OU=NEOLEAP,DC=NEOLEAP,DC=COM,DC=SA",
			"memberOf": []string{
				"CN=tyk_admin,OU=Security Groups,OU=NEOLEAP,DC=NEOLEAP,DC=COM,DC=SA",
				"CN=openshift-uat-users,OU=Security Groups,OU=NEOLEAP,DC=NEOLEAP,DC=COM,DC=SA",
				"CN=Neoleap Contract Employees,OU=Email_Group,OU=NEOLEAP,DC=NEOLEAP,DC=COM,DC=SA",
				"CN=VPN-Group-Outsourced,OU=Security Groups,OU=NEOLEAP,DC=NEOLEAP,DC=COM,DC=SA",
				"CN=Normal Group,OU=Security Groups,OU=NEOLEAP,DC=NEOLEAP,DC=COM,DC=SA",
			},
			"mobile":                     "+381631614015",
			"msExchArchiveQuota":         "104857600",
			"msExchArchiveWarnQuota":     "94371840",
			"msExchCalendarLoggingQuota": "6291456",
			"msExchDumpsterQuota":        "31457280",
			"msExchDumpsterWarningQuota": "20971520",
			"msExchELCMailboxFlags":      "130",
		},
		Provider:          "ADProvider",
		Email:             "dpetrovic.c@neoleap.com.sa",
		Name:              "",
		FirstName:         "Dejan",
		LastName:          "Petrovic",
		NickName:          "",
		Description:       "",
		UserID:            "Dejan Petrovic",
		AvatarURL:         "",
		Location:          "",
		AccessToken:       "",
		AccessTokenSecret: "",
		RefreshToken:      "",
		ExpiresAt:         time.Time{}, // Set the appropriate time if needed
		IDToken:           "",
	}

	mpping := map[string]string{
		"tyk": "tyk-id",
		"CN=tyk_admin,OU=Security Groups,OU=NEOLEAP,DC=NEOLEAP,DC=COM,DC=SA": "tyk-admin-id",
	}
	id := GetGroupId(gUser, "memberOf", "tyk", mpping, "")
	t.Logf("\nGroup: %v\n", id)
}
