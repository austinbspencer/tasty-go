package tasty

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateSession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/sessions", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, sessionResp)
	})

	resp, err := client.CreateSession(LoginInfo{Login: "default", Password: "Password"})
	require.Nil(t, err)

	require.Equal(t, "default@gmail.com", resp.User.Email)
	require.Equal(t, "default", resp.User.Username)
	require.Equal(t, "U0001563674", resp.User.ExternalID)
	require.NotNil(t, resp.SessionToken)
	require.Equal(t, "gwYY_cUTvG3ESb9Sio6f9cX-SA-9GpIyXqFZ4Wx_pL_lnwD_6yQPty+C", *resp.SessionToken)
	require.Nil(t, resp.RememberToken)
}

func TestValidateSession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/sessions/validate", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, sessionResp)
	})

	resp, err := client.ValidateSession()
	require.Nil(t, err)

	require.Equal(t, "default@gmail.com", resp.User.Email)
	require.Equal(t, "default", resp.User.Username)
	require.Equal(t, "U0001563674", resp.User.ExternalID)
	require.NotNil(t, resp.SessionToken)
	require.Equal(t, "gwYY_cUTvG3ESb9Sio6f9cX-SA-9GpIyXqFZ4Wx_pL_lnwD_6yQPty+C", *resp.SessionToken)
	require.Nil(t, resp.RememberToken)
}

func TestDestroySession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/sessions", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, sessionResp)
	})

	err := client.DestroySession()
	require.Nil(t, err)
}

const sessionResp = `{
  "data": {
    "user": {
      "email": "default@gmail.com",
      "username": "default",
      "external-id": "U0001563674"
    },
    "session-token": "gwYY_cUTvG3ESb9Sio6f9cX-SA-9GpIyXqFZ4Wx_pL_lnwD_6yQPty+C"
  },
  "context": "/sessions"
}`