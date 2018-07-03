package handlers

import (
	"net/http"

	"io/ioutil"

	"encoding/json"

	"fmt"

	"database/sql"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/cmd/server/context"
	"github.com/peerpx/peerpx/entities/user"
)

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

// API

type userCreateRequest struct {
	Email    string
	Username string
	Password string
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

	user, err := user.Create(requestData.Email, requestData.Username, requestData.Password)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.UserAdd - user.Create() failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "userCreateFailed", msg)
	}

	response.Data, err = json.Marshal(user)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.UserAdd - json.Marshal(user) failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "userMarshalFailed", msg)
	}

	c.LogInfof("%s - %s - handlers.UserAdd - new user created: %s %s", c.RealIP(), response.UUID, user.Username, user.Email)
	return response.OK(c, http.StatusCreated)
}

type userLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
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
		msg := fmt.Sprintf("%s - %s - handlers.UserLogin - json.Marshal(user) failed: %v", c.RealIP(), response.UUID, err)
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
		msg := fmt.Sprintf("%s - %s - handlers.UserMe - json.Marshal(user) failed: %v", c.RealIP(), response.UUID, err)
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

	// decode json request in UserPostRequest struct
	var userpost UserPostRequest
	err = json.Unmarshal(userDatas, &userpost)
	if err != nil {
		log.Infof("%v - handlers.UserPost - unable to unmarshall json from body: %v", c.RealIP(), err)
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
