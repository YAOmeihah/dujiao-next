package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newPostServiceForTest(t *testing.T) (*PostService, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:post_service_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&models.PostCategory{}, &models.Post{}, &models.PostProduct{}); err != nil {
		t.Fatalf("auto migrate post tables failed: %v", err)
	}

	postRepo := repository.NewPostRepository(db)
	postCategoryRepo := repository.NewPostCategoryRepository(db)
	return NewPostService(postRepo, postCategoryRepo), db
}

func localizedPostText(value string) map[string]interface{} {
	return map[string]interface{}{"zh-CN": value}
}

func createPostCategoryFixture(t *testing.T, db *gorm.DB, slug string, parentID *uint) models.PostCategory {
	t.Helper()

	category := models.PostCategory{
		ParentID: parentID,
		Slug:     slug,
		NameJSON: models.JSON{
			"zh-CN": slug,
		},
		IsActive: true,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create post category fixture failed: %v", err)
	}
	return category
}

func createPostFixture(t *testing.T, db *gorm.DB, slug string, postType string, categoryID *uint) models.Post {
	t.Helper()

	post := models.Post{
		Slug:       slug,
		Type:       postType,
		TitleJSON:  models.JSON{"zh-CN": slug},
		CategoryID: categoryID,
	}
	if err := db.Create(&post).Error; err != nil {
		t.Fatalf("create post fixture failed: %v", err)
	}
	return post
}

func TestPostServiceHomePopupNoticeIsExclusive(t *testing.T) {
	svc, db := newPostServiceForTest(t)
	publish := true

	first, err := svc.Create(CreatePostInput{
		Slug:        "first-notice",
		Type:        constants.PostTypeNotice,
		TitleJSON:   localizedPostText("First notice"),
		SummaryJSON: localizedPostText("First summary"),
		ContentJSON: localizedPostText("First content"),
		IsPublished: &publish,
		IsHomePopup: true,
	})
	if err != nil {
		t.Fatalf("create first popup notice failed: %v", err)
	}

	second, err := svc.Create(CreatePostInput{
		Slug:        "second-notice",
		Type:        constants.PostTypeNotice,
		TitleJSON:   localizedPostText("Second notice"),
		SummaryJSON: localizedPostText("Second summary"),
		ContentJSON: localizedPostText("Second content"),
		IsPublished: &publish,
		IsHomePopup: true,
	})
	if err != nil {
		t.Fatalf("create second popup notice failed: %v", err)
	}

	var posts []models.Post
	if err := db.Order("id asc").Find(&posts).Error; err != nil {
		t.Fatalf("list posts failed: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}
	if posts[0].ID != first.ID || posts[0].IsHomePopup {
		t.Fatalf("expected first notice %d to be cleared as popup, got id=%d is_home_popup=%v", first.ID, posts[0].ID, posts[0].IsHomePopup)
	}
	if posts[1].ID != second.ID || !posts[1].IsHomePopup {
		t.Fatalf("expected second notice %d to be popup, got id=%d is_home_popup=%v", second.ID, posts[1].ID, posts[1].IsHomePopup)
	}
}

func TestPostServiceHomePopupOnlyReturnsPublishedNotice(t *testing.T) {
	svc, _ := newPostServiceForTest(t)
	draft := false
	published := true

	if _, err := svc.Create(CreatePostInput{
		Slug:        "draft-popup",
		Type:        constants.PostTypeNotice,
		TitleJSON:   localizedPostText("Draft popup"),
		SummaryJSON: localizedPostText("Draft summary"),
		ContentJSON: localizedPostText("Draft content"),
		IsPublished: &draft,
		IsHomePopup: true,
	}); err != nil {
		t.Fatalf("create draft popup failed: %v", err)
	}

	popup, err := svc.GetPublicHomePopupNotice()
	if err != nil {
		t.Fatalf("get popup notice failed: %v", err)
	}
	if popup != nil {
		t.Fatalf("expected draft popup to be hidden, got %s", popup.Slug)
	}

	expected, err := svc.Create(CreatePostInput{
		Slug:        "published-popup",
		Type:        constants.PostTypeNotice,
		TitleJSON:   localizedPostText("Published popup"),
		SummaryJSON: localizedPostText("Published summary"),
		ContentJSON: localizedPostText("Published content"),
		IsPublished: &published,
		IsHomePopup: true,
	})
	if err != nil {
		t.Fatalf("create published popup failed: %v", err)
	}

	popup, err = svc.GetPublicHomePopupNotice()
	if err != nil {
		t.Fatalf("get published popup notice failed: %v", err)
	}
	if popup == nil || popup.ID != expected.ID {
		t.Fatalf("expected published popup %d, got %#v", expected.ID, popup)
	}
}

func TestPostServiceCreateRejectsNoticeCategory(t *testing.T) {
	svc, db := newPostServiceForTest(t)
	leaf := createPostCategoryFixture(t, db, "announcements", nil)

	_, err := svc.Create(CreatePostInput{
		Slug:       "notice-with-category",
		Type:       constants.PostTypeNotice,
		TitleJSON:  map[string]interface{}{"zh-CN": "notice-with-category"},
		CategoryID: &leaf.ID,
	})
	if err != ErrPostNoticeCategoryUnsupported {
		t.Fatalf("expected ErrPostNoticeCategoryUnsupported, got %v", err)
	}
}

func TestPostServiceCreateRejectsMissingOrParentCategory(t *testing.T) {
	svc, db := newPostServiceForTest(t)
	parent := createPostCategoryFixture(t, db, "blog", nil)
	_ = createPostCategoryFixture(t, db, "backend", &parent.ID)

	missingID := uint(9999)
	_, err := svc.Create(CreatePostInput{
		Slug:       "blog-missing-category",
		Type:       constants.PostTypeBlog,
		TitleJSON:  map[string]interface{}{"zh-CN": "blog-missing-category"},
		CategoryID: &missingID,
	})
	if err != ErrPostCategoryInvalid {
		t.Fatalf("expected ErrPostCategoryInvalid for missing category, got %v", err)
	}

	_, err = svc.Create(CreatePostInput{
		Slug:       "blog-parent-category",
		Type:       constants.PostTypeBlog,
		TitleJSON:  map[string]interface{}{"zh-CN": "blog-parent-category"},
		CategoryID: &parent.ID,
	})
	if err != ErrPostCategoryInvalid {
		t.Fatalf("expected ErrPostCategoryInvalid for parent category, got %v", err)
	}
}

func TestPostServiceUpdateRejectsUnsupportedOrInvalidCategoryAssignment(t *testing.T) {
	svc, db := newPostServiceForTest(t)
	parent := createPostCategoryFixture(t, db, "blog", nil)
	leaf := createPostCategoryFixture(t, db, "backend", &parent.ID)
	post := createPostFixture(t, db, "service-post", constants.PostTypeBlog, &leaf.ID)

	_, err := svc.Update(fmt.Sprintf("%d", post.ID), CreatePostInput{
		Slug:       post.Slug,
		Type:       constants.PostTypeNotice,
		TitleJSON:  map[string]interface{}{"zh-CN": post.Slug},
		CategoryID: &leaf.ID,
	})
	if err != ErrPostNoticeCategoryUnsupported {
		t.Fatalf("expected ErrPostNoticeCategoryUnsupported on notice update, got %v", err)
	}

	_, err = svc.Update(fmt.Sprintf("%d", post.ID), CreatePostInput{
		Slug:       post.Slug,
		Type:       constants.PostTypeBlog,
		TitleJSON:  map[string]interface{}{"zh-CN": post.Slug},
		CategoryID: &parent.ID,
	})
	if err != ErrPostCategoryInvalid {
		t.Fatalf("expected ErrPostCategoryInvalid on parent category update, got %v", err)
	}
}

func TestPostServiceCategoryAssignmentRespectsInactive(t *testing.T) {
	svc, db := newPostServiceForTest(t)
	inactive := createPostCategoryFixture(t, db, "archived", nil)
	if err := db.Model(&models.PostCategory{}).Where("id = ?", inactive.ID).Update("is_active", false).Error; err != nil {
		t.Fatalf("deactivate category fixture failed: %v", err)
	}

	_, err := svc.Create(CreatePostInput{
		Slug:       "blog-inactive-category",
		Type:       constants.PostTypeBlog,
		TitleJSON:  map[string]interface{}{"zh-CN": "blog-inactive-category"},
		CategoryID: &inactive.ID,
	})
	if err != ErrPostCategoryInvalid {
		t.Fatalf("expected ErrPostCategoryInvalid for inactive category, got %v", err)
	}

	// 文章已有分类后来被禁用：保存时保留原分类应放行
	post := createPostFixture(t, db, "blog-keeps-category", constants.PostTypeBlog, &inactive.ID)
	updated, err := svc.Update(fmt.Sprintf("%d", post.ID), CreatePostInput{
		Slug:       post.Slug,
		Type:       constants.PostTypeBlog,
		TitleJSON:  map[string]interface{}{"zh-CN": post.Slug},
		CategoryID: &inactive.ID,
	})
	if err != nil {
		t.Fatalf("expected keeping now-inactive category to succeed, got %v", err)
	}
	if updated.CategoryID == nil || *updated.CategoryID != inactive.ID {
		t.Fatalf("expected category to be kept, got %v", updated.CategoryID)
	}
}
