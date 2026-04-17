---
phase: planning
title: Project Planning & Task Breakdown
description: Task breakdown for inline player creation feature
feature: inline-player-creation
---

# Project Planning & Task Breakdown

## Milestones

- [x] Milestone 1: MatchForm supports inline player creation
- [x] Milestone 2: CreateTournamentView supports inline player creation
- [ ] Milestone 3: I18n keys added, build passes, runtime verified

## Task Breakdown

### Phase 1: MatchForm

- [x] Task 1.1: Thêm `UserForm` import và state `showQuickCreatePlayer`, `quickCreateTarget` vào `MatchForm.vue`
- [x] Task 1.2: Thêm slot `#footer` vào cả 2 `el-select` (team1, team2) với nút "➕ Tạo player mới"
- [x] Task 1.3: Implement handler `handlePlayerCreated(...)` — gọi `userStore.createUser(...)`, sau đó emit event để parent refresh lại `users`
- [x] Task 1.4: Mount `<UserForm>` dialog trong `MatchForm` template
- [x] Task 1.5: Thêm fallback UI bằng nút cạnh label nếu `el-select #footer` không hoạt động ổn

### Phase 2: CreateTournamentView

- [x] Task 2.1: Thêm `UserForm` import và state `showQuickCreatePlayer` vào `CreateTournamentView.vue`
- [x] Task 2.2: Thêm nút "➕ Tạo player mới" bên dưới player checkbox list
- [x] Task 2.3: Implement handler `handlePlayerCreated(...)` — tạo player mới và refresh list để player xuất hiện trong danh sách
- [x] Task 2.4: Mount `<UserForm>` dialog trong `CreateTournamentView` template

### Phase 3: I18n & Polish

- [x] Task 3.1: Thêm key `players.quickCreate` (vi: "Tạo player mới", en: "Create new player") vào vi.json và en.json
- [x] Task 3.2: Verify build passes (`npm run build`)
- [ ] Task 3.3: Runtime walkthrough — tạo player từ MatchForm và CreateTournamentView, verify player mới hiển thị ngay trong danh sách

Status note: Task 3.3 đang chờ runtime walkthrough trên app đang chạy để xác nhận hành vi thực tế của quick-create trong MatchForm và CreateTournamentView.

## Dependencies

- `UserForm.vue` — không thay đổi, reuse trực tiếp
- `useUserStore` — `createUser()` + `fetchUsers()` đã có
- `useI18n` — đã inject ở cả hai files

## Timeline & Estimates

| Phase | Effort |
|---|---|
| Phase 1 (MatchForm) | ~35 min |
| Phase 2 (CreateTournamentView) | ~20 min |
| Phase 3 (I18n + verify) | ~10 min |

**Total:** ~1 hour

## Risks & Mitigation

| Risk | Mitigation |
|---|---|
| `el-select` slot `#footer` không render đúng với Element Plus version hiện tại | Fallback: dùng nút bên ngoài select thay vì slot footer |
| MatchForm props `users` không cập nhật lại sau khi tạo | Emit explicit refresh event lên parent ngay sau create thành công |
| Player mới bị trùng tên trong list | Backend/UserForm validation đã xử lý |
