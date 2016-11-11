package fakes

import "github.com/cloudfoundry/bosh-bootloader/gcp"

type GCPService struct {
	GetProjectsServiceCall struct {
		CallCount int
		Returns   struct {
			ProjectsService gcp.ProjectsService
			Error           error
		}
	}
}

func (s *GCPService) GetProjectsService() gcp.ProjectsService {
	s.GetProjectsServiceCall.CallCount++
	return s.GetProjectsServiceCall.Returns.ProjectsService
}
