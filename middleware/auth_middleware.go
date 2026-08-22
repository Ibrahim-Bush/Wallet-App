package middleware

import (
	"Wallet-App/utils"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidField = errors.New("Error Invalid claims field")
)

func Auth_middleware(c *gin.Context) {
	//get the authorization header at the request.
	header := c.GetHeader("Authorization")
	//if the request does not have token.
	if header == "" {
		c.JSON(401, gin.H{"error": "Missing token"})
		//refuse the request.
		c.Abort()
		return
	}
	//separate header components to get the token alone.
	parts := strings.Split(header, " ")
	//check the components.
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(401, gin.H{"error": "Invalid token format"})
		c.Abort()
		return
	}
	//extract the token.
	token := parts[1]
	//check if the token is valid and get claims.
	claims_ptr, err := utils.Verify_token(token)
	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}
	//attach claims to gin context.
	c.Set("user", claims_ptr)

	//move request to the appropriate handler.
	c.Next()
}

func Authorize_user_middleware(c *gin.Context) {

	//get the user struct from the gin context.
	user, err := utils.Get_user_claims(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "Server error"})
		c.Abort()
		return
	}
	//check the user role.
	if user.Role != "user" {
		c.JSON(403, gin.H{"error": "Admins have read-only access"})
		c.Abort()
		return
	}
	//if the role is user, move to the appropriate handler.
	c.Next()
}
