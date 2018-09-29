package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/cmd/server/context"
	"github.com/peerpx/peerpx/entities/user"
	"github.com/peerpx/peerpx/services/config"
)

// UserProfile dislay user profile
// Return
// - defaut -> html
// - if header Accept: application/activitypub
// or user.activitypub -> activitypub
func UserProfile(ac echo.Context) error {
	c := ac.(*context.AppContext)

	switch c.GetWantedContentType() {
	case "json":
		// Get user
		// remove .json
		userName := strings.ToLower(c.Param("username"))
		if strings.HasSuffix(userName, ".json") {
			userName = userName[:len(userName)-5]
		}
		u, err := user.GetByUsername(userName)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.NoContent(http.StatusNotFound)
			}
			c.LogErrorf("handlers.UserProfile - user.GetByUsername(%s) failed: %v", userName, err)
			return c.String(http.StatusInternalServerError, "internal server error")
		}

		// Load template
		t, err := tplBox.MustString("activitypub/user_profile.tpl")
		if err != nil {
			c.LogErrorf("handlers.UserProfile -tplBox.MustString(activitypub/user_profile.tpl) failed: %v", err)
			return c.String(http.StatusInternalServerError, "internal server error")
		}
		tpl, err := template.New("up").Parse(t)
		if err != nil {
			c.LogErrorf("handlers.UserProfile - template new failed: %v", err)
			return c.String(http.StatusInternalServerError, "internal server error")
		}

		// /\n -> \n
		PubKey := bytes.Replace([]byte(u.PublicKey.String), []byte{10}, []byte{92, 110}, -1)

		tplData := struct {
			BaseURL  string
			UserName string
			Summary  string
			PubKey   string
		}{
			BaseURL:  config.GetString("hostname"),
			UserName: u.Username,
			Summary:  "",
			PubKey:   string(PubKey),
		}
		out := bytes.NewBuffer(nil)
		if err = tpl.Execute(out, tplData); err != nil {
			c.LogErrorf("handlers.UserProfile - template execute failed: %v", err)
			return c.String(http.StatusInternalServerError, "internal server error")
		}
		return c.Blob(200, "application/activity+json; charset=utf-8", out.Bytes())

	case "atom":
		return c.String(http.StatusNotFound, "atom is not implemented yet")
	default:
		return c.String(http.StatusNotFound, "html is not implemented yet")
	}
}

// Federation KissFed

// UserGetPublicKey return user public key
func UserGetPublicKey(ac echo.Context) error {
	c := ac.(*context.AppContext)
	user, err := user.GetByUsername(c.Param("username"))
	if err != nil {
		if err == sql.ErrNoRows {
			c.LogInfof("handlers.UserGetPublicKey - user.GetBuUserName(%s): no such user", c.Param("username"))
			return c.String(http.StatusNotFound, "no such user")
		}
		c.LogErrorf("handlers.UserGetPublicKey - user.GetBuUserName(%s) failed: %v", c.Param("username"), err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.String(http.StatusOK, user.PublicKey.String)
}

// UserNewFollower follow request
func UserNewFollower(ac echo.Context) error {
	return nil
}

/////////////////////////////////////////////////////
// API

// UserUsernameIsAvailable checek if username if available
func UserUsernameIsAvailable(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewApiResponse(c.UUID)

	username := c.Param("username")

	// should not happen
	if username == "" {
		return response.Error(c, http.StatusBadRequest, "usernameIsEmpty", "")
	}
	if _, err := user.GetByUsername(username); err != nil {
		if err == sql.ErrNoRows {
			return response.OK(c, http.StatusOK)
		}
		msg := fmt.Sprintf("%s - %s - handlers.UserUsernameIsAvailable - user.GetByUsername(%s) failed: %v", c.RealIP(), response.UUID, username, err)
		return response.Error(c, http.StatusInternalServerError, "userGetByUsernameFail", msg)
	}
	// username exist
	response.Code = "usernameNotAvailable"
	return response.KO(c, http.StatusOK)
}

type userCreateRequest struct {
	Email    string
	Password string
	Username string
}

// UserCreate create a new user
func UserCreate(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewApiResponse(c.UUID)

	// get body
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.UserCreate - failed to read request body: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusBadRequest, "requestBodyNotReadable", msg)

	}
	// unmarshal
	requestData := new(userCreateRequest)
	if err = json.Unmarshal(body, requestData); err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.UserAdd - unmarshall request body failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusBadRequest, "requestBodyNotValidJson", msg)
	}

	// check entries
	// todo remove space from password &&
	// todo username must be alnum

	user, err := user.Create(requestData.Email, requestData.Username, requestData.Password)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.UserAdd - user.Create() failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "userCreateFailed", msg)
	}

	response.Data, err = json.Marshal(user)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.UserAdd - activitypub.Marshal(user) failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "userMarshalFailed", msg)
	}

	c.LogInfof("%s - %s - handlers.UserAdd - new user created: %s %s", c.RealIP(), response.UUID, user.Username, user.Email)
	return response.OK(c, http.StatusCreated)
}

type userLoginRequest struct {
	Login    string `activitypub:"login"`
	Password string `activitypub:"password"`
}

// UserLogin used to login
func UserLogin(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewApiResponse(c.UUID)

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.UserLogin - unable to read request body: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusBadRequest, "requestBodyNotReadable", msg)
	}

	// unmarshall
	requestData := new(userLoginRequest)
	if err = json.Unmarshal(body, requestData); err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.UserLogin - unable to unmarshall request body: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusBadRequest, "requestBodyNotValidJson", msg)
	}

	u, err := user.Login(requestData.Login, requestData.Password)
	if err != nil {
		if err == user.ErrNoSuchUser {
			msg := fmt.Sprintf("%s - %s - handlers.UserLogin - no such user %s", c.RealIP(), response.UUID, requestData.Login)
			return response.Error(c, http.StatusNotFound, "noSuchUser", msg)
		}
		msg := fmt.Sprintf("%s - %s - handlers.UserLogin - unable to login: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "userLoginFailed", msg)
	}

	// set user in session
	if err = c.SessionSet("username", u.Username); err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.UserLogin - c.SessionSet(username, %s) failed: %v", c.Request().RemoteAddr, response.UUID, u.Username, err)
		return response.Error(c, http.StatusInternalServerError, "sessionSetFailed", msg)
	}

	response.Data, err = json.Marshal(u)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.UserLogin - activitypub.Marshal(user) failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "userMarshalFailed", msg)
	}
	c.LogInfof("%s - %s - successful login: %s %s", c.RealIP(), response.UUID, u.Username, u.Email)

	return response.OK(c, http.StatusOK)
}

// UserLogout log out an user
func UserLogout(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewApiResponse(c.UUID)

	// expire session
	if err := c.SessionExpire(); err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.UserLogout - sessionExpire failed : %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "sessionExpireFailed", msg)
	}

	return response.OK(c, http.StatusOK)
}

// UserMe return user (auth needed)
func UserMe(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewApiResponse(c.UUID)

	user := c.Get("u")
	if user == nil {
		msg := fmt.Sprintf("%s - %s - handlers.UserMe - c.Get(u) return empty string.", c.RealIP(), response.UUID)
		return response.Error(c, http.StatusUnauthorized, "userNotInContext", msg)
	}
	var err error
	response.Data, err = json.Marshal(user)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.UserMe - activitypub.Marshal(user) failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "userMarshalFailed", msg)
	}
	return response.OK(c, http.StatusOK)
}

// a re-utiliser pour le PUT
/*
// UserPostRequest is request struct for adding user
type UserPostRequest struct {
	Username  string
	Firstname string
	Lastname  string
	Gender    user.Gender
	Email     string
	Address   string
	City      string
	State     string
	Zip       string
	Country   string
	About     string
	Locale    string // char(2)
	ShowNsfw  bool
}

// UserPostResponse is response on adding user request
type UserPostResponse struct {
	Code uint8
	Msg  string
	User user.User
}

// UserPost handle POST /api/v1/user request
func User(c echo.Context) error {
	response := UserPostResponse{}

	// get body request
	userDatas, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.Infof("%v - handlers.UserPost - unable to read request body: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer c.Request().Body.Close()

	// decode activitypub request in UserPostRequest struct
	var userpost UserPostRequest
	err = activitypub.Unmarshal(userDatas, &userpost)
	if err != nil {
		log.Infof("%v - handlers.UserPost - unable to unmarshall activitypub from body: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// validating mail address
	if _, err = mail.ParseAddress(userpost.Email); err != nil {
		log.Infof("%v - handlers.UserPost - invalid mail address: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// validating mail address
	if len(userpost.Username) < 1 {
		log.Infof("%v - handlers.UserPost - invalid username: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// creating user for saving in DB
	user := user.User{
		Username:  userpost.Username,
		Firstname: userpost.Firstname,
		Lastname:  userpost.Lastname,
		Gender:    userpost.Gender,
		Email:     userpost.Email,
		Address:   userpost.Address,
		City:      userpost.City,
		State:     userpost.State,
		Zip:       userpost.Zip,
		Country:   userpost.Country,
		About:     userpost.About,
		Locale:    userpost.Locale,
		ShowNsfw:  userpost.ShowNsfw,
		UserURL:   "",
		Admin:     false,
		AvatarURL: "",
		APIKey:    "",
	}

	if err := user.Create(); err != nil {
		log.Infof("%v - handlers.UserPost - unable to create user in DB: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	response.User = user
	return c.JSON(http.StatusCreated, response)
}
*/
