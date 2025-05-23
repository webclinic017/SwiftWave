package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.48

import (
	"context"
	"errors"
	"time"

	gitmanager "github.com/swiftwave-org/swiftwave/pkg/git_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"
)

// EnvironmentVariables is the resolver for the environmentVariables field.
func (r *applicationResolver) EnvironmentVariables(ctx context.Context, obj *model.Application) ([]*model.EnvironmentVariable, error) {
	// fetch record
	records, err := core.FindEnvironmentVariablesByApplicationId(ctx, r.ServiceManager.DbClient, obj.ID)
	if err != nil {
		return nil, err
	}
	// convert to graphql object
	var result = make([]*model.EnvironmentVariable, 0)
	for _, record := range records {
		result = append(result, environmentVariableToGraphqlObject(record))
	}
	return result, nil
}

// PersistentVolumeBindings is the resolver for the persistentVolumeBindings field.
func (r *applicationResolver) PersistentVolumeBindings(ctx context.Context, obj *model.Application) ([]*model.PersistentVolumeBinding, error) {
	// fetch record
	records, err := core.FindPersistentVolumeBindingsByApplicationId(ctx, r.ServiceManager.DbClient, obj.ID)
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

// ConfigMounts is the resolver for the configMounts field.
func (r *applicationResolver) ConfigMounts(ctx context.Context, obj *model.Application) ([]*model.ConfigMount, error) {
	// fetch record
	records, err := core.FindConfigMountsByApplicationId(ctx, r.ServiceManager.DbClient, obj.ID)
	if err != nil {
		return nil, err
	}
	// convert to graphql object
	var result = make([]*model.ConfigMount, 0)
	for _, record := range records {
		result = append(result, configMountToGraphqlObject(record))
	}
	return result, nil
}

// RealtimeInfo is the resolver for the realtimeInfo field.
func (r *applicationResolver) RealtimeInfo(ctx context.Context, obj *model.Application) (*model.RealtimeInfo, error) {
	dockerManager, err := FetchDockerManager(ctx, &r.ServiceManager.DbClient)
	if err != nil {
		return &model.RealtimeInfo{
			InfoFound:       false,
			DesiredReplicas: 0,
			RunningReplicas: 0,
			DeploymentMode:  obj.DeploymentMode,
		}, nil
	}
	runningCount, err := dockerManager.NoOfRunningTasks(obj.Name)
	if err != nil {
		return &model.RealtimeInfo{
			InfoFound:       false,
			DesiredReplicas: 0,
			RunningReplicas: 0,
			DeploymentMode:  obj.DeploymentMode,
		}, nil
	}
	var desiredReplicas = -1
	if obj.DeploymentMode == model.DeploymentModeReplicated {
		desiredReplicas = int(obj.Replicas)
	}
	return &model.RealtimeInfo{
		InfoFound:       true,
		DesiredReplicas: desiredReplicas,
		RunningReplicas: runningCount,
		DeploymentMode:  obj.DeploymentMode,
	}, nil
}

// LatestDeployment is the resolver for the latestDeployment field.
func (r *applicationResolver) LatestDeployment(ctx context.Context, obj *model.Application) (*model.Deployment, error) {
	// fetch running instance
	record, err := core.FindCurrentDeployedDeploymentByApplicationId(ctx, r.ServiceManager.DbClient, obj.ID)
	if err != nil {
		record, err = core.FindLatestDeploymentByApplicationId(ctx, r.ServiceManager.DbClient, obj.ID)
		if err != nil {
			return nil, err
		}
	}
	return deploymentToGraphqlObject(record), nil
}

// Deployments is the resolver for the deployments field.
func (r *applicationResolver) Deployments(ctx context.Context, obj *model.Application) ([]*model.Deployment, error) {
	// fetch record
	records, err := core.FindDeploymentsByApplicationId(ctx, r.ServiceManager.DbClient, obj.ID)
	if err != nil {
		return nil, err
	}
	// convert to graphql object
	var result = make([]*model.Deployment, 0)
	for _, record := range records {
		result = append(result, deploymentToGraphqlObject(record))
	}
	return result, nil
}

// IngressRules is the resolver for the ingressRules field.
func (r *applicationResolver) IngressRules(ctx context.Context, obj *model.Application) ([]*model.IngressRule, error) {
	// fetch record
	records, err := core.FindIngressRulesByApplicationID(ctx, r.ServiceManager.DbClient, obj.ID)
	if err != nil {
		return nil, err
	}
	// convert to graphql object
	var result = make([]*model.IngressRule, 0)
	for _, record := range records {
		result = append(result, ingressRuleToGraphqlObject(record))
	}
	return result, nil
}

// ApplicationGroup is the resolver for the applicationGroup field.
func (r *applicationResolver) ApplicationGroup(ctx context.Context, obj *model.Application) (*model.ApplicationGroup, error) {
	if obj.ApplicationGroupID == nil {
		return nil, nil
	}
	var record = &core.ApplicationGroup{}
	err := record.FindById(ctx, r.ServiceManager.DbClient, *obj.ApplicationGroupID)
	if err != nil {
		return nil, err
	}
	return applicationGroupToGraphqlObject(record), nil
}

// CreateApplication is the resolver for the createApplication field.
func (r *mutationResolver) CreateApplication(ctx context.Context, input model.ApplicationInput) (*model.Application, error) {
	record := applicationInputToDatabaseObject(&input)
	// create transaction
	transaction := r.ServiceManager.DbClient.Begin()
	dockerManager, err := FetchDockerManager(ctx, &r.ServiceManager.DbClient)
	if err != nil {
		return nil, err
	}
	err = record.Create(ctx, *transaction, *dockerManager, r.Config.LocalConfig.ServiceConfig.TarballDirectoryPath)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}
	err = transaction.Commit().Error
	if err != nil {
		return nil, err
	}
	// fetch latest deployment
	latestDeployment, err := core.FindLatestDeploymentByApplicationId(ctx, r.ServiceManager.DbClient, record.ID)
	if err != nil {
		return nil, errors.New("failed to fetch latest deployment")
	}
	// push build request to worker
	err = r.WorkerManager.EnqueueBuildApplicationRequest(record.ID, latestDeployment.ID)
	if err != nil {
		return nil, errors.New("failed to process application build request")
	}
	return applicationToGraphqlObject(record), nil
}

// UpdateApplication is the resolver for the updateApplication field.
func (r *mutationResolver) UpdateApplication(ctx context.Context, id string, input model.ApplicationInput) (*model.Application, error) {
	// fetch record
	var record = &core.Application{}
	err := record.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return nil, err
	}
	// convert input to database object
	var databaseObject = applicationInputToDatabaseObject(&input)
	databaseObject.ID = record.ID
	databaseObject.LatestDeployment.ApplicationID = record.ID
	if databaseObject.LatestDeployment.UpstreamType == core.UpstreamTypeGit {
		gitUsername := ""
		gitPassword := ""
		gitPrivateKey := ""
		if databaseObject.LatestDeployment.GitCredentialID != nil {
			var gitCredential core.GitCredential
			if err := gitCredential.FindById(ctx, r.ServiceManager.DbClient, *databaseObject.LatestDeployment.GitCredentialID); err != nil {
				return nil, errors.New("invalid git credential provided")
			}
			gitUsername = gitCredential.Username
			gitPassword = gitCredential.Password
			gitPrivateKey = gitCredential.SshPrivateKey
		}

		commitHash, err := gitmanager.FetchLatestCommitHash(databaseObject.LatestDeployment.GitRepositoryURL(), databaseObject.LatestDeployment.RepositoryBranch, gitUsername, gitPassword, gitPrivateKey)
		if err != nil {
			return nil, errors.New("failed to fetch latest commit hash")
		}
		databaseObject.LatestDeployment.CommitHash = commitHash
	}

	// fetch docker manager
	dockerManager, err := FetchDockerManager(ctx, &r.ServiceManager.DbClient)
	if err != nil {
		return nil, err
	}
	// update record
	result, err := databaseObject.Update(ctx, r.ServiceManager.DbClient, *dockerManager)
	if err != nil {
		return nil, err
	} else {
		if result.RebuildRequired {
			err = r.WorkerManager.EnqueueBuildApplicationRequest(record.ID, result.DeploymentId)
			if err != nil {
				return nil, errors.New("failed to process application build request")
			}
		} else if result.ReloadRequired {
			err = r.WorkerManager.EnqueueDeployApplicationRequest(record.ID, result.DeploymentId)
			if err != nil {
				return nil, errors.New("failed to process application deploy request")
			}
		}
	}
	// fetch again to get the latest record
	application := &core.Application{}
	err = application.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return nil, err
	}
	return applicationToGraphqlObject(application), nil
}

// UpdateApplicationGroup is the resolver for the updateApplicationGroup field.
func (r *mutationResolver) UpdateApplicationGroup(ctx context.Context, id string, groupID *string) (bool, error) {
	var application = &core.Application{}
	err := application.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	err = application.UpdateGroup(ctx, r.ServiceManager.DbClient, groupID)
	return err == nil, err
}

// DeleteApplication is the resolver for the deleteApplication field.
func (r *mutationResolver) DeleteApplication(ctx context.Context, id string) (bool, error) {
	// fetch record
	var record = &core.Application{}
	err := record.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	// fetch docker manager
	dockerManager, err := FetchDockerManager(ctx, &r.ServiceManager.DbClient)
	if err != nil {
		return false, err
	}
	// delete record
	err = record.SoftDelete(ctx, r.ServiceManager.DbClient, *dockerManager)
	if err != nil {
		return false, err
	}
	// push delete request to worker
	err = r.WorkerManager.EnqueueDeleteApplicationRequest(record.ID)
	if err != nil {
		return false, errors.New("failed to process application delete request")
	}
	return true, nil
}

// RebuildApplication is the resolver for the rebuildApplication field.
func (r *mutationResolver) RebuildApplication(ctx context.Context, id string) (bool, error) {
	// Start transaction
	tx := r.ServiceManager.DbClient.Begin()
	// fetch record
	var record = &core.Application{
		ID: id,
	}
	deploymentId, err := record.RebuildApplication(ctx, *tx)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	// commit transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return false, errors.New("failed to create new deployment due to database error")
	}
	// enqueue build request
	err = r.WorkerManager.EnqueueBuildApplicationRequest(record.ID, deploymentId)
	if err != nil {
		return false, errors.New("failed to queue build request")
	}
	return true, nil
}

// RestartApplication is the resolver for the restartApplication field.
func (r *mutationResolver) RestartApplication(ctx context.Context, id string) (bool, error) {
	// fetch record
	var record = &core.Application{}
	err := record.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	// service name
	serviceName := record.Name
	// fetch docker manager
	dockerManager, err := FetchDockerManager(ctx, &r.ServiceManager.DbClient)
	if err != nil {
		return false, err
	}
	// restart
	err = dockerManager.RestartService(serviceName)
	if err != nil {
		return false, err
	}
	return true, nil
}

// RegenerateWebhookToken is the resolver for the regenerateWebhookToken field.
func (r *mutationResolver) RegenerateWebhookToken(ctx context.Context, id string) (string, error) {
	// fetch record
	var record = &core.Application{
		ID: id,
	}
	// regenerate token
	err := record.RegenerateWebhookToken(ctx, r.ServiceManager.DbClient)
	if err != nil {
		return "", err
	}
	return record.WebhookToken, nil
}

// SleepApplication is the resolver for the sleepApplication field.
func (r *mutationResolver) SleepApplication(ctx context.Context, id string) (bool, error) {
	tx := r.ServiceManager.DbClient.Begin()
	// fetch record
	var record = &core.Application{}
	err := record.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	// sleep
	err = record.MarkAsSleeping(ctx, *tx)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	// commit transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return false, errors.New("failed to mark application as sleeping due to database error")
	}
	// fetch latest deployment
	latestDeployment, err := core.FindLatestDeploymentByApplicationId(ctx, r.ServiceManager.DbClient, record.ID)
	if err != nil {
		return false, errors.New("failed to fetch latest deployment")
	}
	// fire deploy request
	err = r.WorkerManager.EnqueueDeployApplicationRequest(record.ID, latestDeployment.ID)
	if err != nil {
		return false, errors.New("failed to sleep application due to internal error")
	}
	return true, nil
}

// WakeApplication is the resolver for the wakeApplication field.
func (r *mutationResolver) WakeApplication(ctx context.Context, id string) (bool, error) {
	tx := r.ServiceManager.DbClient.Begin()
	// fetch record
	var record = &core.Application{}
	err := record.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	// sleep
	err = record.MarkAsWake(ctx, *tx)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	// commit transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return false, errors.New("failed to mark application as sleeping due to database error")
	}
	// fetch latest deployment
	latestDeployment, err := core.FindLatestDeploymentByApplicationId(ctx, r.ServiceManager.DbClient, record.ID)
	if err != nil {
		return false, errors.New("failed to fetch latest deployment")
	}
	// fire deploy request
	err = r.WorkerManager.EnqueueDeployApplicationRequest(record.ID, latestDeployment.ID)
	if err != nil {
		return false, errors.New("failed to wake application due to internal error")
	}
	return true, nil
}

// Application is the resolver for the application field.
func (r *queryResolver) Application(ctx context.Context, id string) (*model.Application, error) {
	var record = &core.Application{}
	err := record.FindById(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return nil, err
	}
	return applicationToGraphqlObject(record), nil
}

// Applications is the resolver for the applications field.
func (r *queryResolver) Applications(ctx context.Context, includeGroupedApplications bool) ([]*model.Application, error) {
	var records []*core.Application
	records, err := core.FindAllApplications(ctx, r.ServiceManager.DbClient, includeGroupedApplications)
	if err != nil {
		return nil, err
	}
	var result = make([]*model.Application, 0)
	for _, record := range records {
		result = append(result, applicationToGraphqlObject(record))
	}
	return result, nil
}

// IsExistApplicationName is the resolver for the isExistApplicationName field.
func (r *queryResolver) IsExistApplicationName(ctx context.Context, name string) (bool, error) {
	// fetch docker manager
	dockerManager, err := FetchDockerManager(ctx, &r.ServiceManager.DbClient)
	if err != nil {
		return false, err
	}
	return core.IsExistApplicationName(ctx, r.ServiceManager.DbClient, *dockerManager, name)
}

// ApplicationResourceAnalytics is the resolver for the applicationResourceAnalytics field.
func (r *queryResolver) ApplicationResourceAnalytics(ctx context.Context, id string, timeframe model.ApplicationResourceAnalyticsTimeframe) ([]*model.ApplicationResourceAnalytics, error) {
	var previousTime time.Time = time.Now()
	switch timeframe {
	case model.ApplicationResourceAnalyticsTimeframeLast1Hour:
		previousTime = time.Now().Add(-1 * time.Hour)
	case model.ApplicationResourceAnalyticsTimeframeLast3Hours:
		previousTime = time.Now().Add(-3 * time.Hour)
	case model.ApplicationResourceAnalyticsTimeframeLast6Hours:
		previousTime = time.Now().Add(-6 * time.Hour)
	case model.ApplicationResourceAnalyticsTimeframeLast12Hours:
		previousTime = time.Now().Add(-12 * time.Hour)
	case model.ApplicationResourceAnalyticsTimeframeLast24Hours:
		previousTime = time.Now().Add(-24 * time.Hour)
	case model.ApplicationResourceAnalyticsTimeframeLast7Days:
		previousTime = time.Now().Add(-7 * 24 * time.Hour)
	case model.ApplicationResourceAnalyticsTimeframeLast30Days:
		previousTime = time.Now().Add(-30 * 24 * time.Hour)
	}
	previousTimeUnix := previousTime.Unix()
	// fetch record
	records, err := core.FetchApplicationServiceResourceAnalytics(ctx, r.ServiceManager.DbClient, id, uint(previousTimeUnix))
	if err != nil {
		return nil, err
	}
	// convert to graphql object
	var result = make([]*model.ApplicationResourceAnalytics, 0)
	for _, record := range records {
		result = append(result, applicationServiceResourceStatToGraphqlObject(record))
	}
	return result, nil
}

// HealthStatus is the resolver for the HealthStatus field.
func (r *realtimeInfoResolver) HealthStatus(ctx context.Context, obj *model.RealtimeInfo) (model.HealthStatus, error) {
	if obj == nil || !obj.InfoFound {
		return model.HealthStatusUnknown, nil
	}
	if obj.DeploymentMode == model.DeploymentModeReplicated {
		if obj.DesiredReplicas == obj.RunningReplicas {
			return model.HealthStatusHealthy, nil
		} else {
			return model.HealthStatusUnhealthy, nil
		}
	}
	if obj.DeploymentMode == model.DeploymentModeGlobal {
		if obj.RunningReplicas > 0 {
			return model.HealthStatusHealthy, nil
		} else {
			return model.HealthStatusUnhealthy, nil
		}
	}
	return model.HealthStatusUnknown, nil
}

// Application returns ApplicationResolver implementation.
func (r *Resolver) Application() ApplicationResolver { return &applicationResolver{r} }

// RealtimeInfo returns RealtimeInfoResolver implementation.
func (r *Resolver) RealtimeInfo() RealtimeInfoResolver { return &realtimeInfoResolver{r} }

type applicationResolver struct{ *Resolver }
type realtimeInfoResolver struct{ *Resolver }
