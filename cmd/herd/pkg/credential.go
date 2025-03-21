package pkg

import "github.com/project-safari/zebra"

// generate some credentials.
func GenerateCredential(numCrds int) []*zebra.Credentials {
	credentials := make([]*zebra.Credentials, 0, numCrds)

	for i := 0; i < numCrds; i++ {
		labels := CreateLabels()
		name := Name()

		credential := zebra.NewCredential(name, labels)

		if credential.LabelsValidate() != nil {
			credential.Labels = GroupLabels(credential.Labels, GroupVal(credential))
		}

		credentials = append(credentials, credential)
	}

	return credentials
}
