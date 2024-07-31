package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gitlab.bbdev.team/vh/vh-srv-events/api/middleware"
	"gitlab.bbdev.team/vh/vh-srv-events/common"
	"net/http"
)

func (e *EventsAPI) HasAnyRole(c *gin.Context, roles ...string) bool {

	authData := c.Request.Context().Value(common.CtxAuthClaims)
	if authData == nil {
		c.Status(http.StatusForbidden)
		return false
	}
	claims := authData.(*middleware.IDTokenClaims)

	if !claims.HasAnyRole(roles...) {
		c.Status(http.StatusForbidden)
		return false
	}

	return true
}

func (e *EventsAPI) isSubjectOrHasAnyRole(c *gin.Context, keycloakID string, roles ...string) bool {

	authData := c.Request.Context().Value(common.CtxAuthClaims)
	if authData == nil {
		c.Status(http.StatusForbidden)
		return false
	}
	claims := authData.(*middleware.IDTokenClaims)

	if claims.Sub != keycloakID && !claims.HasAnyRole(roles...) {
		c.Status(http.StatusForbidden)
		return false
	}

	return true
}

func (e *EventsAPI) isEmailOwnerOrHasAnyRole(c *gin.Context, email string, roles ...string) bool {

	authData := c.Request.Context().Value(common.CtxAuthClaims)
	if authData == nil {
		c.Status(http.StatusForbidden)
		return false
	}
	claims := authData.(*middleware.IDTokenClaims)

	if claims.Email != email && !claims.HasAnyRole(roles...) {
		c.Status(http.StatusForbidden)
		return false
	}

	return true
}

func (e *EventsAPI) isUserOrHasAnyRole(c *gin.Context, userID string, roles ...string) bool {

	authData := c.Request.Context().Value(common.CtxAuthClaims)
	if authData == nil {
		c.Status(http.StatusForbidden)
		return false
	}
	claims := authData.(*middleware.IDTokenClaims)

	if claims.HasAnyRole(roles...) {
		return true
	}

	match, err := e.repo.IsSubjectID(c.Request.Context(), claims.Sub, userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.IsSubjectID: %w", err))
		return false
	}

	if !match {
		c.Status(http.StatusForbidden)
		return false
	}

	return true
}

func (e *EventsAPI) getUserKeyFromRequest(c *gin.Context) (string, bool) {

	authData := c.Request.Context().Value(common.CtxAuthClaims)
	if authData == nil {
		c.Status(http.StatusForbidden)
		return "", false
	}
	claims := authData.(*middleware.IDTokenClaims)
	if claims == nil {
		c.Status(http.StatusForbidden)
		return "", false
	}

	return claims.Sub, true
}

func (e *EventsAPI) isAuthUserOrHasAnyRole(c *gin.Context, roles ...string) (bool, bool, string) {
	var (
		keycloakID string
		ok         bool
		isAdmin    bool
	)
	keycloakID, ok = e.getUserKeyFromRequest(c)
	if ok {
		isAdmin = e.HasAnyRole(c, roles...)
	} else {
		c.Status(http.StatusForbidden)
	}
	return ok, isAdmin, keycloakID
}
