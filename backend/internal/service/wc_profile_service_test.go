package service

import (
	"testing"

	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newProfileSvc(db *gorm.DB) *WcProfileService {
	return NewWcProfileService(repository.NewWcUserRepository(db))
}

func strPtr(s string) *string { return &s }

// ─── GetProfile ───────────────────────────────────────────────────────────────

func TestGetProfile_Found(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "ProfGet_"+uuid.NewString()[:6], "pass")
	svc := newProfileSvc(db)

	profile, err := svc.GetProfile(user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, profile.ID)
	assert.Equal(t, user.Name, profile.Name)
}

func TestGetProfile_NotFound(t *testing.T) {
	db := openWcTestDB(t)
	svc := newProfileSvc(db)

	_, err := svc.GetProfile(uuid.New())
	assert.Error(t, err)
}

// ─── UpdateProfile ────────────────────────────────────────────────────────────

func TestUpdateProfile_NameOnly(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "ProfName_"+uuid.NewString()[:6], "pass")
	svc := newProfileSvc(db)

	newName := "ProfRenamed_" + uuid.NewString()[:6]
	updated, err := svc.UpdateProfile(user.ID, strPtr(newName), nil)
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)
	assert.Nil(t, updated.AvatarURL, "avatar should remain nil when not provided")
}

func TestUpdateProfile_AvatarOnly(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "ProfAvatar_"+uuid.NewString()[:6], "pass")
	svc := newProfileSvc(db)

	avatarURL := "https://example.com/avatar.jpg"
	updated, err := svc.UpdateProfile(user.ID, nil, strPtr(avatarURL))
	require.NoError(t, err)
	assert.Equal(t, user.Name, updated.Name, "name must be unchanged")
	require.NotNil(t, updated.AvatarURL)
	assert.Equal(t, avatarURL, *updated.AvatarURL)
}

func TestUpdateProfile_BothFields(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "ProfBoth_"+uuid.NewString()[:6], "pass")
	svc := newProfileSvc(db)

	newName := "ProfBothNew_" + uuid.NewString()[:6]
	updated, err := svc.UpdateProfile(user.ID, strPtr(newName), strPtr("https://pic.com/x.png"))
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)
	require.NotNil(t, updated.AvatarURL)
	assert.Equal(t, "https://pic.com/x.png", *updated.AvatarURL)
}

func TestUpdateProfile_EmptyBody_ReturnsCurrentProfile(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "ProfEmpty_"+uuid.NewString()[:6], "pass")
	svc := newProfileSvc(db)

	updated, err := svc.UpdateProfile(user.ID, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, user.Name, updated.Name, "profile must be returned unchanged when no fields provided")
}

func TestUpdateProfile_NameTaken(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	takenName := "ProfTaken_" + uuid.NewString()[:6]
	seedWcUser(t, authSvc, takenName, "pass")

	user := seedWcUser(t, authSvc, "ProfOwner_"+uuid.NewString()[:6], "pass")
	svc := newProfileSvc(db)

	_, err := svc.UpdateProfile(user.ID, strPtr(takenName), nil)
	assert.ErrorIs(t, err, ErrNameTaken)
}

func TestUpdateProfile_NameTooShort(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "ProfShort_"+uuid.NewString()[:6], "pass")
	svc := newProfileSvc(db)

	_, err := svc.UpdateProfile(user.ID, strPtr("x"), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least 2 characters")
}

func TestUpdateProfile_NameWithWhitespace_Trimmed(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "ProfTrim_"+uuid.NewString()[:6], "pass")
	svc := newProfileSvc(db)

	updated, err := svc.UpdateProfile(user.ID, strPtr("  TrimmedName  "), nil)
	require.NoError(t, err)
	assert.Equal(t, "TrimmedName", updated.Name)
}
