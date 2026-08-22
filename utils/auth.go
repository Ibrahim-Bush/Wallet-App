package utils

import (
	"time"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"Wallet-App/model"
	"errors"
	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidPassword = errors.New("Invalid password")
	ErrInvalidToken = errors.New("Invalid token")
	ErrInvalidClaims = errors.New("Error invalid claims")
)

//define a secret key for signature.
var jwt_signature = "secret_key_4893586"

func Hash_password(password string) (string, error){
	//use bcrypt to hash password.
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	//check the result of hashing.
	if err == nil{
	return string(hashed_password), nil
	}
	return "", ErrInvalidPassword
}

func Compare_password(password, hashed string) bool {
	//compare password with hashed value.
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	//check the result.
	if err == nil{
		return true
	}
	return false
}

func Generate_token(claims model.User_claims) (string, error){
	//add required data in a map claims.
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(24 * time.Hour))
	//create token with the claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//add signature to secure token.
	signed_token, err := token.SignedString([]byte(jwt_signature))
	return signed_token, err
}

func get_secret_key(token *jwt.Token) (interface{}, error){
	return []byte(jwt_signature), nil
}

func Verify_token(input_token string) (*model.User_claims, error){
	//define a new object for claims.
	var claims model.User_claims
	//parse the token parts and get claims.
	token, err := jwt.ParseWithClaims(input_token, &claims, get_secret_key)
	//check the result.
	if err != nil{
		return nil, err
	}
	//if the token is valid.
	if token.Valid == true{
		return &claims, nil
	}

	return nil, ErrInvalidToken
}

func Get_user_claims(c *gin.Context) (*model.User_claims, error) {
	//get the user from gin context.
	value, exists := c.Get("user")
	if !exists {
		return nil, ErrInvalidClaims
	}
	//extract the user struct from interface.
	user, ok := value.(*model.User_claims)
	if !ok {
		return nil, ErrInvalidClaims
	}
	//if the struct extracted successfully.
	return user, nil
}