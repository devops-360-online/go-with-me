package middlewares

import (
    jwt "github.com/appleboy/gin-jwt/v2"
    "github.com/gin-gonic/gin"
    "github.com/devops-360-online/go-with-me/internal/models"
    "gorm.io/gorm"
    "time"
)

var IdentityKey = "id"

func AuthMiddleware() (*jwt.GinJWTMiddleware, error) {
    return jwt.New(&jwt.GinJWTMiddleware{
        Realm:       "go-with-me",
        Key:         []byte("secret key"), // Replace with a secure key
        Timeout:     time.Hour,
        MaxRefresh:  time.Hour,
        IdentityKey: IdentityKey,
        Authenticator: func(c *gin.Context) (interface{}, error) {
            var loginVals struct {
                Email    string `form:"email" json:"email" binding:"required"`
                Password string `form:"password" json:"password" binding:"required"`
            }
            if err := c.ShouldBind(&loginVals); err != nil {
                return nil, jwt.ErrMissingLoginValues
            }
            email := loginVals.Email
            password := loginVals.Password

            // Get the database instance from the context
            dbInterface, exists := c.Get("db")
            if !exists {
                return nil, jwt.ErrFailedAuthentication
            }
            db := dbInterface.(*gorm.DB)

            // Retrieve the user from the database
            user, err := models.GetUserByEmail(db, email)
            if err != nil {
                return nil, jwt.ErrFailedAuthentication
            }

            // Check the password
            if !models.CheckPasswordHash(password, user.Password) {
                return nil, jwt.ErrFailedAuthentication
            }

            return user, nil
        },
        Authorizator: func(data interface{}, c *gin.Context) bool {
            // Implement authorization logic if needed
            return true
        },
        IdentityHandler: func(c *gin.Context) interface{} {
            claims := jwt.ExtractClaims(c)
            return &models.User{
                Email: claims[IdentityKey].(string),
            }
        },
        PayloadFunc: func(data interface{}) jwt.MapClaims {
            if v, ok := data.(*models.User); ok {
                return jwt.MapClaims{
                    IdentityKey: v.Email,
                }
            }
            return jwt.MapClaims{}
        },
    })
}
