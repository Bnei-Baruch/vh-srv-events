package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"vh-srv-event/event"
	"vh-srv-event/util"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestConnectingDatabase(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := util.GetPgxPoolDBConnection(ctx)
	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	}
	defer conn.Close()
}

func Test_get_event_by_id(t *testing.T) {

	w := httptest.NewRecorder()
	ginCtx, r := gin.CreateTestContext(w)

	ctx, cancel := context.WithTimeout(ginCtx, 5*time.Second)
	defer cancel()

	conn, err := util.GetPgxPoolDBConnection(ctx)

	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	}

	defer conn.Close()

	r.GET("/event/:id", event.NewEvent(conn).GetEventByID)

	req, _ := http.NewRequest("GET", "/event/5", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
