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
	if err := db.AutoMigrate(&models.Post{}); err != nil {
		t.Fatalf("auto migrate post failed: %v", err)
	}

	return NewPostService(repository.NewPostRepository(db)), db
}

func localizedPostText(value string) map[string]interface{} {
	return map[string]interface{}{"zh-CN": value}
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
