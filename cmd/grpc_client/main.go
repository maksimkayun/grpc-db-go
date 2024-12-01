package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/wrapperspb"

	desc "grpc/pkg/note_v1"
)

const address = "localhost:5001"

func main() {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	c := desc.NewNoteV1Client(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Создание нескольких заметок
	fmt.Println(color.BlueString("\n=== Creating notes ==="))
	noteIDs := make([]int64, 0)
	for i := 1; i <= 3; i++ {
		createResp, err := c.Create(ctx, &desc.CreateRequest{
			Info: &desc.NoteInfo{
				Title:    fmt.Sprintf("Test Note %d", i),
				Context:  fmt.Sprintf("This is test note number %d", i),
				Author:   "Test Author",
				IsPublic: true,
			},
		})
		if err != nil {
			log.Fatalf("failed to create note: %v", err)
		}
		noteIDs = append(noteIDs, createResp.Id)
		fmt.Printf(color.GreenString("Created note with ID: %d\n"), createResp.Id)
	}

	// Получение списка всех заметок
	fmt.Println(color.BlueString("\n=== Listing all notes ==="))
	listResp, err := c.List(ctx, &desc.ListRequest{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		log.Fatalf("failed to list notes: %v", err)
	}
	for _, note := range listResp.Notes {
		fmt.Printf(color.GreenString("ID: %d, Title: %s, Content: %s, Author: %s, IsPublic: %v\n"),
			note.Id,
			note.Info.Title,
			note.Info.Context,
			note.Info.Author,
			note.Info.IsPublic)
	}

	// Получение конкретной заметки
	if len(noteIDs) > 0 {
		fmt.Println(color.BlueString("\n=== Getting specific note ==="))
		getResp, err := c.Get(ctx, &desc.GetRequest{Id: noteIDs[0]})
		if err != nil {
			log.Fatalf("failed to get note: %v", err)
		}
		fmt.Printf(color.GreenString("Got note: ID: %d, Title: %s, Content: %s\n"),
			getResp.Note.Id,
			getResp.Note.Info.Title,
			getResp.Note.Info.Context)

		// Обновление заметки
		fmt.Println(color.BlueString("\n=== Updating note ==="))
		_, err = c.Update(ctx, &desc.UpdateRequest{
			Id: noteIDs[0],
			Info: &desc.UpdateNoteInfo{
				Title:    wrapperspb.String("Updated Title"),
				Context:  wrapperspb.String("Updated Content"),
				Author:   wrapperspb.String("Updated Author"),
				IsPublic: wrapperspb.Bool(false),
			},
		})
		if err != nil {
			log.Fatalf("failed to update note: %v", err)
		}
		fmt.Println(color.GreenString("Note updated successfully"))

		// Проверка обновленной заметки
		getResp, err = c.Get(ctx, &desc.GetRequest{Id: noteIDs[0]})
		if err != nil {
			log.Fatalf("failed to get updated note: %v", err)
		}
		fmt.Printf(color.GreenString("Updated note: ID: %d, Title: %s, Content: %s, Author: %s, IsPublic: %v\n"),
			getResp.Note.Id,
			getResp.Note.Info.Title,
			getResp.Note.Info.Context,
			getResp.Note.Info.Author,
			getResp.Note.Info.IsPublic)

		// Удаление заметки
		fmt.Println(color.BlueString("\n=== Deleting note ==="))
		_, err = c.Delete(ctx, &desc.DeleteRequest{Id: noteIDs[0]})
		if err != nil {
			log.Fatalf("failed to delete note: %v", err)
		}
		fmt.Println(color.GreenString("Note deleted successfully"))

		// Проверка удаления
		_, err = c.Get(ctx, &desc.GetRequest{Id: noteIDs[0]})
		if err != nil {
			fmt.Printf(color.YellowString("Note not found (as expected): %v\n"), err)
		} else {
			log.Fatal("Note still exists after deletion!")
		}
	}
}
