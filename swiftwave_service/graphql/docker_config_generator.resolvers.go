package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.40

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"
)

// DockerConfigGenerator is the resolver for the dockerConfigGenerator field.
func (r *queryResolver) DockerConfigGenerator(ctx context.Context, input model.DockerConfigGeneratorInput) (*model.DockerConfigGeneratorOutput, error) {
	if input.SourceType == model.DockerConfigSourceTypeSourceCode {
		if input.SourceCodeCompressedFileName == nil {
			return nil, errors.New("invalid source code provided")
		}
		filename := sanitizeFileName(*input.SourceCodeCompressedFileName)
		filename = filepath.Join(r.ServiceConfig.ServiceConfig.DataDir, filename)
		config, err := r.ServiceManager.DockerConfigGenerator.GenerateConfigFromSourceCodeTar(filename)
		if err != nil {
			return nil, errors.New("failed to generate docker config from source code")
		}
		return &model.DockerConfigGeneratorOutput{
			DockerFile:          &config.DockerFile,
			DetectedServiceName: &config.DetectedService,
			DockerBuildArgs:     convertMapToDockerConfigBuildArgs(config.Variables),
		}, nil
	} else if input.SourceType == model.DockerConfigSourceTypeGit {
		gitUsername := ""
		gitPassword := ""
		if input.GitCredentialID != nil {
			var gitCredential core.GitCredential
			if err := gitCredential.FindById(ctx, r.ServiceManager.DbClient, *input.GitCredentialID); err != nil {
				return nil, errors.New("invalid git credential provided")
			}
			gitUsername = gitCredential.Username
			gitPassword = gitCredential.Password
		}
		if input.GitProvider == nil || input.RepositoryOwner == nil || input.RepositoryName == nil || input.RepositoryBranch == nil {
			return nil, errors.New("invalid git provider, repository owner, repository name or branch provided")
		}
		gitUrl := generateGitUrl(*input.GitProvider, *input.RepositoryOwner, *input.RepositoryName)
		config, err := r.ServiceManager.DockerConfigGenerator.GenerateConfigFromGitRepository(gitUrl, *input.RepositoryBranch, gitUsername, gitPassword)
		if err != nil {
			return nil, errors.New("failed to generate docker config from git repository")
		}
		return &model.DockerConfigGeneratorOutput{
			DockerFile:          &config.DockerFile,
			DetectedServiceName: &config.DetectedService,
			DockerBuildArgs:     convertMapToDockerConfigBuildArgs(config.Variables),
		}, nil
	} else if input.SourceType == model.DockerConfigSourceTypeCustom {
		if input.CustomDockerFile == nil {
			return nil, errors.New("invalid custom docker file provided")
		}
		dockerfile := strings.ReplaceAll(*input.CustomDockerFile, "\r\n", "\n")
		dockerfile = strings.ReplaceAll(dockerfile, "\r", "\n")
		dockerfile = strings.ReplaceAll(dockerfile, "\\n", "\r\n")
		config, err := r.ServiceManager.DockerConfigGenerator.GenerateConfigFromCustomDocker(dockerfile)
		if err != nil {
			return nil, errors.New("failed to generate docker config from custom docker file")
		}
		return &model.DockerConfigGeneratorOutput{
			DockerFile:          &config.DockerFile,
			DetectedServiceName: &config.DetectedService,
			DockerBuildArgs:     convertMapToDockerConfigBuildArgs(config.Variables),
		}, nil
	} else {
		return nil, fmt.Errorf("invalid source type")
	}
}
