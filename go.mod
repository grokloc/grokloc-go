module github.com/grokloc/grokloc-go

go 1.16

replace github.com/grokloc/grokloc-go/env => ./pkg/env

require (
	github.com/stretchr/testify v1.7.0
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/tools v0.1.2 // indirect
)
