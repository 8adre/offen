package router

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/offen/offen/server/mailer"
	"github.com/offen/offen/server/persistence"
)

type inviteUserRequest struct {
	EmailAddress     string `json:"emailAddress"`
	ProviderPassword string `json:"password"`
	URLTemplate      string `json:"urlTemplate"`
}

func (rt *router) postInviteUser(c *gin.Context) {
	var req inviteUserRequest
	if err := c.BindJSON(&req); err != nil {
		newJSONError(
			fmt.Errorf("router: error decoding response body: %w", err),
			http.StatusBadRequest,
		).Pipe(c)
		return
	}

	accountUser, ok := c.Value(contextKeyAuth).(persistence.LoginResult)
	if !ok {
		newJSONError(
			errors.New("router: could not find account user object in request context"),
			http.StatusNotFound,
		).Pipe(c)
		return
	}

	accountID := c.Param("accountID")
	if accountID != "" {
		if !accountUser.CanAccessAccount(accountID) {
			newJSONError(
				fmt.Errorf("router: user is not allowed to access account %s", accountID),
				http.StatusUnauthorized,
			).Pipe(c)
			return
		}
	}

	result, err := rt.db.InviteUser(req.EmailAddress, accountUser.AccountUserID, req.ProviderPassword, c.Param("accountID"))
	if err != nil {
		newJSONError(
			fmt.Errorf("router: error inviting user: %w", err),
			http.StatusBadRequest,
		).Pipe(c)
		return
	}
	if !result.UserExists {
		signedCredentials, signErr := rt.cookieSigner.MaxAge(7*24*60*60).Encode("credentials", forgotPasswordCredentials{
			Token:        result.OneTimeSecret,
			EmailAddress: req.EmailAddress,
		})
		if signErr != nil {
			rt.logError(signErr, "error signing token")
			c.Status(http.StatusNoContent)
			return
		}

		resetURL := strings.Replace(req.URLTemplate, "{token}", signedCredentials, -1)
		emailBody, bodyErr := mailer.RenderForgotPasswordMessage(map[string]string{"url": resetURL})
		if bodyErr != nil {
			newJSONError(
				fmt.Errorf("router: error rendering email message: %v", err),
				http.StatusInternalServerError,
			).Pipe(c)
			return
		}
		if err := rt.mailer.Send(rt.config.SMTP.Sender, req.EmailAddress, "Reset your password", emailBody); err != nil {
			newJSONError(
				fmt.Errorf("error sending email message: %v", err),
				http.StatusInternalServerError,
			).Pipe(c)
			return
		}
	}
	c.Status(http.StatusNoContent)
}
