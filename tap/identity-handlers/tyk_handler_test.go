package identityHandlers

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TykTechnologies/tyk-identity-broker/tap"
	tyk "github.com/TykTechnologies/tyk-identity-broker/tyk-api"
	"github.com/markbates/goth"
)

const (
	TestEmail      = "test@tyk.io"
	TestId         = "user-id"
	DefaultGroupId = "default-group-id"
	ProfileOrgId   = "profile-org-id"
	DefaultOrgId   = "default-org-id"
)

var UserGroupMapping = map[string]string{
	"devs":   "devs-group",
	"admins": "admins-group",
	"CN=tyk_admin,OU=Security Groups,OU=GenericOrg,DC=GenericOrg,DC=COM,DC=GEN": "tyk-admin",
}

var OrgMapping = map[string]string{
	"acme":              "acme-org-id",
	"Globex Industries": "globex-org-id",
	"CN=tyk_tenant,OU=Security Groups,OU=GenericOrg,DC=GenericOrg,DC=COM,DC=GEN": "tyk-tenant-org-id",
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
		ExpectedGroupsIDs  []string
		DefaultGroupID     string
		UserGroupMapping   map[string]string
		UserGroupSeparator string
	}{
		{
			TestName:           "Custom group id field empty",
			CustomGroupIDField: "",
			user:               goth.User{},
			ExpectedGroupsIDs:  []string{},
			DefaultGroupID:     "",
			UserGroupMapping:   UserGroupMapping,
		},
		{
			TestName:           "Custom group id field empty & default group set",
			CustomGroupIDField: "",
			user:               goth.User{},
			ExpectedGroupsIDs:  []string{DefaultGroupId},
			DefaultGroupID:     DefaultGroupId,
			UserGroupMapping:   UserGroupMapping,
		},
		{
			TestName:           "Custom group id field not empty but invalid & default group set",
			CustomGroupIDField: "my-custom-group-id-field",
			user:               goth.User{},
			DefaultGroupID:     DefaultGroupId,
			ExpectedGroupsIDs:  []string{DefaultGroupId},
			UserGroupMapping:   UserGroupMapping,
		},
		{
			TestName:           "Custom group id field not empty but invalid & default group not set",
			CustomGroupIDField: "my-custom-group-id-field",
			user:               goth.User{},
			ExpectedGroupsIDs:  []string{},
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
			ExpectedGroupsIDs: []string{"admins-group"},
			DefaultGroupID:    "",
			UserGroupMapping:  UserGroupMapping,
		},
		{
			TestName:           "Receive many groups from idp with blank space separated",
			CustomGroupIDField: "my-custom-group-id-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-group-id-field": "devs admins",
				},
			},
			ExpectedGroupsIDs: []string{"devs-group", "admins-group"},
			DefaultGroupID:    "",
			UserGroupMapping:  UserGroupMapping,
		},
		{
			TestName:           "Receive many groups from idp with comma separated",
			CustomGroupIDField: "my-custom-group-id-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-group-id-field": "devs,admins",
				},
			},
			ExpectedGroupsIDs:  []string{"devs-group", "admins-group"},
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
			ExpectedGroupsIDs: []string{"admins-group"},
			DefaultGroupID:    "devs",
			UserGroupMapping:  UserGroupMapping,
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
			ExpectedGroupsIDs: []string{"tyk-admin"},
			DefaultGroupID:    "devs",
			UserGroupMapping:  UserGroupMapping,
		},
	}

	for _, tc := range cases {
		t.Run(tc.TestName, func(t *testing.T) {
			ids := GetGroupId(tc.user, tc.CustomGroupIDField, tc.DefaultGroupID, tc.UserGroupMapping, tc.UserGroupSeparator)
			assert.Equal(t, tc.ExpectedGroupsIDs, ids)
		})
	}
}

func Test_defaultOrEmptyGroupIDs(t *testing.T) {
	tests := []struct {
		name             string
		defaultUserGroup string
		expectedGroupIDs []string
	}{
		{
			name:             "Empty default user group",
			defaultUserGroup: "",
			expectedGroupIDs: []string{},
		},
		{
			name:             "Non-empty default user group",
			defaultUserGroup: "defaultGroup",
			expectedGroupIDs: []string{"defaultGroup"},
		},
		{
			name:             "Default user group with spaces",
			defaultUserGroup: "default group",
			expectedGroupIDs: []string{"default group"},
		},
		{
			name:             "Default user group with special characters",
			defaultUserGroup: "group@123",
			expectedGroupIDs: []string{"group@123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultOrEmptyGroupIDs(tt.defaultUserGroup)
			assert.Equal(t, tt.expectedGroupIDs, result, "The group IDs should match")
		})
	}
}

func TestGetOrgId(t *testing.T) {
	cases := []struct {
		TestName           string
		CustomUserOrgField string
		user               goth.User
		DefaultOrgID       string
		OrgMapping         map[string]string
		ProfileOrgID       string
		ExpectedOrgID      string
	}{
		{
			TestName:           "Custom org field empty, profile org is used",
			CustomUserOrgField: "",
			user:               goth.User{},
			DefaultOrgID:       "",
			OrgMapping:         OrgMapping,
			ProfileOrgID:       ProfileOrgId,
			ExpectedOrgID:      ProfileOrgId,
		},
		{
			TestName:           "Custom org field empty, profile org wins over the default org",
			CustomUserOrgField: "",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-org-field": "acme",
				},
			},
			DefaultOrgID:  DefaultOrgId,
			OrgMapping:    OrgMapping,
			ProfileOrgID:  ProfileOrgId,
			ExpectedOrgID: ProfileOrgId,
		},
		{
			TestName:           "Custom org field set & claim mapped. With default org not set",
			CustomUserOrgField: "my-custom-org-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-org-field": "acme",
				},
			},
			DefaultOrgID:  "",
			OrgMapping:    OrgMapping,
			ProfileOrgID:  ProfileOrgId,
			ExpectedOrgID: "acme-org-id",
		},
		{
			TestName:           "Custom org field set & claim mapped. With default org set",
			CustomUserOrgField: "my-custom-org-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-org-field": "acme",
				},
			},
			DefaultOrgID:  DefaultOrgId,
			OrgMapping:    OrgMapping,
			ProfileOrgID:  ProfileOrgId,
			ExpectedOrgID: "acme-org-id",
		},
		{
			TestName:           "Claim value containing blank spaces is not split",
			CustomUserOrgField: "my-custom-org-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-org-field": "Globex Industries",
				},
			},
			DefaultOrgID:  DefaultOrgId,
			OrgMapping:    OrgMapping,
			ProfileOrgID:  ProfileOrgId,
			ExpectedOrgID: "globex-org-id",
		},
		{
			TestName:           "Claim present but unmapped & default org set",
			CustomUserOrgField: "my-custom-org-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-org-field": "unknown-tenant",
				},
			},
			DefaultOrgID:  DefaultOrgId,
			OrgMapping:    OrgMapping,
			ProfileOrgID:  ProfileOrgId,
			ExpectedOrgID: DefaultOrgId,
		},
		{
			TestName:           "Claim present but unmapped & default org not set",
			CustomUserOrgField: "my-custom-org-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-org-field": "unknown-tenant",
				},
			},
			DefaultOrgID:  "",
			OrgMapping:    OrgMapping,
			ProfileOrgID:  ProfileOrgId,
			ExpectedOrgID: ProfileOrgId,
		},
		{
			TestName:           "Claim absent & default org set",
			CustomUserOrgField: "my-custom-org-field",
			user:               goth.User{},
			DefaultOrgID:       DefaultOrgId,
			OrgMapping:         OrgMapping,
			ProfileOrgID:       ProfileOrgId,
			ExpectedOrgID:      DefaultOrgId,
		},
		{
			TestName:           "Claim absent & default org not set",
			CustomUserOrgField: "my-custom-org-field",
			user:               goth.User{},
			DefaultOrgID:       "",
			OrgMapping:         OrgMapping,
			ProfileOrgID:       ProfileOrgId,
			ExpectedOrgID:      ProfileOrgId,
		},
		{
			TestName:           "Claim present but nil",
			CustomUserOrgField: "my-custom-org-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-org-field": nil,
				},
			},
			DefaultOrgID:  DefaultOrgId,
			OrgMapping:    OrgMapping,
			ProfileOrgID:  ProfileOrgId,
			ExpectedOrgID: DefaultOrgId,
		},
		{
			TestName:           "Claim is an empty string",
			CustomUserOrgField: "my-custom-org-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-org-field": "",
				},
			},
			DefaultOrgID:  DefaultOrgId,
			OrgMapping:    OrgMapping,
			ProfileOrgID:  ProfileOrgId,
			ExpectedOrgID: DefaultOrgId,
		},
		{
			TestName:           "Claim is an array, the first value that maps wins",
			CustomUserOrgField: "orgs",
			user: goth.User{
				RawData: map[string]interface{}{
					"orgs": []interface{}{
						"unknown-tenant",
						"Globex Industries",
						"acme",
					},
				},
			},
			DefaultOrgID:  DefaultOrgId,
			OrgMapping:    OrgMapping,
			ProfileOrgID:  ProfileOrgId,
			ExpectedOrgID: "globex-org-id",
		},
		{
			TestName:           "Claim is an array of strings",
			CustomUserOrgField: "memberOf",
			user: goth.User{
				RawData: map[string]interface{}{
					"memberOf": []string{
						"CN=Generic Contract Employees,OU=Email_Group,OU=GenericOrg,DC=GenericOrg,DC=COM,DC=GEN",
						"CN=tyk_tenant,OU=Security Groups,OU=GenericOrg,DC=GenericOrg,DC=COM,DC=GEN",
					},
				},
			},
			DefaultOrgID:  DefaultOrgId,
			OrgMapping:    OrgMapping,
			ProfileOrgID:  ProfileOrgId,
			ExpectedOrgID: "tyk-tenant-org-id",
		},
		{
			TestName:           "Claim is an array with no value mapped",
			CustomUserOrgField: "orgs",
			user: goth.User{
				RawData: map[string]interface{}{
					"orgs": []interface{}{"unknown-tenant", "another-unknown-tenant"},
				},
			},
			DefaultOrgID:  "",
			OrgMapping:    OrgMapping,
			ProfileOrgID:  ProfileOrgId,
			ExpectedOrgID: ProfileOrgId,
		},
		{
			TestName:           "Claim is of a type we cannot read",
			CustomUserOrgField: "orgs",
			user: goth.User{
				RawData: map[string]interface{}{
					"orgs": []interface{}{float64(42)},
				},
			},
			DefaultOrgID:  DefaultOrgId,
			OrgMapping:    OrgMapping,
			ProfileOrgID:  ProfileOrgId,
			ExpectedOrgID: DefaultOrgId,
		},
		{
			TestName:           "Claim maps to an empty org id, it is never returned",
			CustomUserOrgField: "my-custom-org-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-org-field": "acme",
				},
			},
			DefaultOrgID:  "",
			OrgMapping:    map[string]string{"acme": ""},
			ProfileOrgID:  ProfileOrgId,
			ExpectedOrgID: ProfileOrgId,
		},
		{
			TestName:           "Mapping not set at all",
			CustomUserOrgField: "my-custom-org-field",
			user: goth.User{
				RawData: map[string]interface{}{
					"my-custom-org-field": "acme",
				},
			},
			DefaultOrgID:  "",
			OrgMapping:    nil,
			ProfileOrgID:  ProfileOrgId,
			ExpectedOrgID: ProfileOrgId,
		},
	}

	for _, tc := range cases {
		t.Run(tc.TestName, func(t *testing.T) {
			orgID := GetOrgId(tc.user, tc.CustomUserOrgField, tc.DefaultOrgID, tc.OrgMapping, tc.ProfileOrgID)
			assert.Equal(t, tc.ExpectedOrgID, orgID)
			assert.NotEmpty(t, orgID, "The org id handed to the dashboard must never be empty")
		})
	}
}

func Test_defaultOrProfileOrgID(t *testing.T) {
	tests := []struct {
		name          string
		defaultOrgID  string
		profileOrgID  string
		expectedOrgID string
	}{
		{
			name:          "Empty default org",
			defaultOrgID:  "",
			profileOrgID:  ProfileOrgId,
			expectedOrgID: ProfileOrgId,
		},
		{
			name:          "Non-empty default org",
			defaultOrgID:  DefaultOrgId,
			profileOrgID:  ProfileOrgId,
			expectedOrgID: DefaultOrgId,
		},
		{
			name:          "Both empty",
			defaultOrgID:  "",
			profileOrgID:  "",
			expectedOrgID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultOrProfileOrgID(tt.defaultOrgID, tt.profileOrgID)
			assert.Equal(t, tt.expectedOrgID, result, "The org id should match")
		})
	}
}

// TestCreateIdentityOrgID checks the org resolution reaches the SSOAccessData we post to the dashboard
func TestCreateIdentityOrgID(t *testing.T) {
	cases := []struct {
		TestName      string
		profile       tap.Profile
		user          goth.User
		ExpectedOrgID string
	}{
		{
			TestName: "Profile without a custom org field keeps posting its own org",
			profile: tap.Profile{
				ActionType: tap.GenerateOrLoginUserProfile,
				OrgID:      ProfileOrgId,
			},
			user: goth.User{
				Email: TestEmail,
				RawData: map[string]interface{}{
					"my-custom-org-field": "acme",
				},
			},
			ExpectedOrgID: ProfileOrgId,
		},
		{
			TestName: "Profile with a custom org field posts the mapped org",
			profile: tap.Profile{
				ActionType:         tap.GenerateOrLoginUserProfile,
				OrgID:              ProfileOrgId,
				CustomUserOrgField: "my-custom-org-field",
				OrgMapping:         OrgMapping,
				DefaultOrgID:       DefaultOrgId,
			},
			user: goth.User{
				Email: TestEmail,
				RawData: map[string]interface{}{
					"my-custom-org-field": "acme",
				},
			},
			ExpectedOrgID: "acme-org-id",
		},
		{
			TestName: "Profile with a custom org field but no claim posts the default org",
			profile: tap.Profile{
				ActionType:         tap.GenerateOrLoginUserProfile,
				OrgID:              ProfileOrgId,
				CustomUserOrgField: "my-custom-org-field",
				OrgMapping:         OrgMapping,
				DefaultOrgID:       DefaultOrgId,
			},
			user: goth.User{
				Email: TestEmail,
			},
			ExpectedOrgID: DefaultOrgId,
		},
	}

	for _, tc := range cases {
		t.Run(tc.TestName, func(t *testing.T) {
			var accessRequest SSOAccessData
			api := &tyk.TykAPI{
				CustomDispatcher: func(_ tyk.Endpoint, _ string, _ string, body io.Reader) ([]byte, int, error) {
					sentData, err := io.ReadAll(body)
					if err != nil {
						return []byte{}, http.StatusInternalServerError, err
					}

					if err := json.Unmarshal(sentData, &accessRequest); err != nil {
						return []byte{}, http.StatusInternalServerError, err
					}

					return []byte(`{"Meta":"the-nonce"}`), http.StatusOK, nil
				},
			}

			handler := TykIdentityHandler{API: api}
			assert.NoError(t, handler.Init(tc.profile))

			nonce, err := handler.CreateIdentity(tc.user)
			assert.NoError(t, err)
			assert.Equal(t, "the-nonce", nonce)
			assert.Equal(t, tc.ExpectedOrgID, accessRequest.OrgID)
		})
	}
}
