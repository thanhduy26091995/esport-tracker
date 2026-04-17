---
phase: testing
title: Testing Strategy & Validation
description: Test scope for inline player creation feature
feature: inline-player-creation
---

# Testing Strategy

## Scope

- Manual runtime testing (primary)
- Unit test cho handler logic (optional)

## Test Files

| File | Layer | Coverage Target |
|------|-------|----------------|
| `frontend/src/components/match/MatchForm.vue` | component | manual walkthrough |
| `frontend/src/views/CreateTournamentView.vue` | view | manual walkthrough |

## Manual Test Checklist

### MatchForm
- [ ] Mở "Record Match" dialog
- [ ] Trong dropdown Team 1, thấy nút "➕ Tạo player mới" ở footer
- [ ] Click nút → `UserForm` dialog mở
- [ ] Điền tên + tier + handicap → Submit
- [ ] `UserForm` đóng, player mới xuất hiện trong dropdown Team 1
- [ ] Player mới không bị auto-select; user tự chọn nếu cần
- [ ] Repeat cho Team 2
- [ ] Cancel `UserForm` → không có gì thay đổi

### CreateTournamentView
- [ ] Mở `/tournaments/create`
- [ ] Thấy nút "➕ Tạo player mới" bên dưới player list
- [ ] Click → `UserForm` dialog mở
- [ ] Submit → player mới xuất hiện trong checkbox list nhưng không auto-check
- [ ] Cancel → không thay đổi

## Edge Cases

- [ ] Tạo player khi list đang empty → player xuất hiện trong list để user chọn tiếp
- [ ] Match type 2v2 với team đã có 2 người → player tạo xong chỉ hiện trong list, không force select
- [ ] Tạo player trùng tên → backend trả lỗi, `UserForm` hiển thị error (xử lý bởi UserForm/userStore)
