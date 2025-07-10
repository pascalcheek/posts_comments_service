package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/lib/pq"
	"posts_comments_service/internal/delivery/graphql"
	"posts_comments_service/internal/delivery/graphql/generated"
	"posts_comments_service/internal/domain/repositories"
	"posts_comments_service/internal/domain/services"
	"posts_comments_service/internal/repository/memory"
	"posts_comments_service/internal/repository/postgres"
)

const defaultPort = "8080"

func main() {
	storeType := flag.String("store", "memory", "Storage type: 'memory' or 'postgres'")
	dsn := flag.String("dsn", "postgres://user:password@localhost:5432/db?sslmode=disable", "PostgreSQL DSN")
	port := flag.String("port", defaultPort, "Port to run server on")
	migrate := flag.Bool("migrate", false, "Run DB migrations on start")
	flag.Parse()

	var (
		postRepo    repositories.PostRepository
		commentRepo repositories.CommentRepository
	)

	switch *storeType {
	case "memory":
		postRepo = memory.NewPostRepository()
		commentRepo = memory.NewCommentRepository()
		log.Println("Using MEMORY storage")

	case "postgres":
		db, err := sql.Open("postgres", *dsn)
		if err != nil {
			log.Fatalf("PostgreSQL connection failed: %v", err)
		}
		if err = db.Ping(); err != nil {
			log.Fatalf("PostgreSQL ping failed: %v", err)
		}

		if *migrate {
			if err := postgres.RunMigrations(*dsn); err != nil {
				log.Fatalf("Migration failed: %v", err)
			}
			log.Println("Migrations applied successfully")
		}

		postRepo = postgres.NewPostRepository(db)
		commentRepo = postgres.NewCommentRepository(db)
		log.Println("Using POSTGRES storage")

	default:
		log.Fatalf("Unsupported store type: %s", *storeType)
	}

	postService := services.NewPostService(postRepo)
	commentService := services.NewCommentService(commentRepo)

	resolver := graphql.NewResolver(postService, commentService)
	executableSchema := generated.NewExecutableSchema(generated.Config{Resolvers: resolver})

	srv := handler.NewDefaultServer(executableSchema)

	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("Server started at http://localhost:%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
