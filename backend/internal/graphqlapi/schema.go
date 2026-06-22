package graphqlapi

import (
	"context"
	"errors"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/graphql-go/graphql"
	"gorm.io/gorm"
)

// Handler は GraphQL Query / Mutation のリゾルバを提供する。
type Handler struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Schema() (graphql.Schema, error) {
	shoeType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Shoe",
		Fields: graphql.Fields{
			"id":    &graphql.Field{Type: graphql.Int},
			"brand": &graphql.Field{Type: graphql.String},
			"model": &graphql.Field{Type: graphql.String},
		},
	})

	trainingLogType := graphql.NewObject(graphql.ObjectConfig{
		Name: "TrainingLog",
		Fields: graphql.Fields{
			"id":            &graphql.Field{Type: graphql.Int},
			"training_date": &graphql.Field{Type: graphql.String},
			"distance":      &graphql.Field{Type: graphql.Float},
			"duration":      &graphql.Field{Type: graphql.Int},
			"pace":          &graphql.Field{Type: graphql.String},
			"kind":          &graphql.Field{Type: graphql.Int},
			"shoe_id":       &graphql.Field{Type: graphql.Int},
		},
	})

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"shoes": &graphql.Field{
				Type: graphql.NewList(shoeType),
				Args: graphql.FieldConfigArgument{
					"limit":  &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 50},
					"offset": &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 0},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					userID, err := requireUserID(p.Context)
					if err != nil {
						return nil, err
					}
					limit := intArg(p.Args, "limit", 50)
					offset := intArg(p.Args, "offset", 0)
					var shoes []model.Shoe
					err = h.db.WithContext(p.Context).
						Where("user_id = ? AND is_active = ?", userID.Uint(), true).
						Order("id DESC").
						Limit(limit).
						Offset(offset).
						Find(&shoes).Error
					if err != nil {
						return nil, err
					}
					out := make([]map[string]any, 0, len(shoes))
					for _, s := range shoes {
						out = append(out, map[string]any{
							"id": s.ID, "brand": s.Brand, "model": s.Model,
						})
					}
					return out, nil
				},
			},
			"trainingLogs": &graphql.Field{
				Type: graphql.NewList(trainingLogType),
				Args: graphql.FieldConfigArgument{
					"limit":  &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 50},
					"offset": &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 0},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					userID, err := requireUserID(p.Context)
					if err != nil {
						return nil, err
					}
					limit := intArg(p.Args, "limit", 50)
					offset := intArg(p.Args, "offset", 0)
					var logs []model.TrainingLog
					err = h.db.WithContext(p.Context).
						Where("user_id = ?", userID.Uint()).
						Order("training_date DESC, id DESC").
						Limit(limit).
						Offset(offset).
						Find(&logs).Error
					if err != nil {
						return nil, err
					}
					out := make([]map[string]any, 0, len(logs))
					for _, l := range logs {
						out = append(out, map[string]any{
							"id":            l.ID,
							"training_date": l.TrainingDate.Format("2006-01-02"),
							"distance":      l.Distance,
							"duration":      l.Duration,
							"pace":          l.Pace,
							"kind":          l.Kind,
							"shoe_id":       l.ShoeID,
						})
					}
					return out, nil
				},
			},
		},
	})

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"ping": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (any, error) {
					if _, err := requireUserID(p.Context); err != nil {
						return nil, err
					}
					return "pong", nil
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
}

func requireUserID(ctx context.Context) (domain.UserID, error) {
	id, ok := domain.UserIDFromStdContext(ctx)
	if !ok {
		return 0, errors.New("unauthorized")
	}
	return id, nil
}

func intArg(args map[string]any, key string, def int) int {
	v, ok := args[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case int:
		return n
	default:
		return def
	}
}
