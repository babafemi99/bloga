package main

import (
	"bloga/bloga/blogapb"
	"bloga/bloga/entity"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	// "github.com/google/uuid"
	_ "github.com/lib/pq"
)

var db *sql.DB
var err error

type server struct {
	blogapb.UnimplementedBlogServiceServer
}

func (*server) CreateBlog(ctx context.Context, req *blogapb.BlogRequest) (*blogapb.Blogresponse, error) {
	fmt.Printf("started create blog function with this response from the client: %v\n", req)
	blog := req.GetBlog()
	data := entity.BlogItem{
		ID:       blog.GetId(),
		AuthorId: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}
	sqlStmt := `INSERT INTO blogs(id, author_id, content, title) VALUES ($1,$2,$3,$4)`
	_, InsertErr := db.Exec(sqlStmt, data.ID, data.AuthorId, data.Content, data.Title)
	if InsertErr != nil {
		fmt.Println(InsertErr)
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error while inserting record for create record: %v", err),
		)
	}

	return &blogapb.Blogresponse{
		Blog: &blogapb.Blog{
			Id:       blog.GetId(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetTitle(),
		},
	}, nil

}

func (*server) GetBlog(ctx context.Context, req *blogapb.GetBlogRequest) (*blogapb.GetBlogResponse, error) {
	fmt.Printf("started read blog function with this response from the client: %v\n", req)
	data := entity.BlogItem{}
	blogId := req.GetBlogId()
	res := db.QueryRow("SELECT * FROM blogs WHERE id = $1", blogId)
	if res == nil {
		fmt.Printf("Error Finding Blog with ID: %v", blogId)
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error while querying for records with ID %v", blogId),
		)
	}
	if err := res.Scan(&data.ID, &data.AuthorId, &data.Title, &data.Content); err != nil {
		fmt.Printf("error scanning: %v\n", err)
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error while scanning records into data %v", err),
		)
	}
	return &blogapb.GetBlogResponse{
		Blog: &blogapb.Blog{
			Id:       data.ID,
			AuthorId: data.AuthorId,
			Title:    data.Title,
			Content:  data.Content,
		},
	}, nil
}
func (*server) UpdateBlog(ctx context.Context, req *blogapb.UpdateBlogRequest) (*blogapb.UpdateBlogResponse, error) {
	fmt.Printf("started read blog function with this response from the client: %v\n", req)
	data := &entity.BlogItem{}
	blog := req.GetBlog()
	blogId := blog.GetId()
	data.AuthorId = blog.GetAuthorId()
	data.Content = blog.Content
	data.Title = blog.Title
	sqlStmt := `UPDATE blogs SET author_id = $1, title = $2, content = $3 WHERE id = $4`
	_, UpdateErr := db.Exec(sqlStmt, data.AuthorId, data.Title, data.Content, blogId)
	if UpdateErr != nil {
		fmt.Println("Error oh", UpdateErr)
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error while updating record : %v", err),
		)
	}
	return &blogapb.UpdateBlogResponse{
		Blog: &blogapb.Blog{
			Id:       blogId,
			AuthorId: data.AuthorId,
			Title:    data.Title,
			Content:  data.Content,
		},
	}, nil
}
func (*server) DeleteBlog(ctx context.Context, req *blogapb.DeleteBlogRequest) (*blogapb.DeleteBlogResponse, error) {
	fmt.Printf("started delete blog function with this response from the client: %v\n", req)
	blogId := req.GetBlogId()
	sqlStmt := `DELETE FROM blogs WHERE id = $1`
	_, err := db.Exec(sqlStmt, blogId)
	if err != nil {
		fmt.Println(err)
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error while deleting record : %v", err),
		)
	}
	return &blogapb.DeleteBlogResponse{
		BlogId: blogId,
	}, nil
}

func (s *server) ListBlogs(req *blogapb.ListBlogRequest, stream blogapb.BlogService_ListBlogsServer) error {
	fmt.Println("started list blog function ")
	rows, SelectErr := db.Query("SELECT * FROM blogs")
	if SelectErr != nil {
		fmt.Println(SelectErr)
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error while selcting record : %v", err),
		)
	}
	defer rows.Close()
	for rows.Next() {
		blog := entity.BlogItem{}
		if err := rows.Scan(&blog.ID, &blog.AuthorId, &blog.Title, &blog.Content); err != nil {
			fmt.Printf("error scanning: %v\n", err)
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while scanning records into data 2 %v", err),
			)
		}
		stream.Send(&blogapb.ListBlogResponse{Blog: &blogapb.Blog{
			Id:       blog.ID,
			AuthorId: blog.AuthorId,
			Title:    blog.Title,
			Content:  blog.Content,
		}})
	}
	return nil
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "aina4orosun"
	dbname   = "bloga"
)

func main() {
	// if we crash the code, we get the line and log
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// DB SETUP
	connStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	stmt := `create table if not exists blogs (id text, author_id text, content text, title text) `
	_, err = db.Exec(stmt)
	if err != nil {
		log.Printf("%q: %s\n", err, stmt)
	}
	fmt.Println("Succesfuly connected")

	fmt.Println("DB up and rnning !")

	fmt.Println("Blog Service Started")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	blogapb.RegisterBlogServiceServer(s, &server{})
	
	//Register Reflection
	reflection.Register(s)

	go func() {
		fmt.Println("Starting Server ...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to server: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	// Block till end
	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Stopping the listener")
	lis.Close()
	fmt.Println("closing postgres connection")
	fmt.Println("End of program")
}
