package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	desc "grpc/pkg/note_v1"
)

type NoteService struct {
	desc.UnimplementedNoteV1Server
	pool *pgxpool.Pool
}

func NewNoteService(pool *pgxpool.Pool) *NoteService {
	return &NoteService{
		pool: pool,
	}
}

func (s *NoteService) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	var id int64
	err := s.pool.QueryRow(ctx,
		`INSERT INTO note (title, body) 
		 VALUES ($1, $2) 
		 RETURNING id`,
		req.GetInfo().GetTitle(),
		req.GetInfo().GetContext(),
	).Scan(&id)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create note: %v", err)
	}

	return &desc.CreateResponse{
		Id: id,
	}, nil
}

func (s *NoteService) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	var note desc.Note
	note.Info = &desc.NoteInfo{}
	var createdAt time.Time
	var updatedAt *time.Time

	err := s.pool.QueryRow(ctx,
		`SELECT id, title, body, created_at, updated_at 
		 FROM note 
		 WHERE id = $1`,
		req.GetId(),
	).Scan(
		&note.Id,
		&note.Info.Title,
		&note.Info.Context,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "note not found: %v", err)
	}

	note.CreatedAt = timestamppb.New(createdAt)
	if updatedAt != nil {
		note.UpdatedAt = timestamppb.New(*updatedAt)
	}

	return &desc.GetResponse{
		Note: &note,
	}, nil
}

func (s *NoteService) List(ctx context.Context, req *desc.ListRequest) (*desc.ListResponse, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, title, body, created_at, updated_at 
		 FROM note 
		 ORDER BY id 
		 LIMIT $1 OFFSET $2`,
		req.GetLimit(),
		req.GetOffset(),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list notes: %v", err)
	}
	defer rows.Close()

	notes := make([]*desc.Note, 0)
	for rows.Next() {
		var note desc.Note
		note.Info = &desc.NoteInfo{}
		var createdAt time.Time
		var updatedAt *time.Time

		err = rows.Scan(
			&note.Id,
			&note.Info.Title,
			&note.Info.Context,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan note: %v", err)
		}

		note.CreatedAt = timestamppb.New(createdAt)
		if updatedAt != nil {
			note.UpdatedAt = timestamppb.New(*updatedAt)
		}
		notes = append(notes, &note)
	}

	return &desc.ListResponse{
		Notes: notes,
	}, nil
}

func (s *NoteService) Update(ctx context.Context, req *desc.UpdateRequest) (*emptypb.Empty, error) {
	result, err := s.pool.Exec(ctx,
		`UPDATE note 
		 SET title = $1, body = $2, updated_at = now() 
		 WHERE id = $3`,
		req.GetInfo().GetTitle().GetValue(),
		req.GetInfo().GetContext().GetValue(),
		req.GetId(),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update note: %v", err)
	}

	if result.RowsAffected() == 0 {
		return nil, status.Error(codes.NotFound, "note not found")
	}

	return &emptypb.Empty{}, nil
}

func (s *NoteService) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	result, err := s.pool.Exec(ctx,
		"DELETE FROM note WHERE id = $1",
		req.GetId(),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete note: %v", err)
	}

	if result.RowsAffected() == 0 {
		return nil, status.Error(codes.NotFound, "note not found")
	}

	return &emptypb.Empty{}, nil
}
