package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.57

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/iota-agency/iota-sdk/pkg/composables"
	"github.com/iota-agency/iota-sdk/pkg/serrors"

	model "github.com/iota-agency/iota-sdk/modules/warehouse/interfaces/graph/gqlmodels"
)

// CompleteInventoryCheck is the resolver for the completeInventoryCheck field.
func (r *mutationResolver) CompleteInventoryCheck(ctx context.Context, items []*model.InventoryItem) (*model.InventoryPosition, error) {
	_, err := composables.UseUser(ctx)
	if err != nil {
		graphql.AddError(ctx, serrors.UnauthorizedGQLError(graphql.GetPath(ctx)))
		return nil, nil
	}
	panic(fmt.Errorf("not implemented: CompleteInventoryCheck - completeInventoryCheck"))
}

// Inventory is the resolver for the inventory field.
func (r *queryResolver) Inventory(ctx context.Context) ([]*model.InventoryPosition, error) {
	_, err := composables.UseUser(ctx)
	if err != nil {
		graphql.AddError(ctx, serrors.UnauthorizedGQLError(graphql.GetPath(ctx)))
		return nil, nil
	}
	positions, err := r.inventoryService.Positions(ctx)
	if err != nil {
		return nil, err
	}
	return InventoryPositionsToGraphModel(positions), nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
