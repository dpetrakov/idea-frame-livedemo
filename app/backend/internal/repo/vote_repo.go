package repo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VoteRepository handles vote data persistence
type VoteRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

// NewVoteRepository creates new vote repository
func NewVoteRepository(db *pgxpool.Pool, logger *slog.Logger) *VoteRepository {
	return &VoteRepository{
		db:     db,
		logger: logger,
	}
}

// UpsertVote creates or updates a vote (value = -1 or 1)
// If value = 0, deletes the vote
func (r *VoteRepository) UpsertVote(ctx context.Context, initiativeID, userID uuid.UUID, value int) error {
	if value == 0 {
		return r.DeleteVote(ctx, initiativeID, userID)
	}

	const query = `
		INSERT INTO initiative_votes (initiative_id, user_id, value, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (initiative_id, user_id)
		DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, initiativeID, userID, value)
	if err != nil {
		r.logger.Error("failed to upsert vote", "error", err, "initiativeID", initiativeID, "userID", userID, "value", value)
		return fmt.Errorf("upsert vote: %w", err)
	}

	r.logger.Debug("vote upserted", "initiativeID", initiativeID, "userID", userID, "value", value)
	return nil
}

// DeleteVote removes a vote for the user/initiative pair
func (r *VoteRepository) DeleteVote(ctx context.Context, initiativeID, userID uuid.UUID) error {
	const query = `
		DELETE FROM initiative_votes 
		WHERE initiative_id = $1 AND user_id = $2
	`

	ct, err := r.db.Exec(ctx, query, initiativeID, userID)
	if err != nil {
		r.logger.Error("failed to delete vote", "error", err, "initiativeID", initiativeID, "userID", userID)
		return fmt.Errorf("delete vote: %w", err)
	}

	r.logger.Debug("vote deleted", "initiativeID", initiativeID, "userID", userID, "rowsAffected", ct.RowsAffected())
	return nil
}

// GetVoteAggregates returns vote aggregates for initiatives
func (r *VoteRepository) GetVoteAggregates(ctx context.Context, initiativeIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]VoteAggregates, error) {
	if len(initiativeIDs) == 0 {
		return make(map[uuid.UUID]VoteAggregates), nil
	}

	// Построим список плейсхолдеров для IN условия
	placeholders := make([]string, len(initiativeIDs))
	args := make([]interface{}, len(initiativeIDs)+1)

	for i, id := range initiativeIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	args[len(initiativeIDs)] = userID

	query := fmt.Sprintf(`
		WITH vote_stats AS (
			SELECT 
				iv.initiative_id,
				COUNT(*) FILTER (WHERE iv.value = 1) AS up_votes,
				COUNT(*) FILTER (WHERE iv.value = -1) AS down_votes
			FROM initiative_votes iv
			WHERE iv.initiative_id IN (%s)
			GROUP BY iv.initiative_id
		)
		SELECT 
			vs.initiative_id,
			COALESCE(vs.up_votes, 0) AS up_votes,
			COALESCE(vs.down_votes, 0) AS down_votes,
			COALESCE(uv.value, 0) AS current_user_vote
		FROM vote_stats vs
		LEFT JOIN initiative_votes uv ON uv.initiative_id = vs.initiative_id AND uv.user_id = $%d
	`, joinStrings(placeholders, ","), len(initiativeIDs)+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to get vote aggregates", "error", err)
		return nil, fmt.Errorf("get vote aggregates: %w", err)
	}
	defer rows.Close()

	aggregates := make(map[uuid.UUID]VoteAggregates)

	// Инициализируем все инициативы нулевыми значениями
	for _, id := range initiativeIDs {
		aggregates[id] = VoteAggregates{
			UpVotes:         0,
			DownVotes:       0,
			VoteScore:       0,
			CurrentUserVote: 0,
		}
	}

	for rows.Next() {
		var initiativeID uuid.UUID
		var upVotes, downVotes, currentUserVote int

		if err := rows.Scan(&initiativeID, &upVotes, &downVotes, &currentUserVote); err != nil {
			r.logger.Error("failed to scan vote aggregate", "error", err)
			return nil, fmt.Errorf("scan vote aggregate: %w", err)
		}

		aggregates[initiativeID] = VoteAggregates{
			UpVotes:         upVotes,
			DownVotes:       downVotes,
			VoteScore:       upVotes - downVotes,
			CurrentUserVote: currentUserVote,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vote aggregates: %w", err)
	}

	return aggregates, nil
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	var result string
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// VoteAggregates represents vote statistics for an initiative
type VoteAggregates struct {
	UpVotes         int
	DownVotes       int
	VoteScore       int
	CurrentUserVote int
}
