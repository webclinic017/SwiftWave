package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.39

import (
	"context"

	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"
)

// CreatePersistentVolume is the resolver for the createPersistentVolume field.
func (r *mutationResolver) CreatePersistentVolume(ctx context.Context, input model.PersistentVolumeInput) (*model.PersistentVolume, error) {
	record := persistentVolumeInputToDatabaseObject(&input)
	err := record.Create(ctx, r.ServiceManager.DbClient, r.ServiceManager.DockerManager)
	if err != nil {
		return nil, err
	}
	return persistentVolumeToGraphqlObject(record), nil
}

// DeletePersistentVolume is the resolver for the deletePersistentVolume field.
func (r *mutationResolver) DeletePersistentVolume(ctx context.Context, id uint) (bool, error) {
	// fetch record
	var record core.PersistentVolume
	err := record.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	// delete record
	err = record.Delete(ctx, r.ServiceManager.DbClient, r.ServiceManager.DockerManager)
	if err != nil {
		return false, err
	}
	return true, nil
}

// PersistentVolumeBindings is the resolver for the persistentVolumeBindings field.
func (r *persistentVolumeResolver) PersistentVolumeBindings(ctx context.Context, obj *model.PersistentVolume) ([]*model.PersistentVolumeBinding, error) {
	// fetch record
	records, err := core.FindPersistentVolumeBindingsByPersistentVolumeId(ctx, r.ServiceManager.DbClient, obj.ID)
	if err != nil {
		return nil, err
	}
	// convert to graphql object
	var result = make([]*model.PersistentVolumeBinding, 0)
	for _, record := range records {
		result = append(result, persistentVolumeBindingToGraphqlObject(record))
	}
	return result, nil
}

// PersistentVolumes is the resolver for the persistentVolumes field.
func (r *queryResolver) PersistentVolumes(ctx context.Context) ([]*model.PersistentVolume, error) {
	records, err := core.FindAllPersistentVolumes(ctx, r.ServiceManager.DbClient)
	if err != nil {
		return nil, err
	}
	var result []*model.PersistentVolume
	for _, record := range records {
		result = append(result, persistentVolumeToGraphqlObject(record))
	}
	return result, nil
}

// PersistentVolume is the resolver for the persistentVolume field.
func (r *queryResolver) PersistentVolume(ctx context.Context, id uint) (*model.PersistentVolume, error) {
	var record core.PersistentVolume
	err := record.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return nil, err
	}
	return persistentVolumeToGraphqlObject(&record), nil
}

// IsExistPersistentVolume is the resolver for the isExistPersistentVolume field.
func (r *queryResolver) IsExistPersistentVolume(ctx context.Context, name string) (bool, error) {
	isExists, err := core.IsExistPersistentVolume(ctx, r.ServiceManager.DbClient, name, r.ServiceManager.DockerManager)
	return isExists, err
}

// PersistentVolume returns PersistentVolumeResolver implementation.
func (r *Resolver) PersistentVolume() PersistentVolumeResolver { return &persistentVolumeResolver{r} }

type persistentVolumeResolver struct{ *Resolver }