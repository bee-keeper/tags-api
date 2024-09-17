# Golang CRUD API and project setup

This repo is a take home test for <>.  The task was to create a basic CRUD API in Go, to discuss some of the architectural decisions and to document and test.

## Testing

`docker compose up` and then in separate terminal tab `./test.sh`.  Or `POSTGRES_DB=test && go test ./... -cover` from within the `tags-api` container.

### Create a Tag

```
curl -i -X POST http://127.0.0.1:8080/v1/tags \
  -H "Content-Type: application/json" \
  -d '{"name": "myTag"}'
```

### List all Tags

```
curl -i http://127.0.0.1:8080/v1/tags
```

### Create Media

```
curl -i -X POST http://127.0.0.1:8080/v1/media \
  -F "Name=media1" \
  -F "File=@./static/tests/bg.png" \
  -F "Tags=[{\"Name\":\"tag1\"}, {\"Name\":\"tag2\"}]"
```

### Retrieve Media by Tag ID

```
curl -i "http://127.0.0.1:8080/v1/media?tag=<tagID>"
```

## Discuss what you would improve if given more time

For a production setup HTTPS would be essential (potentially not required though as the task only specified HTTP).  Ideally some integration tests would be also be good.  Finally there is some logic to deal with sanitising filenames and dealing with duplicate media filenames - this would need to be reworked to deal more throughly with all edge cases.

## Explain your design and technology choices in a markdown file or a separate document.
I've previously used both [Chi](https://github.com/go-chi/chi) and [GorillaMux](https://github.com/gorilla/mux) as routers but because of the simplicity of this task I went with the standard library.  I had expected [Swagger](https://github.com/swaggo/swag) to integrate well with net/http and whilst it is supported, getting this integration to work took more time than I was prepared to spend on the task.  Probably I would have been better off using Chi/Gorilla.

I chose a Docker setup to prevent having to mock DB connections and because it provides a sanitary dev environment which can easily be shared by a team on different platforms.  I used Air to speed up the dev process by compiling in realtime.

I chose Gorm as an ORM as it is the most popular/used and is much more economical than dealing with raw SQL.

## Share any additional thoughts or considerations that went into your development process.
I thought this was a fair task as it covers both coding and approach to project setup.
