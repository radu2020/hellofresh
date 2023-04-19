# Recipe count app

## Setup instructions and run instructions

Prerequisite: 
- `go version go1.20.1 darwin/amd64`


Install dependencies:
- `go mod tidy`

Run:

- `go run main.go`

## Tests

## Supporting Notes

In this example we are going to read a JSON file and decode it. 

However, we are going to read and decode the file as in streaming fashion. 

The reason for this is because, we don't know how big the file would be. 

It could be as little as one kilobyte or as big as hundreds of megabytes or gigabytes so on. 

If we had to read the whole file into the memory, many things could have gone wrong with the application. 

That's why we are using JSON streaming here.
