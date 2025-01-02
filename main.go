package main

import (
	"fmt"

	"github.com/sing3demons/http-kafka-ms/ms"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func logMiddleware(next ms.HandleFunc) ms.HandleFunc {
	return func(ctx ms.IContext) error {
		ctx.Log("Middleware")
		return next(ctx)
	}
}

func main() {
	app := ms.NewApplication()

	app.Use(logMiddleware)

	app.Get("/login/{id}/name", func(ctx ms.IContext) error {
		fmt.Println(ctx.Param("id"))
		ctx.Log("Login")
		ctx.Response(200, "Login")
		return nil
	})
	app.Start("8081")
	// ms := NewMicroservice()
	// servers := "localhost:9092"

	// Producer("login", User{Username: "user", Password: "pass"})

	// ms.StartConsume(servers, "example-group")
	// ms.Consume("login", func(ctx IContext) error {

	// 	var user User
	// 	err := ctx.ReadInput(&user)
	// 	if err != nil {
	// 		ctx.Log("Error reading input")
	// 		return err
	// 	}

	// 	fmt.Println(user)

	// 	return nil
	// })

	// ms.Start()
}
