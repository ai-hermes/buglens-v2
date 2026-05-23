package gitlab

import (
	"context"

	"github.com/ai-hermes/buglens-v2/internal/config"

	gmcp "github.com/mark3labs/mcp-go/mcp"
)

type MCPToolRegistrar func(
	name string,
	description string,
	handler func(context.Context, map[string]any) (map[string]any, error),
	schemaOpts ...gmcp.ToolOption,
)

func RegisterMCPTools(cfg config.Config, register MCPToolRegistrar) {
	c := New(cfg)

	mustStringArray := gmcp.Items(map[string]any{"type": "string"})

	listProjectsOpts := []gmcp.ToolOption{
		gmcp.WithString("search"),
		gmcp.WithBoolean("membership"),
		gmcp.WithBoolean("owned"),
		gmcp.WithNumber("page"),
		gmcp.WithNumber("per_page"),
	}
	register("gitlab_list_projects", "List accessible GitLab projects.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return c.ListProjects(
				ctx,
				strArg(args, "search", ""),
				boolArg(args, "membership", false),
				boolArg(args, "owned", false),
				intArg(args, "page", 1),
				intArg(args, "per_page", 20),
			)
		},
		listProjectsOpts...,
	)

	searchProjectsOpts := []gmcp.ToolOption{
		gmcp.WithString("query", gmcp.Required()),
		gmcp.WithNumber("page"),
		gmcp.WithNumber("per_page"),
	}
	register("gitlab_search_projects", "Search GitLab projects by keyword.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return c.SearchProjects(
				ctx,
				strArg(args, "query", ""),
				intArg(args, "page", 1),
				intArg(args, "per_page", 20),
			)
		},
		searchProjectsOpts...,
	)

	getProjectOpts := []gmcp.ToolOption{
		gmcp.WithString("project_id", gmcp.Required()),
	}
	register("gitlab_get_project", "Get GitLab project metadata.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return c.GetProject(ctx, strArg(args, "project_id", ""))
		},
		getProjectOpts...,
	)

	getFileOpts := []gmcp.ToolOption{
		gmcp.WithString("file_path", gmcp.Required()),
		gmcp.WithString("ref"),
		gmcp.WithString("project_id"),
	}
	register("gitlab_get_file", "Get GitLab file content and metadata.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return c.GetFile(
				ctx,
				strArg(args, "file_path", ""),
				strArg(args, "ref", "main"),
				strArg(args, "project_id", ""),
			)
		},
		getFileOpts...,
	)

	listBranchesOpts := []gmcp.ToolOption{
		gmcp.WithString("project_id"),
		gmcp.WithString("search"),
	}
	register("gitlab_list_branches", "List branches in a GitLab project.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return c.ListBranches(
				ctx,
				strArg(args, "project_id", ""),
				strArg(args, "search", ""),
			)
		},
		listBranchesOpts...,
	)

	getBranchOpts := []gmcp.ToolOption{
		gmcp.WithString("branch", gmcp.Required()),
		gmcp.WithString("project_id"),
	}
	register("gitlab_get_branch", "Get branch details by branch name.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return c.GetBranch(
				ctx,
				strArg(args, "branch", ""),
				strArg(args, "project_id", ""),
			)
		},
		getBranchOpts...,
	)

	findPageCodeOpts := []gmcp.ToolOption{
		gmcp.WithString("file_path", gmcp.Required()),
		gmcp.WithNumber("line", gmcp.Required()),
		gmcp.WithString("branch"),
		gmcp.WithNumber("context"),
		gmcp.WithString("project_id"),
	}
	register("gitlab_find_page_code", "Get GitLab file snippet around a line.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return c.FindPageCode(
				ctx,
				strArg(args, "file_path", ""),
				intArg(args, "line", 1),
				strArg(args, "branch", "main"),
				intArg(args, "context", 10),
				strArg(args, "project_id", ""),
			)
		},
		findPageCodeOpts...,
	)

	getCommitsOpts := []gmcp.ToolOption{
		gmcp.WithString("file_path", gmcp.Required()),
		gmcp.WithString("branch"),
		gmcp.WithString("since"),
		gmcp.WithNumber("limit"),
		gmcp.WithString("project_id"),
	}
	register("gitlab_get_commits", "Get recent commits for a file path.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return c.GetCommits(
				ctx,
				strArg(args, "file_path", ""),
				strArg(args, "branch", "main"),
				strArg(args, "since", ""),
				intArg(args, "limit", 3),
				strArg(args, "project_id", ""),
			)
		},
		getCommitsOpts...,
	)

	listMRsOpts := []gmcp.ToolOption{
		gmcp.WithString("state"),
		gmcp.WithString("source_branch"),
		gmcp.WithString("target_branch"),
		gmcp.WithString("search"),
		gmcp.WithNumber("page"),
		gmcp.WithNumber("per_page"),
		gmcp.WithString("project_id"),
	}
	register("gitlab_list_merge_requests", "List merge requests for project.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return ListMergeRequests(ctx, c, args)
		},
		listMRsOpts...,
	)

	getMROpts := []gmcp.ToolOption{
		gmcp.WithNumber("iid", gmcp.Required()),
		gmcp.WithString("project_id"),
	}
	register("gitlab_get_merge_request", "Get merge request details by IID.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return GetMergeRequest(ctx, c, args)
		},
		getMROpts...,
	)

	createMROpts := []gmcp.ToolOption{
		gmcp.WithString("title", gmcp.Required()),
		gmcp.WithString("source_branch", gmcp.Required()),
		gmcp.WithString("target_branch", gmcp.Required()),
		gmcp.WithString("description"),
		gmcp.WithBoolean("draft"),
		gmcp.WithBoolean("remove_source_branch"),
		gmcp.WithString("project_id"),
	}
	register("gitlab_create_merge_request", "Create a merge request.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return CreateMergeRequest(ctx, c, args)
		},
		createMROpts...,
	)

	updateMROpts := []gmcp.ToolOption{
		gmcp.WithNumber("iid", gmcp.Required()),
		gmcp.WithString("title"),
		gmcp.WithString("description"),
		gmcp.WithString("target_branch"),
		gmcp.WithString("state_event"),
		gmcp.WithString("project_id"),
	}
	register("gitlab_update_merge_request", "Update merge request fields/state.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return UpdateMergeRequest(ctx, c, args)
		},
		updateMROpts...,
	)

	mergeMROpts := []gmcp.ToolOption{
		gmcp.WithNumber("iid", gmcp.Required()),
		gmcp.WithBoolean("merge_when_pipeline_succeeds"),
		gmcp.WithBoolean("should_remove_source_branch"),
		gmcp.WithBoolean("squash"),
		gmcp.WithString("project_id"),
	}
	register("gitlab_merge_merge_request", "Merge a merge request.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return MergeMergeRequest(ctx, c, args)
		},
		mergeMROpts...,
	)

	getMRChangesOpts := []gmcp.ToolOption{
		gmcp.WithNumber("iid", gmcp.Required()),
		gmcp.WithString("project_id"),
	}
	register("gitlab_get_mr_changes", "Get merge request changes/diffs.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return GetMRChanges(ctx, c, args)
		},
		getMRChangesOpts...,
	)

	getMRDiscussionsOpts := []gmcp.ToolOption{
		gmcp.WithNumber("iid", gmcp.Required()),
		gmcp.WithString("project_id"),
	}
	register("gitlab_get_mr_discussions", "Get merge request discussion threads.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return GetMRDiscussions(ctx, c, args)
		},
		getMRDiscussionsOpts...,
	)

	createMRNoteOpts := []gmcp.ToolOption{
		gmcp.WithNumber("iid", gmcp.Required()),
		gmcp.WithString("body", gmcp.Required()),
		gmcp.WithString("project_id"),
	}
	register("gitlab_create_mr_note", "Create note/comment on merge request.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return CreateMRNote(ctx, c, args)
		},
		createMRNoteOpts...,
	)

	listIssuesOpts := []gmcp.ToolOption{
		gmcp.WithString("state"),
		gmcp.WithString("search"),
		gmcp.WithString("labels"),
		gmcp.WithString("assignee_username"),
		gmcp.WithNumber("page"),
		gmcp.WithNumber("per_page"),
		gmcp.WithString("project_id"),
	}
	register("gitlab_list_issues", "List issues in project.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return ListIssues(ctx, c, args)
		},
		listIssuesOpts...,
	)

	getIssueOpts := []gmcp.ToolOption{
		gmcp.WithNumber("iid", gmcp.Required()),
		gmcp.WithString("project_id"),
	}
	register("gitlab_get_issue", "Get issue details by IID.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return GetIssue(ctx, c, args)
		},
		getIssueOpts...,
	)

	createIssueOpts := []gmcp.ToolOption{
		gmcp.WithString("title", gmcp.Required()),
		gmcp.WithString("description", gmcp.Required()),
		gmcp.WithArray("labels", mustStringArray),
		gmcp.WithString("assignee"),
		gmcp.WithString("milestone"),
		gmcp.WithString("project_id"),
	}
	register("gitlab_create_issue", "Create a GitLab issue for diagnostics.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return CreateIssue(ctx, c, args)
		},
		createIssueOpts...,
	)

	updateIssueOpts := []gmcp.ToolOption{
		gmcp.WithNumber("iid", gmcp.Required()),
		gmcp.WithString("title"),
		gmcp.WithString("description"),
		gmcp.WithString("state_event"),
		gmcp.WithArray("labels", mustStringArray),
		gmcp.WithString("project_id"),
	}
	register("gitlab_update_issue", "Update issue fields/state.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return UpdateIssue(ctx, c, args)
		},
		updateIssueOpts...,
	)

	createIssueNoteOpts := []gmcp.ToolOption{
		gmcp.WithNumber("iid", gmcp.Required()),
		gmcp.WithString("body", gmcp.Required()),
		gmcp.WithString("project_id"),
	}
	register("gitlab_create_issue_note", "Create note/comment on issue.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return CreateIssueNote(ctx, c, args)
		},
		createIssueNoteOpts...,
	)

	listIssueNotesOpts := []gmcp.ToolOption{
		gmcp.WithNumber("iid", gmcp.Required()),
		gmcp.WithString("project_id"),
	}
	register("gitlab_list_issue_notes", "List issue notes/comments.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return ListIssueNotes(ctx, c, args)
		},
		listIssueNotesOpts...,
	)

	listPipelinesOpts := []gmcp.ToolOption{
		gmcp.WithString("ref"),
		gmcp.WithString("status"),
		gmcp.WithNumber("page"),
		gmcp.WithNumber("per_page"),
		gmcp.WithString("project_id"),
	}
	register("gitlab_list_pipelines", "List pipelines for project.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return ListPipelines(ctx, c, args)
		},
		listPipelinesOpts...,
	)

	getPipelineOpts := []gmcp.ToolOption{
		gmcp.WithNumber("pipeline_id", gmcp.Required()),
		gmcp.WithString("project_id"),
	}
	register("gitlab_get_pipeline", "Get pipeline detail by ID.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return GetPipeline(ctx, c, args)
		},
		getPipelineOpts...,
	)

	retryPipelineOpts := []gmcp.ToolOption{
		gmcp.WithNumber("pipeline_id", gmcp.Required()),
		gmcp.WithString("project_id"),
	}
	register("gitlab_retry_pipeline", "Retry pipeline by ID.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return RetryPipeline(ctx, c, args)
		},
		retryPipelineOpts...,
	)

	cancelPipelineOpts := []gmcp.ToolOption{
		gmcp.WithNumber("pipeline_id", gmcp.Required()),
		gmcp.WithString("project_id"),
	}
	register("gitlab_cancel_pipeline", "Cancel pipeline by ID.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return CancelPipeline(ctx, c, args)
		},
		cancelPipelineOpts...,
	)

	listPipelineJobsOpts := []gmcp.ToolOption{
		gmcp.WithNumber("pipeline_id", gmcp.Required()),
		gmcp.WithString("project_id"),
	}
	register("gitlab_list_pipeline_jobs", "List jobs in pipeline.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return ListPipelineJobs(ctx, c, args)
		},
		listPipelineJobsOpts...,
	)

	getJobLogOpts := []gmcp.ToolOption{
		gmcp.WithNumber("job_id", gmcp.Required()),
		gmcp.WithString("project_id"),
	}
	register("gitlab_get_job_log", "Get job log trace (preview).",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return GetJobLog(ctx, c, args)
		},
		getJobLogOpts...,
	)

	listLabelsOpts := []gmcp.ToolOption{
		gmcp.WithNumber("page"),
		gmcp.WithNumber("per_page"),
		gmcp.WithString("project_id"),
	}
	register("gitlab_list_labels", "List labels in project.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return ListLabels(ctx, c, args)
		},
		listLabelsOpts...,
	)

	createLabelOpts := []gmcp.ToolOption{
		gmcp.WithString("name", gmcp.Required()),
		gmcp.WithString("color", gmcp.Required()),
		gmcp.WithString("description"),
		gmcp.WithString("project_id"),
	}
	register("gitlab_create_label", "Create project label.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return CreateLabel(ctx, c, args)
		},
		createLabelOpts...,
	)

	updateLabelOpts := []gmcp.ToolOption{
		gmcp.WithString("name", gmcp.Required()),
		gmcp.WithString("new_name"),
		gmcp.WithString("color"),
		gmcp.WithString("description"),
		gmcp.WithString("project_id"),
	}
	register("gitlab_update_label", "Update project label.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return UpdateLabel(ctx, c, args)
		},
		updateLabelOpts...,
	)

	deleteLabelOpts := []gmcp.ToolOption{
		gmcp.WithString("name", gmcp.Required()),
		gmcp.WithString("project_id"),
	}
	register("gitlab_delete_label", "Delete project label by name.",
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return DeleteLabel(ctx, c, args)
		},
		deleteLabelOpts...,
	)
}
