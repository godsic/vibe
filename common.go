package main

func year(date string) string {
	if len(date) == 0 {
		return ""
	}
	return date[0:4]
}
