package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go-final-project/internal/db"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const layout = "20060102"

const secret = "my_secret_key"

type API struct {
	repo         *db.SchedulerRepo
	logger       *log.Logger
	passwordHash string
}

func NewAPI(repo *db.SchedulerRepo, logger *log.Logger, password string) *API {
	passwordHash := sha256.Sum256([]byte(password))
	passwordHashString := hex.EncodeToString(passwordHash[:])

	return &API{
		repo:         repo,
		logger:       logger,
		passwordHash: passwordHashString,
	}
}

type Password struct {
	Password string `json:"password"`
}

func (a *API) NextDateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// example: "/api/nextdate?now=<20060102>&date=<20060102>&repeat=<правило>"
	paramNow := r.URL.Query().Get("now")
	paramDate := r.URL.Query().Get("date")
	paramRepeat := r.URL.Query().Get("repeat")

	a.logger.Printf("request: NextDareHendler() paramNow: %v, paramDate: %v, paramRepeat: %v\n", paramNow, paramDate, paramRepeat)

	var now time.Time

	if paramNow == "" {
		now = time.Now()
	} else {
		parsedParamNow, err := time.Parse(layout, paramNow)
		if err != nil {
			a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
			return
		}
		now = parsedParamNow
	}

	nextDate, err := NextDate(now, paramDate, paramRepeat)
	if err != nil {
		a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
		return
	}

	a.logger.Printf("response: %v\n", nextDate)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", nextDate)
}

func (a *API) TaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodDelete:
		a.DeleteTaskHandler(w, r)
		return
	case http.MethodGet:
		a.GetTaskByIDHandler(w, r)
		return
	case http.MethodPost:
		a.addTaskHandler(w, r)
		return
	case http.MethodPut:
		a.UpdateTaskHandler(w, r)
		return
	}
}

func (a *API) TasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		a.GetTasksHandler(w, r)
		return
	}
}

func (a *API) WriteResponse(w http.ResponseWriter, status int, key string, value any) {
	w.WriteHeader(status)
	response := map[string]any{
		key: value,
	}
	a.logger.Printf("response: %+v\n", response)
	json.NewEncoder(w).Encode(response)
}

func (a *API) SigninHandler(w http.ResponseWriter, r *http.Request) {
	a.logger.Println("request: SigninHandler()")

	var password Password

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &password)
	if err != nil {
		a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
		return
	}

	result := sha256.Sum256([]byte(password.Password))
	hashString := hex.EncodeToString(result[:])

	if hashString != a.passwordHash {
		a.WriteResponse(w, http.StatusForbidden, "error", "wrong password")
		return
	}
	signedToken, err := createToken(hashString)
	if err != nil {
		a.WriteResponse(w, http.StatusInternalServerError, "error", err.Error())
		return
	}

	a.WriteResponse(w, http.StatusOK, "token", signedToken)
}

func createToken(hash string) (string, error) {
	secret := []byte(secret)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"hash": hash,
		"exp":  time.Now().Add(time.Hour * 8).Unix(),
		"iat":  time.Now().Unix(),
	})

	signedToken, err := jwtToken.SignedString(secret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (a *API) Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(a.passwordHash) > 0 {
			var token string
			cookie, err := r.Cookie("token")
			if err == nil {
				token = cookie.Value
			}

			jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(secret), nil 
			})
			if err != nil {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			if !jwtToken.Valid {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			claims, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			hash := claims["hash"].(string)
			if hash == "" {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			exp := claims["exp"].(float64)

			expTime := time.Unix(int64(exp), 0)

			if time.Now().After(expTime){
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
