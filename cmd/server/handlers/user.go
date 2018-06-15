package handlers

import (
	"net/http"

	"io/ioutil"

	"encoding/json"

	"fmt"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/cmd/server/middlewares"
	"github.com/peerpx/peerpx/entities/user"
	"github.com/peerpx/peerpx/services/log"
)

type userCreateRequest struct {
	Email    string
	Username string
	Password string
}

// UserCreate create a new user
func UserCreate(ac echo.Context) error {
	c := ac.(*middlewares.AppContext)
	response := NewApiResponse(c.UUID)

	// get body
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.UserCreate - failed to read request body: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusBadRequest
		response.Code = "requestBodyNotReadable"
		return c.JSON(response.HttpStatus, response)
	}

	// unmarshal
	requestData := new(userCreateRequest)
	if err = json.Unmarshal(body, requestData); err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.UserAdd - unmarshall request body failed: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusBadRequest
		response.Code = "requestBodyNotValidJson"
		return c.JSON(response.HttpStatus, response)
	}

	user, err := user.Create(requestData.Email, requestData.Username, requestData.Password)
	if err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.UserAdd - user.Create() failed: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusInternalServerError
		response.Code = "userCreateFailed"
		return c.JSON(response.HttpStatus, response)
	}
	response.Data, err = json.Marshal(user)
	if err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.UserAdd - json.Marshal(user) failed: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusInternalServerError
		response.Code = "userMarshalFailed"
		return c.JSON(response.HttpStatus, response)
	}

	response.Success = true
	response.HttpStatus = http.StatusCreated
	c.LogInfof("new user created: %s %s", user.Username, user.Email)
	return c.JSON(response.HttpStatus, response)
}

type userLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// UserLogin used to login
func UserLogin(ac echo.Context) error {
	c := ac.(*middlewares.AppContext)
	response := NewApiResponse(c.UUID)

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.UserLogin - unable to read request body: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.Code = "requestBodyNotReadable"
		response.HttpStatus = http.StatusBadRequest
		return c.JSON(response.HttpStatus, response)
	}
	// unmarshall
	requestData := new(userLoginRequest)
	if err = json.Unmarshal(body, requestData); err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.UserLogin - unable to unmarshall request body: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusBadRequest
		response.Code = "requestBodyNotValidJson"
		return c.JSON(response.HttpStatus, response)
	}

	u, err := user.Login(requestData.Login, requestData.Password)
	if err != nil {
		if err == user.ErrNoSuchUser {
			response.Message = fmt.Sprintf("%v - %s - handlers.UserLogin - no such user %s", c.RealIP(), response.UUID, requestData.Login)
			log.Info(response.Message)
			response.Code = "noSuchUser"
			response.HttpStatus = http.StatusNotFound
		} else {
			response.Message = fmt.Sprintf("%v - %s - handlers.UserLogin - unable to login: %v", c.RealIP(), response.UUID, err)
			log.Error(response.Message)
			response.Code = "userLoginFailed"
			response.HttpStatus = http.StatusInternalServerError
		}
		return c.JSON(response.HttpStatus, response)
	}

	// set user in session
	if err = c.SessionSet("username", u.Username); err != nil {
		response.Message = fmt.Sprintf("%s - %s - handlers.UserLogin - c.SessionSet(username, %s) failed: %v", c.Request().RemoteAddr, response.UUID, u.Username, err)
		log.Error(response.Message)
		response.Code = "sessionSetFailed"
		response.HttpStatus = http.StatusInternalServerError
		return c.JSON(response.HttpStatus, response)
	}

	response.Data, err = json.Marshal(u)
	if err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.UserLogin - json.Marshal(user) failed: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusInternalServerError
		response.Code = "userMarshalFailed"
		return c.JSON(response.HttpStatus, response)
	}
	c.LogInfof("successful login: %s %s", u.Username, u.Email)
	response.Success = true
	response.HttpStatus = http.StatusOK
	return c.JSON(response.HttpStatus, response)
}

// UserMe return user (auth needed)
func UserMe(ac echo.Context) error {
	c := ac.(*middlewares.AppContext)
	response := NewApiResponse(c.UUID)

	user := c.Get("u")
	if user == nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.UserMe - c.Get(u) return empty string. it should not happen!", c.RealIP(), response.UUID)
		log.Error(response.Message)
		response.HttpStatus = http.StatusUnauthorized
		response.Code = "userNotInContext"
		return c.JSON(response.HttpStatus, response)
	}
	var err error
	response.Data, err = json.Marshal(user)
	if err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.UserMe - json.Marshal(user) failed: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusInternalServerError
		response.Code = "userMarshalFailed"
		return c.JSON(response.HttpStatus, response)
	}
	response.Success = true
	return c.JSON(response.HttpStatus, response)
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
