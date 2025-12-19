package mcpserver

import (
	"context"

	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/review"
	"github.com/mchowning/diffguide/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SubmitReviewInput struct {
	WorkingDirectory string          `json:"workingDirectory" jsonschema:"The absolute path to the project directory"`
	Title            string          `json:"title" jsonschema:"Title of the review"`
	Sections         []model.Section `json:"sections" jsonschema:"Review sections with hunks"`
}

type SubmitReviewOutput struct {
	Success  bool   `json:"success" jsonschema:"Whether the review was stored successfully"`
	FilePath string `json:"filePath,omitempty" jsonschema:"Path where review was stored"`
	Error    string `json:"error,omitempty" jsonschema:"Error message if Success is false"`
}

type Server struct {
	reviewService *review.Service
}

func New(store *storage.Store) *Server {
	return &Server{
		reviewService: review.NewService(store),
	}
}

func (s *Server) SubmitReview(ctx context.Context, input SubmitReviewInput) (SubmitReviewOutput, error) {
	reviewData := model.Review{
		WorkingDirectory: input.WorkingDirectory,
		Title:            input.Title,
		Sections:         input.Sections,
	}

	result, err := s.reviewService.Submit(ctx, reviewData)
	if err != nil {
		return SubmitReviewOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return SubmitReviewOutput{
		Success:  true,
		FilePath: result.FilePath,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "diffguide",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "submit_review",
		Description: "Submit a code review for display in the diffguide TUI viewer",
	}, s.handleSubmitReview)

	return mcpServer.Run(ctx, &mcp.StdioTransport{})
}

func (s *Server) handleSubmitReview(ctx context.Context, req *mcp.CallToolRequest, input SubmitReviewInput) (*mcp.CallToolResult, *SubmitReviewOutput, error) {
	output, err := s.SubmitReview(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	return nil, &output, nil
}
