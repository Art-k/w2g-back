package include

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type UserSignIn struct {
	Name     string
	Password string
}

// Token Described Roles
type Token struct {
	Model
	Token   string `gorm:"unique_index"`
	UserID  string
	RoleID  string
	Expired time.Time
}

// RefreshToken Described Roles
type RefreshToken struct {
	Model
	RefreshToken string `gorm:"unique_index"`
	UserID       string
	RoleID       string
	Expired      time.Time
}

// TokenResponse response to front end with the token and expiry time
type TokenResponse struct {
	UserID             string
	Token              string
	TokenExpire        time.Time
	RefreshToken       string
	RefreshTokenExpire time.Time
}

func APIlogOut(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":

		currentUser, err := GetUserByID(r.Context().Value("user").(string))
		if err != nil {
			ResponseInternalServerError(w)
		}

		if currentUser.ID != "" {
			var tokens []Token
			if result := Db.Where("user_id = ?", currentUser.ID).Delete(&tokens); result.Error != nil {
				ResponseInternalServerError(w)
			}

			var rTokens []RefreshToken
			if result := Db.Where("user_id = ?", currentUser.ID).Delete(&rTokens); result.Error != nil {
				ResponseInternalServerError(w)
			}
		}

		ResponseOK(w, []byte(""))

	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func APIMe(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET":

		Authorization := r.Header.Get("Authorization")
		isUser, cUser := IsLegalUser(Authorization)
		if isUser {
			response, err := json.Marshal(&cUser)
			if err != nil {
				ResponseForbidden(w, "", "log_out")
				return
			}
			ResponseOK(w, response)
			return
		} else {
			ResponseForbidden(w, "", "log_out")
		}

	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func APIToken(w http.ResponseWriter, r *http.Request) {

	FillAnswerHeader(w)
	OptionsAnswer(w)

	switch r.Method {

	case "GET":

	case "POST":

		log.Println("POST /token")
		var usi UserSignIn
		err := json.NewDecoder(r.Body).Decode(&usi)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var currentUser User
		Db.Where("name = ?", usi.Name).Find(&currentUser)
		if currentUser.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			ResponseNotFound(w)
			return
		}

		if !currentUser.Enabled {
			ResponseBadRequest(w, nil, "user not found")
			return
		}

		if comparePasswords(currentUser.Hash, []byte(usi.Password)) {

			apiTokenResponse, _ := json.Marshal(APITokenResponse(currentUser))
			ResponseOK(w, apiTokenResponse)

			log.Println("POST /token DONE")

		} else {
			ResponseBadRequest(w, nil, "user not found")
			return
		}

	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func getPwd() []byte {
	// Prompt the user to enter a password
	fmt.Println("Enter a password")
	// We will use this to store the users input
	var pwd string
	// Read the users input
	_, err := fmt.Scan(&pwd)
	if err != nil {
		log.Println(err)
	}
	// Return the users input as a byte slice which will save us
	// from having to do this conversion later on
	return []byte(pwd)
}

// HashAndSalt странная и не понятная функция 04-12-2019
func HashAndSalt(pwd []byte) string {

	// Use GenerateFromPassword to hash & salt pwd
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func APITokenResponse(cu User) TokenResponse {
	now := time.Now()

	var TokenRec Token
	TokenRec.Token = GetHash()
	TokenRec.UserID = cu.ID
	TokenRec.RoleID = cu.Role
	TokenRec.Expired = now.AddDate(0, 0, 7)
	Db.Create(&TokenRec)

	var RTokenRec RefreshToken
	RTokenRec.RefreshToken = GetHash()
	RTokenRec.UserID = cu.ID
	RTokenRec.RoleID = cu.Role
	RTokenRec.Expired = now.AddDate(0, 0, 14)
	Db.Create(&RTokenRec)

	var TR TokenResponse
	TR.UserID = cu.ID
	TR.Token = TokenRec.Token
	TR.TokenExpire = TokenRec.Expired
	TR.RefreshToken = RTokenRec.RefreshToken
	TR.RefreshTokenExpire = RTokenRec.Expired

	return TR
}

func SetPassword(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	switch r.Method {
	case "GET":

		if params["id"] == "" {
			w.WriteHeader(http.StatusBadRequest)
			msg := "Set Password Bad Request"
			fmt.Println(msg)
			n, _ := fmt.Fprintf(w, "{\"message\" : \""+msg+"\"}")
			log.Println(n)
			return
		}

		var user User
		Db.Where("set_pass = ?", params["id"]).Find(&user)
		if user.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			msg := "Set Password Bad Request"
			fmt.Println(msg)
			n, _ := fmt.Fprintf(w, "{\"message\" : \""+msg+"\"}")
			log.Println(n)
			return
		} else {
			user.SetPass = GetHash()
			Db.Save(&user)
		}
		tmpl := template.Must(template.ParseFiles("password.html"))

		type TemplateData struct {
			Host string
			Hash string
			Code string
		}
		var data TemplateData
		data.Host = os.Getenv("HOST")
		data.Hash = params["id"]
		data.Code = user.SetPass
		w.Header().Set("content-type", "content-type: text/html;")
		tmpl.Execute(w, data)

	case "POST":
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg := "Unexpected Object Received"
			//if DEV {
			//	if r.FormValue("Code") == "" {
			//		msg = msg + "\nEmpty Code received"
			//	}
			//	if r.FormValue("pwd1") == "" {
			//		msg = msg + "\nEmpty Password received"
			//	}
			//	if r.FormValue("pwd2") == "" {
			//		msg = msg + "\nEmpty Password received"
			//	}
			//	if r.FormValue("pwd2") != r.FormValue("pwd1") {
			//		msg = msg + "\nPasswords doesn't match"
			//	}
			//}
			fmt.Println(msg)
			n, _ := fmt.Fprintf(w, "{\"message\" : \""+msg+"\"}")
			log.Println(n)
			return
		}

		var user User
		Db.Where("set_pass = ?", r.FormValue("Code")).Find(&user)
		if user.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			msg := "Unexpected Object Received"
			//if DEV {
			//	msg = msg + "\nThere is no user associated with that hash2 " + params["id"] + " in the database"
			//}
			fmt.Println(msg)
			n, _ := fmt.Fprintf(w, "{\"message\" : \""+msg+"\"}")
			log.Println(n)
			return
		}
		user.Hash = HashAndSalt([]byte(r.FormValue("pwd1")))
		user.SetPass = ""
		user.Active = true
		user.PwdChanged = time.Now()
		if result := Db.Model(&User{}).Update(&user); result.Error != nil {
			ResponseInternalServerError(w)
			return
		}
		ResponseNoContent(w)
		return
	}
}

func IsLegalUser(Auth string) (bool, User) {

	var Answer bool
	var currentToken Token
	var currentUser User

	token := strings.Replace(Auth, "Bearer ", "", -1)
	// var blankid uuid.UUID
	Db.Where("token = ?", token).Last(&currentToken)
	if currentToken.Token != "" {

		if currentToken.Expired.After(time.Now()) {

			Db.Where("id = ?", currentToken.UserID).Last(&currentUser)

			if currentUser.Name != "" {
				Answer = true
			} else {
				Answer = false
			}
		}

	}

	return Answer, currentUser

}

func Invite(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	switch r.Method {
	case "GET":
		var user User
		Db.Where("id = ?", params["id"]).Find(&user)
		if user.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			msg := "Not Found"
			fmt.Println(msg)
			n, _ := fmt.Fprintf(w, "{\"message\" : \""+msg+"\"}")
			log.Println(n)
			return
		}
		user.SetPass = GetHash()
		Db.Save(&user)
		type inviteResponse struct {
			Link string
			Hash string
		}
		var invResp inviteResponse
		invResp.Hash = user.SetPass
		invResp.Link = os.Getenv("HOST") + "/password/" + user.SetPass

		addedRecordString, err := json.Marshal(invResp)
		if err == nil {
			w.WriteHeader(http.StatusOK)
			n, _ := fmt.Fprintf(w, string(addedRecordString))
			fmt.Println(n)
			return
		}
	}
}
