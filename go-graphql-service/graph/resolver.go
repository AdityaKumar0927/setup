package graph

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/lib/pq"

	// Adjust import path to match your go.mod module name:
	"github.com/AdityaKumar0927/setup/go-graphql-service/graph/generated"
)

// Resolver holds your DB (or other dependencies).
type Resolver struct {
	DB *sql.DB
}

// NewExecutableSchema returns a GraphQL server built from your generated schema + single-file resolver.
// This is called by server.go:  schema := graph.NewExecutableSchema(db)
func NewExecutableSchema(db *sql.DB) *handler.Server {
	cfg := generated.Config{
		Resolvers: &Resolver{DB: db},
	}
	return handler.NewDefaultServer(generated.NewExecutableSchema(cfg))
}

// Because your schema has Query and Mutation, we define sub-resolvers:
func (r *Resolver) Query() generated.QueryResolver       { return &queryResolver{r} }
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type queryResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }

// ------------------- Query -------------------

// If your schema has Query { healthCheck: String! }
func (q *queryResolver) HealthCheck(ctx context.Context) (string, error) {
	log.Println("[HealthCheck] Called")
	return "OK", nil
}

// ------------------ Mutation ------------------

// If your schema has Mutation { generateQuestions(uploadID: ID!): [Question!]! }
func (m *mutationResolver) GenerateQuestions(ctx context.Context, uploadID string) ([]*generated.Question, error) {
	// 1. Gather text from DB
	rows, err := m.DB.QueryContext(ctx, `
		SELECT text_content
		FROM content_fragments
		WHERE upload_id = $1
		ORDER BY order_index ASC
	`, uploadID)
	if err != nil {
		return nil, fmt.Errorf("DB query: %w", err)
	}
	defer rows.Close()

	var sb strings.Builder
	for rows.Next() {
		var txt string
		if scanErr := rows.Scan(&txt); scanErr != nil {
			return nil, scanErr
		}
		sb.WriteString(txt + "\n")
	}
	combinedText := sb.String()
	log.Printf("[GenerateQuestions] Combined text length: %d", len(combinedText))

	// 2. Create a local variable to store the explanation string,
	// because the generated code expects a *string if answerExplanation is optional in your schema.
	explanation := "C is correct because ..."
	mockQ := &generated.Question{
		ID:                "q1",
		QuestionText:      "What is the main topic?",
		Options:           []string{"Option A", "Option B", "Option C", "Option D"},
		CorrectIndex:      2,
		AnswerExplanation: &explanation, // pass a pointer to the string
	}

	// 3. (Optional) Insert into DB with explanation
	_, insErr := m.DB.ExecContext(ctx, `
		INSERT INTO questions (id, upload_id, question_text, question_type, options, correct_index, answer_explanation)
		VALUES ($1, $2, $3, 'multiple_choice', $4, $5, $6)
	`,
		mockQ.ID,
		uploadID,
		mockQ.QuestionText,
		pq.StringArray(mockQ.Options),
		mockQ.CorrectIndex,
		explanation, // pass the actual string, not the pointer, for DB insertion
	)
	if insErr != nil {
		log.Printf("[GenerateQuestions] Insert error: %v", insErr)
	}

	// Return a slice with one question for now
	return []*generated.Question{mockQ}, nil
}
