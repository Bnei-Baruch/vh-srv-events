package operationtrace

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

type OperationTrace interface {
	HandleOperationCreate(ctx *gin.Context)
	HandleOperationRevert(ctx *gin.Context)
}

type operationReq struct {
	NewEmail      *string `json:"new_email" form:"new_email" binding:"required"`
	OldEmail      *string `json:"old_email" form:"old_email" binding:"required"`
	NewKeycloakID *string `json:"new_keycloak_id" form:"new_keycloak_id"`
	OldKeycloakID *string `json:"old_keycloak_id" form:"old_keycloak_id"`
	Type          *string `json:"type" form:"type"`
}

type OperationTraceDB struct {
	db *pgxpool.Pool
}

func NewOperationTrace(db *pgxpool.Pool) OperationTrace {
	return &OperationTraceDB{
		db,
	}
}

// handleOperationRevert
func (r *OperationTraceDB) HandleOperationRevert(c *gin.Context) {

	var opr operationReq

	if err := c.Bind(&opr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if opr.NewEmail == nil || opr.OldEmail == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new email and old email missing"})
		return
	}

	revertErr := revertOperation(r, c, *opr.NewEmail, *opr.OldEmail)

	if errors.Is(revertErr, common.ErrOperationNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": revertErr.Error()})
		return
	}

	if revertErr != nil {
		_ = c.Error(fmt.Errorf("error while reverting operation: %w", revertErr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": revertErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Reverted!"})
}

func (r *OperationTraceDB) HandleOperationCreate(c *gin.Context) {

	var opr operationReq

	if err := c.Bind(&opr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if opr.Type == nil || *opr.Type != "email_update" ||
		opr.NewEmail == nil || opr.NewKeycloakID == nil {
		if opr.Type == nil || *opr.Type != "email_update" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type should be email_update"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new email or new keycloak id missing"})
		}
		return
	}

	ID, dbErr := performOperation(r, c, opr)

	if dbErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while creating grant: %w", dbErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Created!", "data": ID})
}

// sqlQuery is one parameterized statement, kept in operation_trace so a
// revert can run it later with the same arguments.
type sqlQuery struct {
	SQL  string   `json:"sql"`
	Args []string `json:"args"`
}

func (q sqlQuery) args() []interface{} {
	args := make([]interface{}, len(q.Args))
	for i, a := range q.Args {
		args[i] = a
	}
	return args
}

type QueryLog struct {
	Queries []sqlQuery `json:"queries"`
	Logs    []string   `json:"logs"`
}

type emailInput struct {
	NewEmail      *string `json:"new_email"`
	NewKeycloakID *string `json:"new_keycloak_id"`
	OldKeycloakID *string `json:"old_keycloak_id"`
	OldEmail      *string `json:"old_email"`
}

func performOperation(r *OperationTraceDB, ctx *gin.Context, req operationReq) (int, error) {

	newKcId := req.NewKeycloakID
	oldKcId := req.OldKeycloakID
	newEmail := req.NewEmail
	oldEmail := req.OldEmail

	var output QueryLog
	var input emailInput
	var revert QueryLog

	var query, revertQuery sqlQuery

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if oldKcId == nil {
		sql := `UPDATE participant SET email = $1 WHERE email = $2 AND keycloak_id = $3`
		query = sqlQuery{SQL: sql, Args: []string{*newEmail, *oldEmail, *newKcId}}
		revertQuery = sqlQuery{SQL: sql, Args: []string{*oldEmail, *newEmail, *newKcId}}
	} else {
		sql := `UPDATE participant SET keycloak_id = $1, email = $2 WHERE keycloak_id = $3`
		query = sqlQuery{SQL: sql, Args: []string{*newKcId, *newEmail, *oldKcId}}
		revertQuery = sqlQuery{SQL: sql, Args: []string{*oldKcId, *oldEmail, *newKcId}}
	}

	updatedRes, err := tx.Exec(ctx, query.SQL, query.args()...)

	if err != nil {
		return 0, fmt.Errorf("problem updating users: %w", err)
	}

	output.Queries = append(output.Queries, query)
	output.Logs = append(output.Logs, updatedRes.String())

	input.NewEmail = newEmail
	input.NewKeycloakID = newKcId
	input.OldKeycloakID = oldKcId
	input.OldEmail = req.OldEmail

	revert.Queries = append(revert.Queries, revertQuery)
	revert.Logs = []string{}

	// pgx encodes the structs as JSON for the json columns.
	var ID int
	if err := tx.QueryRow(ctx, `INSERT INTO operation_trace (input, output, revert, status, type)
		VALUES ($1, $2, $3, 'success', 'email_update') RETURNING id`,
		input, output, revert).Scan(&ID); err != nil {
		return 0, fmt.Errorf("problem creating operation_trace: %w", err)
	}

	return ID, tx.Commit(ctx)
}

// revertOperation undoes the latest email_update for this email pair. The
// row is locked, so two reverts at once cannot both run.
func revertOperation(r *OperationTraceDB, ctx *gin.Context, newEmail string, oldEmail string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("r.db.Begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		id     int
		status string
		revert QueryLog
	)
	if err := tx.QueryRow(ctx, `SELECT id, status, revert FROM operation_trace
		WHERE input->>'new_email'=$1 AND input->>'old_email'=$2
		ORDER BY id DESC LIMIT 1 FOR UPDATE`, newEmail, oldEmail).Scan(&id, &status, &revert); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return common.ErrOperationNotFound
		}
		return fmt.Errorf("problem getting operation_trace: %w", err)
	}

	if status == "reverted" {
		return fmt.Errorf("operation already reverted")
	}

	for _, query := range revert.Queries {
		revertRes, err := tx.Exec(ctx, query.SQL, query.args()...)
		if err != nil {
			return fmt.Errorf("problem reverting operation: %w", err)
		}
		revert.Logs = append(revert.Logs, revertRes.String())
	}

	if _, err := tx.Exec(ctx, `UPDATE operation_trace SET revert = $1, status = 'reverted' WHERE id = $2`,
		revert, id); err != nil {
		return fmt.Errorf("problem updating operation_trace: %w", err)
	}

	return tx.Commit(ctx)
}
