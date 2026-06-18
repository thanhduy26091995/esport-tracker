package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

// ─── Test helpers ─────────────────────────────────────────────────────────────

// fakePayload builds a minimal idtoken.Payload for tests.
func fakePayload(sub, name, picture string) *idtoken.Payload {
	return &idtoken.Payload{
		Subject: sub,
		Claims: map[string]interface{}{
			"name":    name,
			"picture": picture,
		},
	}
}

// okVerifier returns a verifier that always returns the given payload.
func okVerifier(p *idtoken.Payload) googleVerifier {
	return func(_ context.Context, _, _ string) (*idtoken.Payload, error) {
		return p, nil
	}
}

// errVerifier returns a verifier that always fails with an error.
func errVerifier() googleVerifier {
	return func(_ context.Context, _, _ string) (*idtoken.Payload, error) {
		return nil, errors.New("token invalid")
	}
}

// newAuthSvcWithVerifier wires auth service with a custom Google verifier.
func newAuthSvcWithVerifier(db *gorm.DB, v googleVerifier) *WcAuthService {
	wcRepo := repository.NewWcRepository(db)
	wcUserRepo := repository.NewWcUserRepository(db)
	return NewWcAuthService(wcUserRepo, wcRepo).withVerifier(v)
}

// ─── GoogleLoginOrCreate ──────────────────────────────────────────────────────

func TestGoogleLoginOrCreate_ExistingLinkedUser(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	googleID := "gsub_exist_" + uuid.NewString()[:8]
	user := seedWcUserWithGoogle(t, db, authSvc, "Exist_"+uuid.NewString()[:6], googleID)

	svc := newAuthSvcWithVerifier(db, okVerifier(fakePayload(googleID, user.Name, "")))

	resp, err := svc.GoogleLoginOrCreate(context.Background(), "fake-token")
	require.NoError(t, err)
	assert.Equal(t, user.ID, resp.UserID)
	assert.True(t, resp.GoogleLinked)
	assert.NotEmpty(t, resp.Token)
}

func TestGoogleLoginOrCreate_NewPlayer_CreatesAccountAndWallet(t *testing.T) {
	db := openWcTestDB(t)
	wcRepo := repository.NewWcRepository(db)

	googleID := "gsub_new_" + uuid.NewString()[:8]
	svc := newAuthSvcWithVerifier(db, okVerifier(fakePayload(googleID, "NewGooglePlayer", "https://example.com/pic.jpg")))

	resp, err := svc.GoogleLoginOrCreate(context.Background(), "fake-token")
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, resp.UserID)
	assert.True(t, resp.GoogleLinked)
	assert.NotEmpty(t, resp.Token)

	wallet, err := wcRepo.GetWallet(resp.UserID)
	require.NoError(t, err, "wallet must be created alongside new account")
	assert.Equal(t, resp.UserID, wallet.WcUserID)

	t.Cleanup(func() {
		db.Where("wc_user_id = ?", resp.UserID).Delete(&model.WcWalletLog{})
		db.Where("wc_user_id = ?", resp.UserID).Delete(&model.WcWallet{})
		db.Delete(&model.WcUser{}, "id = ?", resp.UserID)
	})
}

func TestGoogleLoginOrCreate_NameTaken_AppendsSuffix(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	// Pre-occupy the base name
	baseName := "Suffix_" + uuid.NewString()[:6]
	seedWcUser(t, authSvc, baseName, "pass")

	googleID := "gsub_suffix_" + uuid.NewString()[:8]
	svc := newAuthSvcWithVerifier(db, okVerifier(fakePayload(googleID, baseName, "")))

	resp, err := svc.GoogleLoginOrCreate(context.Background(), "fake-token")
	require.NoError(t, err)
	assert.NotEqual(t, baseName, resp.Name, "name must be suffixed when base name is taken")

	t.Cleanup(func() {
		db.Where("wc_user_id = ?", resp.UserID).Delete(&model.WcWallet{})
		db.Delete(&model.WcUser{}, "id = ?", resp.UserID)
	})
}

func TestGoogleLoginOrCreate_NameTaken_MultipleConflicts_AppendsSuffix(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	// Pre-occupy "Multi" and "Multi1"
	base := "Multi_" + uuid.NewString()[:6]
	seedWcUser(t, authSvc, base, "pass")
	seedWcUser(t, authSvc, base+"1", "pass")

	googleID := "gsub_multi_" + uuid.NewString()[:8]
	svc := newAuthSvcWithVerifier(db, okVerifier(fakePayload(googleID, base, "")))

	resp, err := svc.GoogleLoginOrCreate(context.Background(), "fake-token")
	require.NoError(t, err)
	// Must pick base+"2" or higher
	assert.NotEqual(t, base, resp.Name)
	assert.NotEqual(t, base+"1", resp.Name)

	t.Cleanup(func() {
		db.Where("wc_user_id = ?", resp.UserID).Delete(&model.WcWallet{})
		db.Delete(&model.WcUser{}, "id = ?", resp.UserID)
	})
}

func TestGoogleLoginOrCreate_InvalidToken(t *testing.T) {
	db := openWcTestDB(t)
	svc := newAuthSvcWithVerifier(db, errVerifier())

	_, err := svc.GoogleLoginOrCreate(context.Background(), "bad-token")
	assert.ErrorIs(t, err, ErrInvalidGoogleToken)
}

// ─── LinkGoogleToAccount ──────────────────────────────────────────────────────

func TestLinkGoogleToAccount_Success(t *testing.T) {
	db := openWcTestDB(t)
	wcUserRepo := repository.NewWcUserRepository(db)
	_, authSvc := newWcServices(db)

	user := seedWcUser(t, authSvc, "ToLink_"+uuid.NewString()[:6], "pass")
	googleID := "gsub_link_" + uuid.NewString()[:8]
	svc := newAuthSvcWithVerifier(db, okVerifier(fakePayload(googleID, user.Name, "https://pic.com/a.png")))

	_, err := svc.LinkGoogleToAccount(context.Background(), user.ID, "fake-token")
	require.NoError(t, err)

	linked, err := wcUserRepo.GetByGoogleID(googleID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, linked.ID)
}

func TestLinkGoogleToAccount_AlreadyLinked(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	googleID := "gsub_already_" + uuid.NewString()[:8]
	user := seedWcUserWithGoogle(t, db, authSvc, "AlreadyLinked_"+uuid.NewString()[:6], googleID)

	// Try to link a different google_id — user already has one, so 0 rows affected
	svc := newAuthSvcWithVerifier(db, okVerifier(fakePayload("gsub_other_"+uuid.NewString()[:8], user.Name, "")))

	_, err := svc.LinkGoogleToAccount(context.Background(), user.ID, "fake-token")
	assert.ErrorIs(t, err, ErrAlreadyLinked)
}

func TestLinkGoogleToAccount_GoogleIDTakenByAnotherUser(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	googleID := "gsub_taken_" + uuid.NewString()[:8]
	seedWcUserWithGoogle(t, db, authSvc, "Owner_"+uuid.NewString()[:6], googleID)
	target := seedWcUser(t, authSvc, "Target_"+uuid.NewString()[:6], "pass")

	svc := newAuthSvcWithVerifier(db, okVerifier(fakePayload(googleID, target.Name, "")))

	_, err := svc.LinkGoogleToAccount(context.Background(), target.ID, "fake-token")
	assert.ErrorIs(t, err, ErrGoogleAlreadyLinked)
}

func TestLinkGoogleToAccount_Concurrent_OnlyOneSucceeds(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	user1 := seedWcUser(t, authSvc, "Conc1_"+uuid.NewString()[:6], "pass")
	user2 := seedWcUser(t, authSvc, "Conc2_"+uuid.NewString()[:6], "pass")

	googleID := "gsub_conc_" + uuid.NewString()[:8]
	payload := fakePayload(googleID, "ConcUser", "")
	svc := newAuthSvcWithVerifier(db, okVerifier(payload))

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, uid := range []uuid.UUID{user1.ID, user2.ID} {
		wg.Add(1)
		go func(i int, uid uuid.UUID) {
			defer wg.Done()
			_, errs[i] = svc.LinkGoogleToAccount(context.Background(), uid, "fake-token")
		}(i, uid)
	}
	wg.Wait()

	successCount := 0
	for _, err := range errs {
		if err == nil {
			successCount++
		}
	}
	assert.Equal(t, 1, successCount, "exactly one concurrent link must succeed")
}

// ─── Login: google_linked response field ──────────────────────────────────────

func TestLogin_ReturnsGoogleLinkedFalse_WhenNotLinked(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	name := "LoginUnlinked_" + uuid.NewString()[:6]
	seedWcUser(t, authSvc, name, "secret")

	resp, err := authSvc.Login(name, "secret")
	require.NoError(t, err)
	assert.False(t, resp.GoogleLinked)
}

func TestLogin_ReturnsGoogleLinkedTrue_WhenLinked(t *testing.T) {
	db := openWcTestDB(t)
	wcRepo := repository.NewWcRepository(db)

	name := "LoginLinked_" + uuid.NewString()[:6]
	googleID := "gsub_loginlinked_" + uuid.NewString()[:8]

	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), 4)
	hashStr := string(hash)
	user := &model.WcUser{Name: name, GoogleID: &googleID, PasswordHash: &hashStr}
	require.NoError(t, db.Create(user).Error)
	require.NoError(t, wcRepo.CreateWallet(db, user.ID))
	t.Cleanup(func() {
		db.Where("wc_user_id = ?", user.ID).Delete(&model.WcWalletLog{})
		db.Where("wc_user_id = ?", user.ID).Delete(&model.WcWallet{})
		db.Delete(user)
	})

	_, authSvc := newWcServices(db)
	resp, err := authSvc.Login(name, "pass")
	require.NoError(t, err)
	assert.True(t, resp.GoogleLinked)
}

func TestLogin_GoogleOnlyAccount_NoPassword_Fails(t *testing.T) {
	db := openWcTestDB(t)
	_, authSvc := newWcServices(db)

	googleID := "gsub_nopass_" + uuid.NewString()[:8]
	user := seedWcUserWithGoogle(t, db, authSvc, "NoPass_"+uuid.NewString()[:6], googleID)

	_, err := authSvc.Login(user.Name, "anything")
	assert.Error(t, err, "login with password must fail for Google-only accounts")
	assert.Contains(t, err.Error(), "invalid name or password")
}
