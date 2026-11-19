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
	if errors.Is(revertErr, common.ErrOperationReverted) || errors.Is(revertErr, common.ErrOperationNoMatch) ||
		errors.Is(revertErr, common.ErrOperationConflict) {
		c.JSON(http.StatusConflict, gin.H{"error": revertErr.Error()})
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

	if errors.Is(dbErr, common.ErrOperationNoMatch) || errors.Is(dbErr, common.ErrOperationConflict) {
		c.JSON(http.StatusConflict, gin.H{"error": dbErr.Error()})
		return
	}
	if dbErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while creating grant: %w", dbErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Created!", "data": ID})
}

// sqlQuery is one parameterized statement. operation_trace keeps them as a
// record of what ran; a revert never runs SQL read back from the table.
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

// renameStatements moves fromEmail to toEmail. With fromKc set, the
// participant also moves from fromKc to toKc; without it, only the email of
// toKc's participant changes. Either way the participant must still have
// both fromEmail and the starting key. A revert is the same call with the
// sides swapped.
func renameStatements(fromEmail, toEmail string, fromKc *string, toKc string) []sqlQuery {
	if fromKc == nil {
		return []sqlQuery{{`UPDATE participant SET email = $1 WHERE email = $2 AND keycloak_id = $3`,
			[]string{toEmail, fromEmail, toKc}}}
	}
	return []sqlQuery{{`UPDATE participant SET keycloak_id = $1, email = $2 WHERE keycloak_id = $3 AND email = $4`,
		[]string{toKc, toEmail, *fromKc, fromEmail}}}
}

// revertStatements undoes renameStatements for the same input.
func revertStatements(in emailInput) []sqlQuery {
	if in.OldKeycloakID == nil {
		return renameStatements(*in.NewEmail, *in.OldEmail, nil, *in.NewKeycloakID)
	}
	return renameStatements(*in.NewEmail, *in.OldEmail, in.NewKeycloakID, *in.OldKeycloakID)
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

	input := emailInput{
		NewEmail:      req.NewEmail,
		NewKeycloakID: req.NewKeycloakID,
		OldKeycloakID: req.OldKeycloakID,
		OldEmail:      req.OldEmail,
	}
	var output QueryLog
	revert := QueryLog{Queries: revertStatements(input), Logs: []string{}}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, query := range renameStatements(*input.OldEmail, *input.NewEmail, input.OldKeycloakID, *input.NewKeycloakID) {
		updatedRes, err := tx.Exec(ctx, query.SQL, query.args()...)
		if err != nil {
			if isUniqueViolation(err) {
				return 0, fmt.Errorf("new keycloak id or email: %w", common.ErrOperationConflict)
			}
			return 0, fmt.Errorf("problem updating users: %w", err)
		}
		// Exactly one participant must start from this email and key, or nothing changes.
		if updatedRes.RowsAffected() != 1 {
			return 0, common.ErrOperationNoMatch
		}
		output.Queries = append(output.Queries, query)
		output.Logs = append(output.Logs, updatedRes.String())
	}

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
		input  emailInput
	)
	if err := tx.QueryRow(ctx, `SELECT id, status, input FROM operation_trace
		WHERE input->>'new_email'=$1 AND input->>'old_email'=$2
		ORDER BY id DESC LIMIT 1 FOR UPDATE`, newEmail, oldEmail).Scan(&id, &status, &input); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return common.ErrOperationNotFound
		}
		return fmt.Errorf("problem getting operation_trace: %w", err)
	}

	if status == "reverted" {
		return common.ErrOperationReverted
	}

	// The statements come from the code, built from the stored input.
	revert := QueryLog{Queries: revertStatements(input), Logs: []string{}}
	for _, query := range revert.Queries {
		revertRes, err := tx.Exec(ctx, query.SQL, query.args()...)
		if err != nil {
			if isUniqueViolation(err) {
				return fmt.Errorf("old keycloak id or email: %w", common.ErrOperationConflict)
			}
			return fmt.Errorf("problem reverting operation: %w", err)
		}
		if revertRes.RowsAffected() != 1 {
			return common.ErrOperationNoMatch
		}
		revert.Logs = append(revert.Logs, revertRes.String())
	}

	if _, err := tx.Exec(ctx, `UPDATE operation_trace SET revert = $1, status = 'reverted' WHERE id = $2`,
		revert, id); err != nil {
		return fmt.Errorf("problem updating operation_trace: %w", err)
	}

	return tx.Commit(ctx)
}

// isUniqueViolation: the key or email being moved to is already taken.
func isUniqueViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	return errors.As(err, &pgErr) && pgErr.SQLState() == "23505"
}
