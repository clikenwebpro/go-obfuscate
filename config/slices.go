package config

func duplicatesInSlice(arr []string) []string {
	visited := make(map[string]int, 0)
	duplicates := make([]string, 0)
	for i := 0; i < len(arr); i++ {
		visited[arr[i]] = visited[arr[i]] + 1
		if visited[arr[i]] == 2 {
			duplicates = append(duplicates, arr[i])
		}
	}
	return duplicates
}

func contains(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
