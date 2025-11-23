package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Raisondetr3/Avito-test-assignment/internal/repository/postgres"
	"github.com/Raisondetr3/Avito-test-assignment/internal/service"
	httpTransport "github.com/Raisondetr3/Avito-test-assignment/internal/transport/http"
	"github.com/Raisondetr3/Avito-test-assignment/internal/transport/http/handlers"
)

type TestSuite struct {
	db     *postgres.DB
	router http.Handler
	server *httptest.Server
}

func setupTestSuite(t *testing.T) *TestSuite {
	dbDSN := os.Getenv("TEST_DB_DSN")
	if dbDSN == "" {
		dbDSN = "postgres://reviewer_user:reviewer_pass@localhost:5432/pr_reviewer_test?sslmode=disable"
	}

	db, err := postgres.NewDB(dbDSN)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.RunMigrations("../migrations"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanupDatabase(t, db)

	userRepo := postgres.NewUserRepository(db)
	teamRepo := postgres.NewTeamRepository(db)
	prRepo := postgres.NewPRRepository(db)
	statsRepo := postgres.NewStatsRepository(db)

	reviewerAssigner := service.NewReviewerAssigner()

	userService := service.NewUserService(userRepo)
	teamService := service.NewTeamService(teamRepo)
	prService := service.NewPRService(prRepo, userRepo, reviewerAssigner)
	statsService := service.NewStatsService(statsRepo)
	bulkDeactivationService := service.NewBulkDeactivationService(userRepo, teamRepo, prRepo, reviewerAssigner)

	teamHandler := handlers.NewTeamHandler(teamService, bulkDeactivationService)
	userHandler := handlers.NewUserHandler(userService, prService)
	prHandler := handlers.NewPRHandler(prService)
	statsHandler := handlers.NewStatsHandler(statsService)

	router := httpTransport.NewRouter(teamHandler, userHandler, prHandler, statsHandler)

	return &TestSuite{
		db:     db,
		router: router,
	}
}

func (ts *TestSuite) cleanup(t *testing.T) {
	cleanupDatabase(t, ts.db)
	ts.db.Close()
}

func cleanupDatabase(t *testing.T, db *postgres.DB) {
	ctx := context.Background()
	queries := []string{
		"DELETE FROM pr_reviewers",
		"DELETE FROM pull_requests",
		"DELETE FROM users",
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			t.Logf("Warning: cleanup query failed: %v", err)
		}
	}
}

func (ts *TestSuite) request(method, path string, body any) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonData, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	return w
}

func TestCompleteWorkflow(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.cleanup(t)

	t.Run("Create team and verify members", func(t *testing.T) {
		teamReq := map[string]any{
			"team_name": "backend",
			"members": []map[string]any{
				{"user_id": "u1", "username": "Alice", "is_active": true},
				{"user_id": "u2", "username": "Bob", "is_active": true},
				{"user_id": "u3", "username": "Charlie", "is_active": true},
			},
		}

		resp := ts.request("POST", "/team/add", teamReq)
		if resp.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", resp.Code, resp.Body.String())
		}

		resp = ts.request("GET", "/team/get?team_name=backend", nil)
		if resp.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.Code)
		}

		var teamResp map[string]any
		json.Unmarshal(resp.Body.Bytes(), &teamResp)
		members := teamResp["members"].([]any)
		if len(members) != 3 {
			t.Fatalf("Expected 3 members, got %d", len(members))
		}
	})

	t.Run("Create PR and auto-assign reviewers", func(t *testing.T) {
		prReq := map[string]any{
			"pull_request_id":   "pr-1",
			"pull_request_name": "Add feature",
			"author_id":         "u1",
		}

		resp := ts.request("POST", "/pullRequest/create", prReq)
		if resp.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", resp.Code, resp.Body.String())
		}

		var prResp map[string]any
		json.Unmarshal(resp.Body.Bytes(), &prResp)
		pr := prResp["pr"].(map[string]any)
		reviewers := pr["assigned_reviewers"].([]any)

		if len(reviewers) == 0 || len(reviewers) > 2 {
			t.Fatalf("Expected 1-2 reviewers, got %d", len(reviewers))
		}

		for _, r := range reviewers {
			if r.(string) == "u1" {
				t.Fatal("Author should not be assigned as reviewer")
			}
		}
	})

	t.Run("Merge PR and verify immutability", func(t *testing.T) {
		mergeReq := map[string]any{
			"pull_request_id": "pr-1",
		}

		resp := ts.request("POST", "/pullRequest/merge", mergeReq)
		if resp.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.Code, resp.Body.String())
		}

		var mergeResp map[string]any
		json.Unmarshal(resp.Body.Bytes(), &mergeResp)
		pr := mergeResp["pr"].(map[string]any)
		if pr["status"] != "MERGED" {
			t.Fatal("PR should be marked as MERGED")
		}

		reassignReq := map[string]any{
			"pull_request_id": "pr-1",
			"old_reviewer_id": "u2",
		}

		resp = ts.request("POST", "/pullRequest/reassign", reassignReq)
		if resp.Code == http.StatusOK {
			t.Fatal("Should not allow reassignment on merged PR")
		}
	})
}

func TestReassignmentLogic(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.cleanup(t)

	teamReq := map[string]any{
		"team_name": "frontend",
		"members": []map[string]any{
			{"user_id": "f1", "username": "User1", "is_active": true},
			{"user_id": "f2", "username": "User2", "is_active": true},
			{"user_id": "f3", "username": "User3", "is_active": true},
			{"user_id": "f4", "username": "User4", "is_active": true},
		},
	}
	ts.request("POST", "/team/add", teamReq)

	prReq := map[string]any{
		"pull_request_id":   "pr-reassign",
		"pull_request_name": "Test reassign",
		"author_id":         "f1",
	}
	resp := ts.request("POST", "/pullRequest/create", prReq)

	var prResp map[string]any
	json.Unmarshal(resp.Body.Bytes(), &prResp)
	pr := prResp["pr"].(map[string]any)
	reviewers := pr["assigned_reviewers"].([]any)

	if len(reviewers) == 0 {
		t.Fatal("Expected at least one reviewer")
	}

	oldReviewerID := reviewers[0].(string)

	reassignReq := map[string]any{
		"pull_request_id": "pr-reassign",
		"old_reviewer_id": oldReviewerID,
	}

	resp = ts.request("POST", "/pullRequest/reassign", reassignReq)
	if resp.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var reassignResp map[string]any
	json.Unmarshal(resp.Body.Bytes(), &reassignResp)
	newReviewerID := reassignResp["replaced_by"].(string)

	if newReviewerID == oldReviewerID {
		t.Fatal("New reviewer should be different from old reviewer")
	}

	if newReviewerID == "f1" {
		t.Fatal("New reviewer should not be the author")
	}
}

func TestUserDeactivation(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.cleanup(t)

	teamReq := map[string]any{
		"team_name": "platform",
		"members": []map[string]any{
			{"user_id": "p1", "username": "PlatUser1", "is_active": true},
			{"user_id": "p2", "username": "PlatUser2", "is_active": true},
			{"user_id": "p3", "username": "PlatUser3", "is_active": true},
		},
	}
	ts.request("POST", "/team/add", teamReq)

	prReq := map[string]any{
		"pull_request_id":   "pr-deactivate",
		"pull_request_name": "Test deactivation",
		"author_id":         "p1",
	}
	ts.request("POST", "/pullRequest/create", prReq)

	setActiveReq := map[string]any{
		"user_id":   "p2",
		"is_active": false,
	}
	resp := ts.request("POST", "/users/setIsActive", setActiveReq)
	if resp.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.Code)
	}

	resp = ts.request("GET", "/team/get?team_name=platform", nil)
	var teamResp map[string]any
	json.Unmarshal(resp.Body.Bytes(), &teamResp)
	members := teamResp["members"].([]any)

	deactivatedCount := 0
	for _, m := range members {
		member := m.(map[string]any)
		if member["user_id"] == "p2" && member["is_active"] == false {
			deactivatedCount++
		}
	}

	if deactivatedCount != 1 {
		t.Fatal("User p2 should be deactivated")
	}
}

func TestBulkDeactivation(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.cleanup(t)

	teamReq := map[string]any{
		"team_name": "bulk-team",
		"members": []map[string]any{
			{"user_id": "b1", "username": "BulkUser1", "is_active": true},
			{"user_id": "b2", "username": "BulkUser2", "is_active": true},
			{"user_id": "b3", "username": "BulkUser3", "is_active": true},
			{"user_id": "b4", "username": "BulkUser4", "is_active": true},
		},
	}
	ts.request("POST", "/team/add", teamReq)

	ts.request("POST", "/pullRequest/create", map[string]any{
		"pull_request_id":   "pr-bulk-1",
		"pull_request_name": "Bulk PR 1",
		"author_id":         "b1",
	})

	ts.request("POST", "/pullRequest/create", map[string]any{
		"pull_request_id":   "pr-bulk-2",
		"pull_request_name": "Bulk PR 2",
		"author_id":         "b2",
	})

	bulkReq := map[string]any{
		"team_name": "bulk-team",
		"user_ids":  []string{"b1", "b2"},
	}

	start := time.Now()
	resp := ts.request("POST", "/team/deactivateUsers", bulkReq)
	duration := time.Since(start)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var bulkResp map[string]any
	json.Unmarshal(resp.Body.Bytes(), &bulkResp)

	deactivatedUsers := bulkResp["deactivated_users"].([]any)
	if len(deactivatedUsers) != 2 {
		t.Fatalf("Expected 2 deactivated users, got %d", len(deactivatedUsers))
	}

	t.Logf("Bulk deactivation took %v", duration)
	if duration > 200*time.Millisecond {
		t.Logf("Warning: Bulk deactivation took %v (target: <100ms)", duration)
	}

	resp = ts.request("GET", "/team/get?team_name=bulk-team", nil)
	var teamResp map[string]any
	json.Unmarshal(resp.Body.Bytes(), &teamResp)
	members := teamResp["members"].([]any)

	activeCount := 0
	for _, m := range members {
		member := m.(map[string]any)
		if member["is_active"] == true {
			activeCount++
		}
	}

	if activeCount != 2 {
		t.Fatalf("Expected 2 active users remaining, got %d", activeCount)
	}
}

func TestStatistics(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.cleanup(t)

	teamReq := map[string]any{
		"team_name": "stats-team",
		"members": []map[string]any{
			{"user_id": "s1", "username": "StatsUser1", "is_active": true},
			{"user_id": "s2", "username": "StatsUser2", "is_active": true},
			{"user_id": "s3", "username": "StatsUser3", "is_active": false},
		},
	}
	ts.request("POST", "/team/add", teamReq)

	ts.request("POST", "/pullRequest/create", map[string]any{
		"pull_request_id":   "pr-stats-1",
		"pull_request_name": "Stats PR 1",
		"author_id":         "s1",
	})

	ts.request("POST", "/pullRequest/create", map[string]any{
		"pull_request_id":   "pr-stats-2",
		"pull_request_name": "Stats PR 2",
		"author_id":         "s2",
	})

	ts.request("POST", "/pullRequest/merge", map[string]any{
		"pull_request_id": "pr-stats-1",
	})

	resp := ts.request("GET", "/stats", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.Code)
	}

	var stats map[string]any
	json.Unmarshal(resp.Body.Bytes(), &stats)

	prs := stats["pull_requests"].(map[string]any)
	if prs["total"].(float64) != 2 {
		t.Fatalf("Expected 2 total PRs, got %v", prs["total"])
	}
	if prs["merged"].(float64) != 1 {
		t.Fatalf("Expected 1 merged PR, got %v", prs["merged"])
	}
	if prs["open"].(float64) != 1 {
		t.Fatalf("Expected 1 open PR, got %v", prs["open"])
	}

	users := stats["users"].(map[string]any)
	if users["total"].(float64) != 3 {
		t.Fatalf("Expected 3 users, got %v", users["total"])
	}
	if users["active"].(float64) != 2 {
		t.Fatalf("Expected 2 active users, got %v", users["active"])
	}
	if users["inactive"].(float64) != 1 {
		t.Fatalf("Expected 1 inactive user, got %v", users["inactive"])
	}
}

func TestErrorHandling(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.cleanup(t)

	t.Run("Team not found", func(t *testing.T) {
		resp := ts.request("GET", "/team/get?team_name=nonexistent", nil)
		if resp.Code != http.StatusNotFound {
			t.Fatalf("Expected 404, got %d", resp.Code)
		}
	})

	t.Run("Duplicate PR", func(t *testing.T) {
		ts.request("POST", "/team/add", map[string]any{
			"team_name": "dup-team",
			"members": []map[string]any{
				{"user_id": "d1", "username": "DupUser1", "is_active": true},
				{"user_id": "d2", "username": "DupUser2", "is_active": true},
			},
		})

		prReq := map[string]any{
			"pull_request_id":   "pr-dup",
			"pull_request_name": "Duplicate PR",
			"author_id":         "d1",
		}

		resp := ts.request("POST", "/pullRequest/create", prReq)
		if resp.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d", resp.Code)
		}

		resp = ts.request("POST", "/pullRequest/create", prReq)
		if resp.Code == http.StatusCreated {
			t.Fatal("Should not allow duplicate PR")
		}
	})

	t.Run("Reassign non-assigned reviewer", func(t *testing.T) {
		ts.request("POST", "/team/add", map[string]any{
			"team_name": "reassign-team",
			"members": []map[string]any{
				{"user_id": "r1", "username": "ReassignUser1", "is_active": true},
				{"user_id": "r2", "username": "ReassignUser2", "is_active": true},
			},
		})

		ts.request("POST", "/pullRequest/create", map[string]any{
			"pull_request_id":   "pr-reassign-err",
			"pull_request_name": "Reassign Error PR",
			"author_id":         "r1",
		})

		resp := ts.request("POST", "/pullRequest/reassign", map[string]any{
			"pull_request_id": "pr-reassign-err",
			"old_reviewer_id": "r1",
		})

		if resp.Code == http.StatusOK {
			t.Fatal("Should not allow reassigning a reviewer who is not assigned")
		}
	})
}
