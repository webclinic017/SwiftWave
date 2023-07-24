package dockerconfiggenerator

import (
	"errors"
	GIT "keroku/m/git_manager"
	"strings"
)

func (m Manager) DetectService(manager GIT.Manager, repo GIT.Repository) (string, error) {
	folderStructure, err := manager.FetchFolderStructure(repo)
	if err != nil {
		return "", errors.New("failed to fetch folder structure")
	}
	var lookupFiles map[string]string = map[string]string{}
	for _, lookupFile := range m.Config.LookupFiles {
		if existsInArray(folderStructure, lookupFile) {
			file, err := manager.FetchFileContent(repo, lookupFile)
			if err != nil {
				return "", errors.New("failed to fetch file content for " + lookupFile + "")
			}
			lookupFiles[lookupFile] = file
		} else {
			lookupFiles[lookupFile] = ""
		}
	}

	for _, serviceName := range m.Config.ServiceOrder {
		// Fetch service selectors
		identifiers := m.Config.Identifiers[serviceName]
		for _, identifier := range identifiers {
			// Fetch file content for each selector
			isIdentifierMatched := false
			for _, selector := range identifier.Selector {
				isMatched := true
				// Check if file content contains keywords
				for _, keyword := range selector.Keywords {
					isMatched = isMatched && strings.Contains(lookupFiles[selector.File], keyword)
				}
				isIdentifierMatched = isIdentifierMatched || isMatched
			}
			if isIdentifierMatched {
				return serviceName, nil
			}
		}
	}

	return "", errors.New("failed to detect service")
}

func (m Manager) DefaultArgs(serviceName string) map[string]string {
	args := map[string]string{}
	if _, ok := m.Config.Templates[serviceName]; !ok {
		return args
	}
	for key, variable := range m.Config.Templates[serviceName].Variables {
		args[key] = variable.Default
	}
	return args
}