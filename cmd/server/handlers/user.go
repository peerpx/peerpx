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

	"github.com/gofrs/uuid"

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
	response := NewAPIResponse(c)

	username := c.Param("username")

	// should not happen
	if username == "" {
		response.Code = "usernameIsEmpty"
		return response.KO(http.StatusBadRequest)
	}

	if _, err := user.GetByUsername(username); err != nil {
		if err == sql.ErrNoRows {
			return response.OK(http.StatusOK)
		}
		response.Log = fmt.Sprintf("handlers.UserUsernameIsAvailable - user.GetByUsername(%s) failed: %v", username, err)
		response.Code = "userGetByUsernameFail"
		return response.KO(http.StatusInternalServerError)
	}
	// username exist
	response.Code = "usernameNotAvailable"
	return response.KO(http.StatusOK)
}

// UserCreate create a new user
func UserCreate(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewAPIResponse(c)

	// get body
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		response.Log = fmt.Sprintf("handlers.UserCreate - failed to read request body: %v", err)
		response.Code = "requestBodyNotReadable"
		return response.KO(http.StatusBadRequest)

	}
	// unmarshal
	requestData := struct {
		Email    string
		Password string
		Username string
	}{}

	if err = json.Unmarshal(body, &requestData); err != nil {
		response.Log = fmt.Sprintf("handlers.UserAdd - unmarshall request body failed: %v", err)
		response.Code = "requestBodyNotValidJson"
		return response.KO(http.StatusBadRequest)
	}

	// check entries
	// todo remove space from password &&
	// todo username must be alnum

	user, err := user.Create(requestData.Email, requestData.Username, requestData.Password)
	if err != nil {
		response.Log = fmt.Sprintf("handlers.UserAdd - user.Create() failed: %v", err)
		response.Code = "userCreateFailed"
		return response.KO(http.StatusInternalServerError)
	}

	// todo set username in session
	// todo pas besoin de retourner l'user
	response.Data, err = json.Marshal(user)
	if err != nil {
		response.Log = fmt.Sprintf("handlers.UserAdd - activitypub.Marshal(user) failed: %v", err)
		response.Code = "userMarshalFailed"
		return response.KO(http.StatusInternalServerError)
	}

	response.Log = fmt.Sprintf("handlers.UserAdd - new user created: %s %s", user.Username, user.Email)
	return response.OK(http.StatusCreated)
}

// UserLogin used to login
func UserLogin(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewAPIResponse(c)

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		response.Log = fmt.Sprintf("handlers.UserLogin - unable to read request body: %v", err)
		response.Code = "requestBodyNotReadable"
		return response.KO(http.StatusBadRequest)
	}

	// unmarshall
	data := struct {
		Login    string `activitypub:"login"`
		Password string `activitypub:"password"`
	}{}

	if err = json.Unmarshal(body, &data); err != nil {
		response.Log = fmt.Sprintf("handlers.UserLogin - unable to unmarshall request body: %v", err)
		response.Code = "requestBodyNotValidJson"
		return response.KO(http.StatusBadRequest)
	}

	u, err := user.Login(data.Login, data.Password)
	if err != nil {
		if err == user.ErrNoSuchUser {
			response.Log = fmt.Sprintf("handlers.UserLogin - no such user %s", data.Login)
			response.Code = "noSuchUser"
			return response.KO(http.StatusNotFound)
		}
		response.Log = fmt.Sprintf("handlers.UserLogin - unable to login: %v", err)
		response.Code = "userLoginFailed"
		return response.KO(http.StatusInternalServerError)
	}

	// set user in session
	if err = c.SessionSet("username", u.Username); err != nil {
		response.Log = fmt.Sprintf("handlers.UserLogin - c.SessionSet(username, %s) failed: %v", u.Username, err)
		response.Code = "sessionSetFailed"
		return response.KO(http.StatusInternalServerError)
	}

	response.Data, err = json.Marshal(u)
	if err != nil {
		response.Log = fmt.Sprintf("handlers.UserLogin - json.Marshal(user) failed: %v", err)
		response.Code = "userMarshalFailed"
		return response.KO(http.StatusInternalServerError)
	}
	response.Log = fmt.Sprintf("successful login: %s %s", u.Username, u.Email)
	return response.OK(http.StatusOK)
}

// UserLogout log out an user
func UserLogout(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewAPIResponse(c)

	// expire session
	if err := c.SessionExpire(); err != nil {
		response.Log = fmt.Sprintf("handlers.UserLogout - sessionExpire failed : %v", err)
		response.Code = "sessionExpireFailed"
		return response.KO(http.StatusInternalServerError)
	}
	return response.OK(http.StatusOK)
}

// UserPasswordLost utilities to recover password
// GET -> send an email with an auth link
// POST -> reset password
func UserPasswordLost(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewAPIResponse(c)

	// GET or POST
	switch c.Request().Method {
	// ask for resetting password -> email with authlink
	case echo.GET:
		// check if user exists
		userEmail := strings.ToLower(c.Param("email"))
		// should not happen
		if userEmail == "" {
			response.Code = "paramEmpty"
			return response.KO(http.StatusBadRequest)
		}
		u, err := user.GetByEmail(userEmail)
		if err != nil {
			if err == sql.ErrNoRows {
				c.LogInfof("handlers.UserPasswordLost - no such user: %s", userEmail)
				response.Code = "noSuchUser"
				// todo throttle ?
				return response.KO(http.StatusNotFound)
			}
			response.Log = fmt.Sprintf("handlers.UserPasswordLost - user.GetByEmail(%s) fail: %v", userEmail, err)
			response.Code = "userMarshalFailed"
			return response.KO(http.StatusInternalServerError)
		}

		// generate UUID
		uid, err := uuid.NewV4()
		if err != nil {
			response.Log = fmt.Sprintf("handlers.UserPasswordLost - uuid.NewV4() fail: %v", err)
			return response.KO(http.StatusInternalServerError)
		}

		u.AuthUUID.String = uid.String()
		if err = u.Update(); err != nil {
			response.Log = fmt.Sprintf("handlers.UserPasswordLost - user.Update() fail: %v", err)
			response.Code = "userUpdateFail"
			return response.KO(http.StatusInternalServerError)
		}

		authLink := fmt.Sprintf("%s/login/%s", config.GetString("ui.baseurl"), uid)

		// todo real template with i18n support
		mailBody := fmt.Sprintf(`Hi

To reset your password click on the link below:
%s		

`, authLink)

		c.LogInfof("MAILBODY: %s", mailBody)

		// send mail

		_ = u

	case echo.POST:
		c.LogInfo("UserPasswordLost: On a un post -> todo")
	default:
		response.Message = "bad HTTP method"
		response.Code = "badHTTPMethod"
		return response.KO(http.StatusBadRequest)
	}
	return response.OK(http.StatusOK)
}

// UserMe return user (auth needed)
func UserMe(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewAPIResponse(c)

	user := c.Get("u")
	if user == nil {
		response.Log = "handlers.UserMe - c.Get(u) return empty string."
		response.Code = "userNotInContext"
		return response.KO(http.StatusUnauthorized)
	}
	var err error
	response.Data, err = json.Marshal(user)
	if err != nil {
		response.Log = fmt.Sprintf("handlers.UserMe - json.Marshal(user) failed: %v", err)
		response.Code = "userMarshalFailed"
		return response.KO(http.StatusInternalServerError)
	}
	return response.OK(http.StatusOK)
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
