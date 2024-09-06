# json-masker
Opinionated Json masker package, allowing masking using json paths

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/AndrewHany/json-masker)

How to use:

```go
	jsonRaw := `{
		"name": "John Doe",
		"age": 30,
		"jobs": [{
			"id": 1,
			"name": "Software Engineer",
			"list": ["task1", "task2"]
		},
		{
			"id": 2,
			"name": "DevOps Engineer",
			"list": ["task1", "task2"]
		}]
	}`

	maskedPaths := []string{
		"$.name",
		"$.jobs[].name",
		"$.jobs[].list[]",
	}

	masked, err := masker.Mask(jsonRaw, maskedPaths, "[REDACTED]")
	if err != nil {
		panic(err)
	}
	println(masked)

    // Output:
    // {
    //     "age": 30,
    //     "name": "[REDACTED]",
    //     "jobs": [
    //         {
    //             "id": 1,
    //             "list": [
    //                 "[REDACTED]",
    //                 "[REDACTED]"
    //             ],
    //             "name": "[REDACTED]"
    //         },
    //         {
    //             "id": 2,
    //             "list": [
    //                 "[REDACTED]",
    //                 "[REDACTED]"
    //             ],
    //             "name": "[REDACTED]"
    //         }
    // }
```
