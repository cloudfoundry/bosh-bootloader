package integration_test

import "gopkg.in/yaml.v2"

type concourseManifestInputs struct {
	boshDirectorUUID        string
	webExternalURL          string
	stemcellVersion         string
	concourseReleaseVersion string
	gardenReleaseVersion    string
	bindPort                int
}

type concourseManifest struct {
	Name           string                 `yaml:"name"`
	DirectorUUID   string                 `yaml:"director_uuid"`
	Releases       []map[string]string    `yaml:"releases"`
	Stemcells      []map[string]string    `yaml:"stemcells"`
	InstanceGroups []instanceGroup        `yaml:"instance_groups"`
	Update         map[string]interface{} `yaml:"update"`
}

type instanceGroup struct {
	Name               string              `yaml:"name"`
	Instances          int                 `yaml:"instances"`
	VMType             string              `yaml:"vm_type"`
	VMExtensions       []string            `yaml:"vm_extensions,omitempty"`
	Stemcell           string              `yaml:"stemcell"`
	AZs                []string            `yaml:"azs"`
	Networks           []map[string]string `yaml:"networks"`
	Jobs               []job               `yaml:"jobs"`
	PersistentDiskType string              `yaml:"persistent_disk_type,omitempty"`
}

type properties struct {
	ExternalURL        string                `yaml:"external_url,omitempty"`
	BindPort           int                   `yaml:"bind_port,omitempty"`
	BasicAuthUsername  string                `yaml:"basic_auth_username,omitempty"`
	BasicAuthPassword  string                `yaml:"basic_auth_password,omitempty"`
	PostgreSQLDatabase interface{}           `yaml:"postgresql_database,omitempty"`
	Databases          []*propertiesDatabase `yaml:"databases,omitempty"`
	Garden             map[string]string     `yaml:"garden,omitempty"`
}

type propertiesDatabase struct {
	Name     string `yaml:"name"`
	Role     string `yaml:"role"`
	Password string `yaml:"password"`
}

type job struct {
	Name       string     `yaml:"name"`
	Release    string     `yaml:"release"`
	Properties properties `yaml:"properties"`
}

func populateManifest(baseManifest string, concourseManifestInputs concourseManifestInputs) (string, error) {
	var concourseManifest concourseManifest
	err := yaml.Unmarshal([]byte(baseManifest), &concourseManifest)
	if err != nil {
		return "", err
	}

	concourseManifest.DirectorUUID = concourseManifestInputs.boshDirectorUUID
	concourseManifest.Stemcells[0]["version"] = concourseManifestInputs.stemcellVersion

	for releaseIdx, release := range concourseManifest.Releases {
		switch release["name"] {
		case "concourse":
			concourseManifest.Releases[releaseIdx]["version"] = concourseManifestInputs.concourseReleaseVersion
		case "garden-runc":
			concourseManifest.Releases[releaseIdx]["version"] = concourseManifestInputs.gardenReleaseVersion
		}
	}

	for i, _ := range concourseManifest.InstanceGroups {
		concourseManifest.InstanceGroups[i].VMType = "m3.medium"

		switch concourseManifest.InstanceGroups[i].Name {
		case "web":
			concourseManifest.InstanceGroups[i].VMExtensions = []string{"lb"}
			concourseManifest.InstanceGroups[i].Jobs[0].Properties.BasicAuthUsername = "admin"
			concourseManifest.InstanceGroups[i].Jobs[0].Properties.BasicAuthPassword = "admin"
			concourseManifest.InstanceGroups[i].Jobs[0].Properties.ExternalURL = concourseManifestInputs.webExternalURL
			concourseManifest.InstanceGroups[i].Jobs[0].Properties.BindPort = concourseManifestInputs.bindPort
		case "worker":
			concourseManifest.InstanceGroups[i].VMExtensions = []string{"50GB_ephemeral_disk"}
		case "db":
			concourseManifest.InstanceGroups[i].PersistentDiskType = "1GB"
			concourseManifest.InstanceGroups[i].Jobs[0].Properties.Databases[0].Role = "admin"
			concourseManifest.InstanceGroups[i].Jobs[0].Properties.Databases[0].Password = "admin"
		}
	}

	finalConcourseManifestYAML, err := yaml.Marshal(concourseManifest)
	if err != nil {
		return "", err
	}

	return string(finalConcourseManifestYAML), nil
}
