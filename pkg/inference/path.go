package inference

import (
	"regexp"
	"strconv"
	"strings"
)

// Path parameter patterns
var (
	// UUID pattern: 8-4-4-4-12 hex digits
	uuidPathPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

	// Numeric ID pattern: digits only
	numericPattern = regexp.MustCompile(`^\d+$`)

	// Short hash pattern: 6-12 hex characters (like short git hashes)
	shortHashPattern = regexp.MustCompile(`^[0-9a-fA-F]{6,12}$`)

	// Long hash pattern: 32+ hex characters (like MD5, SHA1, SHA256)
	longHashPattern = regexp.MustCompile(`^[0-9a-fA-F]{32,64}$`)

	// MongoDB ObjectId: 24 hex characters
	objectIdPattern = regexp.MustCompile(`^[0-9a-fA-F]{24}$`)

	// Base64 ID pattern: alphanumeric with possible - and _
	base64IdPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{16,}$`)

	// Date pattern: YYYY-MM-DD
	datePathPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

	// Version pattern: v1, v2, v1.0, etc.
	versionPattern = regexp.MustCompile(`^v\d+(\.\d+)?$`)
)

// SegmentType represents the type of a path segment.
type SegmentType int

const (
	SegmentLiteral SegmentType = iota
	SegmentNumericID
	SegmentUUID
	SegmentHash
	SegmentObjectID
	SegmentBase64ID
	SegmentDate
	SegmentUnknownID
)

// PathInferrer handles path template inference.
type PathInferrer struct {
	// resourceNames maps parent segments to parameter names
	// e.g., "users" -> "userId", "posts" -> "postId"
	resourceNames map[string]string
}

// NewPathInferrer creates a new PathInferrer with default settings.
func NewPathInferrer() *PathInferrer {
	return &PathInferrer{
		resourceNames: map[string]string{
			// Common user-related resources
			"users":     "userId",
			"user":      "userId",
			"members":   "memberId",
			"member":    "memberId",
			"customers": "customerId",
			"customer":  "customerId",
			"employees": "employeeId",
			"employee":  "employeeId",
			"authors":   "authorId",
			"author":    "authorId",
			"owners":    "ownerId",
			"owner":     "ownerId",
			"admins":    "adminId",
			"admin":     "adminId",

			// Content resources
			"posts":    "postId",
			"post":     "postId",
			"articles": "articleId",
			"article":  "articleId",
			"comments": "commentId",
			"comment":  "commentId",
			"reviews":  "reviewId",
			"review":   "reviewId",
			"replies":  "replyId",
			"reply":    "replyId",
			"messages": "messageId",
			"message":  "messageId",
			"threads":  "threadId",
			"thread":   "threadId",
			"channels": "channelId",
			"channel":  "channelId",
			"feeds":    "feedId",
			"feed":     "feedId",
			"pages":    "pageId",
			"page":     "pageId",
			"blogs":    "blogId",
			"blog":     "blogId",

			// E-commerce resources
			"orders":        "orderId",
			"order":         "orderId",
			"products":      "productId",
			"product":       "productId",
			"items":         "itemId",
			"item":          "itemId",
			"carts":         "cartId",
			"cart":          "cartId",
			"invoices":      "invoiceId",
			"invoice":       "invoiceId",
			"payments":      "paymentId",
			"payment":       "paymentId",
			"transactions":  "transactionId",
			"transaction":   "transactionId",
			"subscriptions": "subscriptionId",
			"subscription":  "subscriptionId",
			"plans":         "planId",
			"plan":          "planId",
			"coupons":       "couponId",
			"coupon":        "couponId",
			"discounts":     "discountId",
			"discount":      "discountId",

			// Organization resources
			"accounts":      "accountId",
			"account":       "accountId",
			"organizations": "organizationId",
			"organization":  "organizationId",
			"orgs":          "orgId",
			"org":           "orgId",
			"companies":     "companyId",
			"company":       "companyId",
			"workspaces":    "workspaceId",
			"workspace":     "workspaceId",
			"tenants":       "tenantId",
			"tenant":        "tenantId",

			// Project/work resources
			"projects":    "projectId",
			"project":     "projectId",
			"tasks":       "taskId",
			"task":        "taskId",
			"issues":      "issueId",
			"issue":       "issueId",
			"tickets":     "ticketId",
			"ticket":      "ticketId",
			"milestones":  "milestoneId",
			"milestone":   "milestoneId",
			"sprints":     "sprintId",
			"sprint":      "sprintId",
			"releases":    "releaseId",
			"release":     "releaseId",
			"versions":    "versionId",
			"version":     "versionId",
			"builds":      "buildId",
			"build":       "buildId",
			"deployments": "deploymentId",
			"deployment":  "deploymentId",
			"jobs":        "jobId",
			"job":         "jobId",
			"runs":        "runId",
			"run":         "runId",
			"pipelines":   "pipelineId",
			"pipeline":    "pipelineId",

			// Team/group resources
			"teams":  "teamId",
			"team":   "teamId",
			"groups": "groupId",
			"group":  "groupId",
			"roles":  "roleId",
			"role":   "roleId",

			// File/document resources
			"files":       "fileId",
			"file":        "fileId",
			"documents":   "documentId",
			"document":    "documentId",
			"attachments": "attachmentId",
			"attachment":  "attachmentId",
			"images":      "imageId",
			"image":       "imageId",
			"assets":      "assetId",
			"asset":       "assetId",
			"media":       "mediaId",
			"folders":     "folderId",
			"folder":      "folderId",
			"directories": "directoryId",
			"directory":   "directoryId",

			// Event/notification resources
			"notifications": "notificationId",
			"notification":  "notificationId",
			"events":        "eventId",
			"event":         "eventId",
			"webhooks":      "webhookId",
			"webhook":       "webhookId",
			"alerts":        "alertId",
			"alert":         "alertId",
			"logs":          "logId",
			"log":           "logId",

			// Auth/session resources
			"sessions": "sessionId",
			"session":  "sessionId",
			"tokens":   "tokenId",
			"token":    "tokenId",
			"keys":     "keyId",
			"key":      "keyId",
			"secrets":  "secretId",
			"secret":   "secretId",

			// Classification resources
			"categories": "categoryId",
			"category":   "categoryId",
			"tags":       "tagId",
			"tag":        "tagId",
			"labels":     "labelId",
			"label":      "labelId",
			"types":      "typeId",
			"type":       "typeId",
			"statuses":   "statusId",
			"status":     "statusId",

			// Location resources
			"locations":  "locationId",
			"location":   "locationId",
			"addresses":  "addressId",
			"address":    "addressId",
			"regions":    "regionId",
			"region":     "regionId",
			"countries":  "countryId",
			"country":    "countryId",
			"cities":     "cityId",
			"city":       "cityId",
			"stores":     "storeId",
			"store":      "storeId",
			"warehouses": "warehouseId",
			"warehouse":  "warehouseId",

			// API/integration resources
			"apis":         "apiId",
			"api":          "apiId",
			"endpoints":    "endpointId",
			"endpoint":     "endpointId",
			"integrations": "integrationId",
			"integration":  "integrationId",
			"connections":  "connectionId",
			"connection":   "connectionId",
			"apps":         "appId",
			"app":          "appId",
			"applications": "applicationId",
			"application":  "applicationId",
			"services":     "serviceId",
			"service":      "serviceId",
			"resources":    "resourceId",
			"resource":     "resourceId",

			// Repository resources
			"repositories": "repositoryId",
			"repository":   "repositoryId",
			"repos":        "repoId",
			"repo":         "repoId",
			"branches":     "branchId",
			"branch":       "branchId",
			"commits":      "commitId",
			"commit":       "commitId",
			"pulls":        "pullId",
			"pull":         "pullId",
			"merges":       "mergeId",
			"merge":        "mergeId",

			// Database resources
			"databases":   "databaseId",
			"database":    "databaseId",
			"tables":      "tableId",
			"table":       "tableId",
			"collections": "collectionId",
			"collection":  "collectionId",
			"records":     "recordId",
			"record":      "recordId",
			"entries":     "entryId",
			"entry":       "entryId",
			"rows":        "rowId",
			"row":         "rowId",

			// Metrics/analytics resources
			"metrics":    "metricId",
			"metric":     "metricId",
			"reports":    "reportId",
			"report":     "reportId",
			"dashboards": "dashboardId",
			"dashboard":  "dashboardId",
			"charts":     "chartId",
			"chart":      "chartId",
			"widgets":    "widgetId",
			"widget":     "widgetId",

			// Settings/config resources
			"settings":       "settingId",
			"setting":        "settingId",
			"preferences":    "preferenceId",
			"preference":     "preferenceId",
			"configurations": "configurationId",
			"configuration":  "configurationId",
			"configs":        "configId",
			"config":         "configId",
			"options":        "optionId",
			"option":         "optionId",
			"features":       "featureId",
			"feature":        "featureId",
			"flags":          "flagId",
			"flag":           "flagId",
		},
	}
}

// InferTemplate converts a concrete path to a parameterized template.
// Returns the template and extracted parameter values.
func (p *PathInferrer) InferTemplate(path string) (template string, params map[string]string) {
	params = make(map[string]string)

	// Remove query string if present
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// Split path into segments
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) == 0 || (len(segments) == 1 && segments[0] == "") {
		return "/", params
	}

	result := make([]string, len(segments))
	paramCounts := make(map[string]int) // Track param name usage to avoid duplicates

	for i, segment := range segments {
		if segment == "" {
			result[i] = ""
			continue
		}

		segType := p.classifySegment(segment)

		if segType == SegmentLiteral {
			result[i] = segment
			continue
		}

		// Determine parameter name
		paramName := p.inferParamName(segments, i, segType, paramCounts)
		paramCounts[paramName]++

		// Store the actual value
		params[paramName] = segment

		// Replace with parameter placeholder
		result[i] = "{" + paramName + "}"
	}

	template = "/" + strings.Join(result, "/")
	return template, params
}

// classifySegment determines the type of a path segment.
func (p *PathInferrer) classifySegment(segment string) SegmentType {
	// Check for version patterns first (these should stay literal)
	if versionPattern.MatchString(segment) {
		return SegmentLiteral
	}

	// Check for known patterns
	switch {
	case uuidPathPattern.MatchString(segment):
		return SegmentUUID
	case objectIdPattern.MatchString(segment):
		return SegmentObjectID
	case numericPattern.MatchString(segment):
		// Only treat as ID if it looks like an ID (not too short)
		if len(segment) >= 1 {
			return SegmentNumericID
		}
		return SegmentLiteral
	case longHashPattern.MatchString(segment):
		return SegmentHash
	case shortHashPattern.MatchString(segment):
		// Short hashes might be IDs or literal paths
		// Be conservative - only if 8+ chars
		if len(segment) >= 8 {
			return SegmentHash
		}
		return SegmentLiteral
	case datePathPattern.MatchString(segment):
		return SegmentDate
	case base64IdPattern.MatchString(segment):
		return SegmentBase64ID
	default:
		// Check if it looks like a slug with numbers (e.g., "post-123-title")
		if looksLikeIDSegment(segment) {
			return SegmentUnknownID
		}
		return SegmentLiteral
	}
}

// looksLikeIDSegment checks if a segment might be a dynamic ID.
func looksLikeIDSegment(segment string) bool {
	// Contains mostly numbers
	numCount := 0
	for _, c := range segment {
		if c >= '0' && c <= '9' {
			numCount++
		}
	}
	// If more than 50% numbers and length > 4, probably an ID
	if len(segment) > 4 && float64(numCount)/float64(len(segment)) > 0.5 {
		return true
	}
	return false
}

// inferParamName determines the parameter name based on context.
func (p *PathInferrer) inferParamName(segments []string, idx int, segType SegmentType, counts map[string]int) string {
	// Try to get name from previous segment (resource name)
	if idx > 0 {
		prevSegment := strings.ToLower(segments[idx-1])
		if paramName, ok := p.resourceNames[prevSegment]; ok {
			if counts[paramName] > 0 {
				return paramName + strconv.Itoa(counts[paramName]+1)
			}
			return paramName
		}

		// Generate name from previous segment
		singular := singularize(prevSegment)
		paramName := singular + "Id"
		if counts[paramName] > 0 {
			return paramName + strconv.Itoa(counts[paramName]+1)
		}
		return paramName
	}

	// Fallback based on segment type
	switch segType {
	case SegmentUUID:
		name := "uuid"
		if counts[name] > 0 {
			return name + strconv.Itoa(counts[name]+1)
		}
		return name
	case SegmentDate:
		name := "date"
		if counts[name] > 0 {
			return name + strconv.Itoa(counts[name]+1)
		}
		return name
	default:
		name := "id"
		if counts[name] > 0 {
			return name + strconv.Itoa(counts[name]+1)
		}
		return name
	}
}

// singularize attempts to convert a plural word to singular.
// This is a simple implementation - not comprehensive.
func singularize(word string) string {
	if len(word) < 2 {
		return word
	}

	// Common patterns
	switch {
	case strings.HasSuffix(word, "ies"):
		return word[:len(word)-3] + "y"
	case strings.HasSuffix(word, "es"):
		// Check for special cases
		if strings.HasSuffix(word, "sses") || strings.HasSuffix(word, "shes") ||
			strings.HasSuffix(word, "ches") || strings.HasSuffix(word, "xes") {
			return word[:len(word)-2]
		}
		return word[:len(word)-1]
	case strings.HasSuffix(word, "s") && !strings.HasSuffix(word, "ss"):
		return word[:len(word)-1]
	default:
		return word
	}
}

// NormalizePath normalizes a path for comparison.
// Removes trailing slashes and lowercases.
func NormalizePath(path string) string {
	// Remove query string
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// Remove trailing slash (but keep leading)
	path = strings.TrimSuffix(path, "/")

	// Ensure leading slash
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}

// EndpointKey creates a unique key for an endpoint (method + path template).
func EndpointKey(method, pathTemplate string) string {
	return strings.ToUpper(method) + " " + pathTemplate
}

// InferPathTemplate is a convenience function for inferring path templates.
// It creates a new PathInferrer and calls InferTemplate.
func InferPathTemplate(path string) (template string, params map[string]string) {
	inferrer := NewPathInferrer()
	return inferrer.InferTemplate(path)
}
