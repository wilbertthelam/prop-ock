package utils

func MapStringSliceToInterfaceSlice(stringItems []string) []interface{} {
	interfaceSlice := make([]interface{}, len(stringItems))

	for index, stringItem := range stringItems {
		interfaceSlice[index] = stringItem
	}

	return interfaceSlice
}
