package main

import (
	"bloga/bloga/blogapb"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	fmt.Println("BLog client service started")
	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer cc.Close()

	c := blogapb.NewBlogServiceClient(cc)
	fmt.Println("Creating the blog")
	blog := &blogapb.Blog{
		Id: uuid.New().String(),
		AuthorId: "Nimi",
		Title:    "My Last Blog",
		Content:  "Content of first blog",
	}
	createBlogRes, err :=  c.CreateBlog(context.Background(), &blogapb.BlogRequest{Blog: blog})
	if err != nil {
		log.Fatalf("Unexpected error from creating blog: %v", err)
	}
	fmt.Println("Created the blog", createBlogRes.GetBlog())

	blog, GetErr := c.GetBlog(context.Background(), &blogapb.GetBlogRequest{BlogId: "c20b30d2-3328-4be6-8945-bf28e580c061"})
	if GetErr != nil {
		fmt.Println("Error Finding Blog")
	}
	fmt.Println("Blog Found", blog.GetBlog())

	blog := &blogapb.Blog{
		Id:       "c20b30d2-3328-4be6-8945-bf28e580c061",
		AuthorId: "New Author",
		Title:    "My Edited Blog",
		Content:  "Content of edited blog",
	}
	UpdateRes, UpdateErr := c.UpdateBlog(context.Background(), &blogapb.UpdateBlogRequest{Blog: blog})
	if UpdateErr != nil {
		fmt.Println("Error Finding Blog")
	}
	fmt.Println("Blog Updated", UpdateRes.GetBlog())


	delRes, delErr := c.DeleteBlog(context.Background(), &blogapb.DeleteBlogRequest{BlogId: "c20b30d2-3328-4be6-8945-bf28e580c061"})
	if delErr != nil {
		fmt.Printf("Error Deleting: %v", delErr)
	}
	fmt.Println("Blog Deleted", delRes.GetBlogId())


	stream, err := c.ListBlogs(context.Background(), &blogapb.ListBlogRequest{})
	if err != nil {
		log.Fatalf("Error from ListBlog: %v", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Something happened: %v", err)
		}
		fmt.Printf("Blog-> %v\n", res.GetBlog())
	}
}
