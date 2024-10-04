package handlers_test

import (
	"errors"
	"net/http/httptest"
)

//var ctrl *gomock.Controller
//var userHandler *handlers.UserHandler
//var mockUserService service_interfaces.UserService
//
//func setup(t *testing.T) func() {
//	ctrl = gomock.NewController(t)
//
//	// Create mock userService
//	mockUserService = mocks.NewMockUserService(ctrl)
//
//	// Create an instance of UserHandler with the mocked service
//	userHandler = handlers.NewUserHandler(mockUserService)
//
//	return func() {
//		defer ctrl.Finish()
//	}
//}

// Custom ResponseWriter that simulates a write failure
type failingResponseWriter struct {
	*httptest.ResponseRecorder
}

func (f *failingResponseWriter) Write(b []byte) (int, error) {
	return 0, errors.New("failed to write")
}
