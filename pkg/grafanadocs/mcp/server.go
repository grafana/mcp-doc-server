// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

// Package mcp provides the MCP tool adapter for grafanadocs. It registers
// documentation tools on a mark3labs/mcp-go server.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grafana/hack-doc-server/pkg/grafanadocs"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server wraps a grafanadocs.Index and exposes it as MCP tools.
type Server struct {
	idx *grafanadocs.Index
}

// New creates a Server with a pre-loaded index.
func New(idx *grafanadocs.Index) *Server {
	return &Server{idx: idx}
}

// Register adds all documentation tools to the MCP server.
func (s *Server) Register(srv *server.MCPServer) {
	srv.AddTool(s.searchDocsTool(), s.handleSearchDocs)
	srv.AddTool(s.getDocTool(), s.handleGetDoc)
	srv.AddTool(s.getDocOutlineTool(), s.handleGetDocOutline)
	srv.AddTool(s.listProductsTool(), s.handleListProducts)
}

// NewMCPServer creates a fully configured MCP server with docs tools registered.
func NewMCPServer(idx *grafanadocs.Index, version string) *server.MCPServer {
	srv := server.NewMCPServer("hack-doc-server", version)
	s := New(idx)
	s.Register(srv)
	return srv
}

// Tool definitions

func (s *Server) searchDocsTool() mcp.Tool {
	return newReadOnlyTool("search_docs",
		mcp.WithDescription("Search Grafana documentation. Returns matching pages with title, URL, description, and product."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithString("product", mcp.Description("Filter results to a specific product")),
		mcp.WithNumber("limit", mcp.Description("Maximum results to return (default 5)")),
	)
}

func (s *Server) getDocTool() mcp.Tool {
	return newReadOnlyTool("get_doc",
		mcp.WithDescription("Fetch a Grafana documentation page. Returns cleaned markdown content. Supports section extraction and offset/limit paging for bounded retrieval."),
		mcp.WithString("url", mcp.Required(), mcp.Description("The grafana.com/docs/ URL to fetch")),
		mcp.WithString("section", mcp.Description("Heading text to extract (returns only that section)")),
		mcp.WithNumber("offset", mcp.Description("Line offset for paging (0-indexed)")),
		mcp.WithNumber("limit", mcp.Description("Max lines to return")),
	)
}

func (s *Server) getDocOutlineTool() mcp.Tool {
	return newReadOnlyTool("get_doc_outline",
		mcp.WithDescription("Get the heading outline of a Grafana documentation page. Use this to find section names before calling get_doc with a section parameter."),
		mcp.WithString("url", mcp.Required(), mcp.Description("The grafana.com/docs/ URL")),
	)
}

func (s *Server) listProductsTool() mcp.Tool {
	return newReadOnlyTool("list_products",
		mcp.WithDescription("List all Grafana product documentation groups with their entry counts."),
	)
}

// Handlers

func (s *Server) handleSearchDocs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	query, _ := args["query"].(string)
	if query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}

	opts := grafanadocs.SearchOpts{}
	if product, ok := args["product"].(string); ok {
		opts.Product = product
	}
	if limit, ok := args["limit"].(float64); ok && limit > 0 {
		opts.Limit = int(limit)
	}

	results := grafanadocs.Search(s.idx, query, opts)
	if len(results) == 0 {
		msg := "No results found."
		if opts.Product != "" {
			msg += " Try broadening the product filter or using list_products to see available products."
		} else {
			msg += " Try different search terms."
		}
		return mcp.NewToolResultText(msg), nil
	}
	return jsonResult(results)
}

func (s *Server) handleGetDoc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	url, _ := args["url"].(string)
	if url == "" {
		return mcp.NewToolResultError("url is required"), nil
	}

	doc, err := grafanadocs.FetchDoc(ctx, url)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	opts := grafanadocs.ExcerptOpts{}
	if section, ok := args["section"].(string); ok {
		opts.Section = section
	}
	if offset, ok := args["offset"].(float64); ok {
		opts.Offset = int(offset)
	}
	if limit, ok := args["limit"].(float64); ok {
		opts.Limit = int(limit)
	}

	result := grafanadocs.Excerpt(doc, opts)

	if result.Content == "" && opts.Section != "" {
		return mcp.NewToolResultError(
			fmt.Sprintf("section %q not found. Use get_doc_outline to see available headings.", opts.Section),
		), nil
	}

	return jsonResult(map[string]any{
		"content":        result.Content,
		"url":            doc.URL,
		"total_lines":    result.Total,
		"returned_range": []int{result.Start, result.End},
	})
}

func (s *Server) handleGetDocOutline(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	url, _ := args["url"].(string)
	if url == "" {
		return mcp.NewToolResultError("url is required"), nil
	}

	doc, err := grafanadocs.FetchDoc(ctx, url)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	headings := grafanadocs.Outline(doc)
	return jsonResult(map[string]any{
		"url":      doc.URL,
		"headings": headings,
	})
}

func (s *Server) handleListProducts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return jsonResult(map[string]any{
		"products": s.idx.Products(),
	})
}

// Helpers

func newReadOnlyTool(name string, opts ...mcp.ToolOption) mcp.Tool {
	opts = append(opts,
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
	return mcp.NewTool(name, opts...)
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}
