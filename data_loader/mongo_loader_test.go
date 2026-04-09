package data_loader

import (
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2"
)

// testCase represents a structure for test cases
type testCase struct {
	name        string
	errorToPass error
	expectError bool
}

// TestHandleEmptyProfilesError tests the handleEmptyProfilesError function using a matrix of test cases.
func TestHandleEmptyProfilesError(t *testing.T) {
	testCases := []testCase{
		{
			name:        "Nil Error",
			errorToPass: nil,
			expectError: false,
		},
		{
			name:        "mongo: no documents in result",
			errorToPass: mongo.ErrNoDocuments,
			expectError: false,
		},
		{
			name:        "not found",
			errorToPass: mgo.ErrNotFound,
			expectError: false,
		},
		{
			name:        "Other Error",
			errorToPass: errors.New("some other error"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := handleEmptyProfilesError(tc.errorToPass)

			isErr := err != nil
			if isErr != tc.expectError {
				t.Errorf("Test '%s' failed: Expected error: %v, Got error: %v", tc.name, tc.expectError, err)
			}
		})
	}
}
