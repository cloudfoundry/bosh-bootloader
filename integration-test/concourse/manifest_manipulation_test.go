package integration_test

import "gopkg.in/yaml.v2"

type concourseManifestInputs struct {
	webExternalURL          string
	stemcellVersion         string
	concourseReleaseVersion string
	gardenReleaseVersion    string
	tlsBindPort             int
	tlsMode                 bool
}

type concourseManifest struct {
	Name           string                 `yaml:"name"`
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
	TLSBindPort        int                   `yaml:"tls_bind_port,omitempty"`
	TLSCert            string                `yaml:"tls_cert,omitempty"`
	TLSKey             string                `yaml:"tls_key,omitempty"`
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

const (
	tlsCert = `-----BEGIN CERTIFICATE-----
MIIE7DCCAtSgAwIBAgIBATANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQDEwtjb25j
b3Vyc2VDQTAeFw0xNjEyMTYyMjE2NThaFw0yNjEyMTYyMjE3MDFaMBYxFDASBgNV
BAMTC2NvbmNvdXJzZUNBMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA
tN8o/wTV8xozaBcyNtwray2HTt5vAZXwKC6bHBU8sCxmT+inNG/fTiM31fBGbGty
/H3+M20EsssiOT6rdDrjC20kw/8yKLzwoPGIoL3cqbU/DcxJYGyB/Q+6N653ACWf
gSYsf92NSyMJraCJyGLQLSILTBPsZGYbv/SVsN+kQMC4zX0yYoXvAYPoGpa4OUTj
JOrKDdyAWCmhbXOoxwdNqktO5sEFZNmMsscaTxwmIup0TMFDUvsFYof+EUbNAPvB
QC92TdO71TUxnHvD5C3IX3TC5p5PfdFYZLIhJHan0Q46nsfpSTCDB1xrtJoOOK7m
L/rwsTjU250ctxsQgGizLfCMQPRPy3Z8akSd4UUJfFWcw0d1WlOdC8T+pql5fT8N
tt2m/IPU8PyX8LuV8osBwtrT+35Buj8sjkttHuiG/Oq5T8MsfdxbdRSkZtuJFHIl
+KueciYEqDk2KVciJ9rp4dsO+cnfF4TQ30+k6xQXFSOR0Dlzu1cYuWGdcRg+TjxC
yGBEyOcwBr4ywupoBvWipMgUZ1L/wBJfltF0twjGaXxcY79N/GTeyGjPDJXkSl7W
UhSSTXq+LtQZTUgsG8Ot0V4A8jQ/7gSz/UHAZi82dOtwdYsn8I9LXVZgguNMYcxn
AUMd2/IrZRBop7Iy0pnF5RcTRqEgZm1/jBEK154XgqUCAwEAAaNFMEMwDgYDVR0P
AQH/BAQDAgEGMBIGA1UdEwEB/wQIMAYBAf8CAQAwHQYDVR0OBBYEFLM/O3fIHOhv
SJUlBENmtdLeq+7GMA0GCSqGSIb3DQEBCwUAA4ICAQBZHESRGnDeIpzjToAUjF+m
NCy6eNQqtTIWwT104rUAjR1xoBB8m0JB2ENzrBZJvv/BGEj65RSI6hMo6cm0osg2
yvPTzKNWAFvb+1XUwqVGzUpKi8iKiiofhz2FSjepU2/bsIeQFa00259VTUnBjmJ/
ZHxsIzh5eZuFsverwecvjSSEWUFRD5N2Jntb3ww2kD9/5emOM6eS8TX8m9fpuWY3
fyghPXsrjgPhPUk44inHwOTuIqTLg7ULnOtOpKCjgpw8KfoIWeqruH8QCivqKt9h
r8GC3Ok0pn56dicCsB5MpU1VY8phR9ngVT0V3oKuxclCAl6T6UlWPumRol/4Xe8o
KbGtiy340fBIxJQsfo7pFjkSdXUJ5LJNW5YM7a3BwX8zSpXWMozWzKkfweyJuwQG
q2v2k9kswMVskErTTKxykmp4K8U6jtPFS8CDfTZFpaTkH9eT52V3M+ii6C6alvJE
AdJEnzVKnkouUjg1s4kwlWfPSLefBctZIJW7qIUxRbr3xlCc6phVJj1+5BT1MO8h
+aN9vZrqYqa3m4xScW2aenG9rxsjQKyHbVox+7ubGWrW9bQ9PWHHJrkeud5TpY2T
YNkWfEbt6m9LEfH0LLXsb5SmOSqctb2O5p3OLGiI31t7Xvuz0eaoseK9nBjVtuLB
s+ELmkKVamfD6JdHP9QmPg==
-----END CERTIFICATE-----`

	tlsKey = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`
)

func populateManifest(baseManifest string, concourseManifestInputs concourseManifestInputs) (string, error) {
	var concourseManifest concourseManifest
	err := yaml.Unmarshal([]byte(baseManifest), &concourseManifest)
	if err != nil {
		return "", err
	}

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
		concourseManifest.InstanceGroups[i].VMType = "default"
		concourseManifest.Update["update_watch_time"] = "1000-60000"

		switch concourseManifest.InstanceGroups[i].Name {
		case "web":
			concourseManifest.InstanceGroups[i].VMExtensions = []string{"lb"}
			concourseManifest.InstanceGroups[i].Jobs[0].Properties.BasicAuthUsername = "admin"
			concourseManifest.InstanceGroups[i].Jobs[0].Properties.BasicAuthPassword = "admin"
			concourseManifest.InstanceGroups[i].Jobs[0].Properties.ExternalURL = concourseManifestInputs.webExternalURL
			if concourseManifestInputs.tlsMode {
				concourseManifest.InstanceGroups[i].Jobs[0].Properties.TLSBindPort = concourseManifestInputs.tlsBindPort
				concourseManifest.InstanceGroups[i].Jobs[0].Properties.TLSCert = tlsCert
				concourseManifest.InstanceGroups[i].Jobs[0].Properties.TLSKey = tlsKey
			}
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
