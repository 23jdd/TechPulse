package api

import (
	"config"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var oauthConfig *oauth2.Config

func InitOAuth(c config.Config) {
	oauthConfig = &oauth2.Config{
		ClientID:     c.ClientId,
		ClientSecret: c.ClientSecret,
		RedirectURL:  fmt.Sprintf("https://%s/callback", c.Domain),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
}

// Callback Accept  OAuth callback
func Callback(ctx *gin.Context) {
	// TODO Add State Verify
	code := ctx.Query("code")
	tokens, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	//TODO  Storage token
	fmt.Println(tokens)
}

// Auth Is Generate Redirect
func Auth(ctx *gin.Context) {
	url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}
