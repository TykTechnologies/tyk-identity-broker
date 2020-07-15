package identityHandlers

import (
	"github.com/markbates/goth"
	"testing"
)

const TestEmail = "test@tyk.io"

func TestGetEmail(t *testing.T) {
	cases := []struct {
		TestName         string
		CustomEmailField string
		user             goth.User
		ExpectedEmail    string
	}{
		{
			TestName: "Custom email field empty & goth.User email not empty",
			CustomEmailField: "",
			user: goth.User{
				Email:TestEmail,
			},
			ExpectedEmail:TestEmail,
		},
		{
			TestName: "Custom email empty & goth.User email empty",
			CustomEmailField:"",
			user: goth.User{
				Email:"",
			},
			ExpectedEmail: DefaultSSOEmail,
		},
		{
			TestName: "Custom email not empty but field doesn't exist",
			CustomEmailField: "myEmailField",
			user:goth.User{},
			ExpectedEmail: DefaultSSOEmail,
		},
		{
			TestName: "Custom email not empty and is a valid field",
			CustomEmailField: "myEmailField",
			user:goth.User{
				RawData: map[string]interface{}{
					"myEmailField":TestEmail,
				},
			},
			ExpectedEmail:TestEmail,
		},
	}

	for _, tc := range cases {
		t.Run(tc.TestName, func(t *testing.T) {
			email := GetEmail(tc.user,tc.CustomEmailField)
			if email != tc.ExpectedEmail {
				t.Errorf("Email for SSO incorrect. Expected:%v got:%v",tc.ExpectedEmail,email)
			}
		})
	}
}
