package ota

func appendArgIfNotPresent(osArgs, extraArgs []string) []string {
	osArgsMap := make(map[string]any, len(osArgs))
	for _, osArg := range osArgs {
		osArgsMap[osArg] = nil
	}

	acceptedExtraArgs := make([]string, 0)

	for _, extraArg := range extraArgs {
		if _, isAlreadyOSArg := osArgsMap[extraArg]; !isAlreadyOSArg {
			acceptedExtraArgs = append(acceptedExtraArgs, extraArg)
		}
	}

	return append(osArgs, acceptedExtraArgs...)
}
