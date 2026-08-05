package service

import (
	"bytes"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/duyb/esport-score-tracker/internal/cache"
	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	mime "mime/multipart"
)

// mockFile wraps bytes.Reader to satisfy multipart.File without touching disk.
type mockFile struct {
	*bytes.Reader
}

func (m *mockFile) Close() error { return nil }

func newMockFile(data []byte) *mockFile {
	return &mockFile{bytes.NewReader(data)}
}

// ─── validClubSlugs completeness ─────────────────────────────────────────────

func TestValidClubSlugs_AllLeaguesPresent(t *testing.T) {
	expected := []string{
		// Premier League
		"man-city", "liverpool", "man-utd", "chelsea", "arsenal",
		"spurs", "newcastle", "aston-villa", "west-ham", "everton",
		// La Liga
		"real-madrid", "barcelona", "atletico", "sevilla", "betis",
		"valencia", "villarreal",
		// Bundesliga
		"bayern", "dortmund", "rb-leipzig", "leverkusen", "frankfurt", "gladbach",
		// Serie A
		"juventus", "inter", "ac-milan", "napoli",
		"roma", "lazio", "atalanta", "fiorentina",
		// Ligue 1
		"psg", "marseille", "lyon", "monaco", "lille",
		// Others
		"porto", "benfica", "ajax", "flamengo",
		// Default sentinel
		"none",
	}
	for _, slug := range expected {
		assert.True(t, validClubSlugs[slug], "validClubSlugs is missing %q", slug)
	}
}

// ─── UpdateClub — slug validation (no DB required) ───────────────────────────

func TestUpdateClub_UnknownSlug_ReturnsError(t *testing.T) {
	svc := &UserService{}
	err := svc.UpdateClub(uuid.New(), "fake-club")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown club slug")
}

func TestUpdateClub_MalformedSlugs_AllRejected(t *testing.T) {
	svc := &UserService{}
	cases := []string{"FC Barcelona", "REAL_MADRID", "man city", "???", "real madrid"}
	for _, slug := range cases {
		err := svc.UpdateClub(uuid.New(), slug)
		require.Errorf(t, err, "expected error for slug %q", slug)
		assert.Contains(t, err.Error(), "unknown club slug", "slug: %q", slug)
	}
}

// ─── allowedAvatarMIME map — detection coverage ──────────────────────────────

func TestAllowedAvatarMIME_AcceptsJPEG(t *testing.T) {
	magic := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F'}
	detected := http.DetectContentType(magic)
	_, ok := allowedAvatarMIME[detected]
	assert.True(t, ok, "JPEG magic bytes should map to an allowed MIME, got %q", detected)
}

func TestAllowedAvatarMIME_AcceptsPNG(t *testing.T) {
	magic := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	detected := http.DetectContentType(magic)
	_, ok := allowedAvatarMIME[detected]
	assert.True(t, ok, "PNG magic bytes should map to an allowed MIME, got %q", detected)
}

func TestAllowedAvatarMIME_AcceptsGIF(t *testing.T) {
	magic := []byte("GIF89a\x01\x00\x01\x00\x00\xff\x00")
	detected := http.DetectContentType(magic)
	_, ok := allowedAvatarMIME[detected]
	assert.True(t, ok, "GIF magic bytes should map to an allowed MIME, got %q", detected)
}

func TestAllowedAvatarMIME_AcceptsWebP(t *testing.T) {
	// Full RIFF+WEBP+VP8 header required for Go's sniffer to identify image/webp.
	magic := []byte{
		'R', 'I', 'F', 'F', // RIFF marker
		0x24, 0x00, 0x00, 0x00, // file size (arbitrary)
		'W', 'E', 'B', 'P', // WEBP marker
		'V', 'P', '8', ' ', // VP8 chunk type (lossy)
		0x10, 0x00, 0x00, 0x00, // chunk size
	}
	detected := http.DetectContentType(magic)
	_, ok := allowedAvatarMIME[detected]
	assert.True(t, ok, "WebP magic bytes should map to an allowed MIME, got %q", detected)
}

func TestAllowedAvatarMIME_RejectsSVG(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect width="100" height="100"/></svg>`)
	detected := http.DetectContentType(svg)
	_, ok := allowedAvatarMIME[detected]
	assert.False(t, ok, "SVG should not be in allowedAvatarMIME, got %q", detected)
}

func TestAllowedAvatarMIME_RejectsPDF(t *testing.T) {
	pdf := []byte("%PDF-1.4 fake content")
	detected := http.DetectContentType(pdf)
	_, ok := allowedAvatarMIME[detected]
	assert.False(t, ok, "PDF should not be in allowedAvatarMIME, got %q", detected)
}

// ─── UploadAvatar — size & MIME guards (no DB required) ──────────────────────

func TestUploadAvatar_FileTooLarge_ReturnsError(t *testing.T) {
	svc := &UserService{}
	header := &mime.FileHeader{Size: 6 << 20} // 6 MB > 5 MB limit
	_, err := svc.UploadAvatar(uuid.New(), newMockFile(nil), header)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too large")
}

func TestUploadAvatar_ExactlyAtSizeLimit_PassesSizeGuard(t *testing.T) {
	// Exactly 5 MB is within the limit. The call proceeds to MIME detection on
	// empty bytes and fails there — confirming the size guard is boundary-correct.
	svc := &UserService{}
	header := &mime.FileHeader{Size: 5 << 20}
	_, err := svc.UploadAvatar(uuid.New(), newMockFile(nil), header)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "too large", "size guard must not fire at exactly 2 MB")
	assert.Contains(t, err.Error(), "unsupported file type")
}

func TestUploadAvatar_UnsupportedMIME_SVG_ReturnsError(t *testing.T) {
	svc := &UserService{}
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><circle r="50"/></svg>`)
	header := &mime.FileHeader{Size: int64(len(svg))}
	_, err := svc.UploadAvatar(uuid.New(), newMockFile(svg), header)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file type")
}

func TestUploadAvatar_UnsupportedMIME_PDF_ReturnsError(t *testing.T) {
	svc := &UserService{}
	pdf := []byte("%PDF-1.4 fake content that looks like a PDF header")
	header := &mime.FileHeader{Size: int64(len(pdf))}
	_, err := svc.UploadAvatar(uuid.New(), newMockFile(pdf), header)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file type")
}

// ─── Integration tests (require TEST_DATABASE_URL) ───────────────────────────

func openAvatarTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping avatar/club integration tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("cannot connect to test DB (%v) — skipping", err)
	}
	require.NoError(t, db.AutoMigrate(&model.User{}))
	return db
}

func seedAvatarTestUser(t *testing.T, db *gorm.DB) *model.User {
	t.Helper()
	u := &model.User{
		ID:       uuid.New(),
		Name:     "AvatarTest-" + uuid.New().String()[:8],
		Tier:     "normal",
		IsActive: true,
	}
	require.NoError(t, db.Create(u).Error)
	t.Cleanup(func() { db.Unscoped().Delete(u) })
	return u
}

// newUserServiceForTest builds a UserService backed by db. A cache store is mandatory:
// mutating methods invalidate the user cache, and a nil store panics.
func newUserServiceForTest(db *gorm.DB) *UserService {
	return &UserService{
		repo:  repository.NewUserRepository(db),
		cache: cache.NewGoCacheStore(time.Minute, time.Minute),
	}
}

func TestUpdateClub_ValidSlug_Persisted(t *testing.T) {
	db := openAvatarTestDB(t)
	svc := newUserServiceForTest(db)
	user := seedAvatarTestUser(t, db)

	require.NoError(t, svc.UpdateClub(user.ID, "liverpool"))

	updated, err := svc.repo.GetByID(user.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.FavoriteClub)
	assert.Equal(t, "liverpool", *updated.FavoriteClub)
}

func TestUpdateClub_EmptyString_ClearsClub(t *testing.T) {
	db := openAvatarTestDB(t)
	svc := newUserServiceForTest(db)
	user := seedAvatarTestUser(t, db)

	require.NoError(t, svc.UpdateClub(user.ID, "barcelona"))
	require.NoError(t, svc.UpdateClub(user.ID, ""))

	updated, err := svc.repo.GetByID(user.ID)
	require.NoError(t, err)
	if updated.FavoriteClub != nil {
		assert.Equal(t, "", *updated.FavoriteClub, "clearing club should result in empty or null")
	}
}

func TestUpdateClub_NewLeagueSlugs_AllAccepted(t *testing.T) {
	db := openAvatarTestDB(t)
	svc := newUserServiceForTest(db)
	user := seedAvatarTestUser(t, db)

	// These are the slugs that were recently added — regression check.
	newSlugs := []string{
		"spurs", "newcastle", "aston-villa", "west-ham", "everton",
		"sevilla", "betis", "valencia", "villarreal",
		"rb-leipzig", "leverkusen", "frankfurt", "gladbach",
		"roma", "lazio", "atalanta", "fiorentina",
		"marseille", "lyon", "monaco", "lille",
	}
	for _, slug := range newSlugs {
		err := svc.UpdateClub(user.ID, slug)
		assert.NoErrorf(t, err, "slug %q should be accepted", slug)
	}
}

func TestUpdateClub_UnknownUser_ReturnsNotFound(t *testing.T) {
	db := openAvatarTestDB(t)
	svc := newUserServiceForTest(db)

	err := svc.UpdateClub(uuid.New(), "real-madrid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUploadAvatar_ValidJPEG_WritesFileAndPersistsURL(t *testing.T) {
	db := openAvatarTestDB(t)
	svc := newUserServiceForTest(db)
	user := seedAvatarTestUser(t, db)

	// Minimal JPEG (SOI marker + APP0 + EOI)
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0xFF, 0xD9}
	header := &mime.FileHeader{Filename: "test.jpg", Size: int64(len(jpegData))}

	url, err := svc.UploadAvatar(user.ID, newMockFile(jpegData), header)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(url, "/uploads/avatars/"), "URL must be under /uploads/avatars/")
	assert.True(t, strings.HasSuffix(url, ".jpg"), "URL must end in .jpg")

	filePath := strings.TrimPrefix(url, "/")
	t.Cleanup(func() { os.Remove(filePath) })

	_, err = os.Stat(filePath)
	require.NoError(t, err, "uploaded file must exist on disk")

	updated, err := svc.repo.GetByID(user.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.AvatarURL)
	assert.Equal(t, url, *updated.AvatarURL)
}

func TestUploadAvatar_SecondUpload_DeletesPreviousFile(t *testing.T) {
	db := openAvatarTestDB(t)
	svc := newUserServiceForTest(db)
	user := seedAvatarTestUser(t, db)

	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0xFF, 0xD9}
	mkHeader := func() *mime.FileHeader {
		return &mime.FileHeader{Filename: "test.jpg", Size: int64(len(jpegData))}
	}

	firstURL, err := svc.UploadAvatar(user.ID, newMockFile(jpegData), mkHeader())
	require.NoError(t, err)

	secondURL, err := svc.UploadAvatar(user.ID, newMockFile(jpegData), mkHeader())
	require.NoError(t, err)
	assert.NotEqual(t, firstURL, secondURL, "each upload must generate a unique filename")

	t.Cleanup(func() { os.Remove(strings.TrimPrefix(secondURL, "/")) })

	firstPath := strings.TrimPrefix(firstURL, "/")
	_, statErr := os.Stat(firstPath)
	assert.True(t, os.IsNotExist(statErr), "first avatar file must be deleted after second upload")
}

func TestDeleteAvatar_RemovesFileAndClearsURL(t *testing.T) {
	db := openAvatarTestDB(t)
	svc := newUserServiceForTest(db)
	user := seedAvatarTestUser(t, db)

	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0xFF, 0xD9}
	header := &mime.FileHeader{Filename: "test.jpg", Size: int64(len(jpegData))}

	url, err := svc.UploadAvatar(user.ID, newMockFile(jpegData), header)
	require.NoError(t, err)
	filePath := strings.TrimPrefix(url, "/")

	require.NoError(t, svc.DeleteAvatar(user.ID))

	_, statErr := os.Stat(filePath)
	assert.True(t, os.IsNotExist(statErr), "avatar file must be removed after DeleteAvatar")

	updated, err := svc.repo.GetByID(user.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.AvatarURL, "avatar_url must be null after DeleteAvatar")
}
