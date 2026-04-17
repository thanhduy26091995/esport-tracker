---
phase: requirements
title: Requirements & Problem Understanding
description: Clarify the problem space, gather requirements, and define success criteria
feature: inline-player-creation
---

# Requirements & Problem Understanding

## Problem Statement

Khi user đang ở form **Record Match** hoặc **Create Tournament**, để chọn player họ cần đã có player tồn tại. Nếu player chưa tồn tại, user phải:
1. Thoát khỏi form đang làm dở
2. Vào `/users` → Tạo player
3. Quay lại form và làm lại từ đầu

Đây là friction không cần thiết, đặc biệt khi tổ chức giải đấu hoặc ghi trận mới với player mới tham gia.

**Người bị ảnh hưởng:** Tất cả user sử dụng hệ thống, đặc biệt khi onboarding player mới.

**Workaround hiện tại:** Mở tab mới sang `/users`, tạo player, quay lại tab cũ.

## Goals & Objectives

**Primary goals:**
- Cho phép tạo player mới trực tiếp từ `MatchForm` (dialog Record Match)
- Cho phép tạo player mới trực tiếp từ `CreateTournamentView` (form Create Tournament)
- Sau khi tạo, player mới hiển thị ngay trong danh sách player của context hiện tại để user chọn tiếp nếu cần

**Secondary goals:**
- Reuse component `UserForm.vue` hiện có, không tạo lại form
- Không làm phức tạp thêm flow chọn player hiện tại

**Non-goals:**
- Không cho phép edit/delete player từ các route này — chỉ tạo mới
- Không thêm inline quick-form (chỉ name) — luôn dùng full `UserForm` để đảm bảo data quality (tier, handicap)

## User Stories & Use Cases

- **US-01:** Là người tổ chức giải đấu, tôi muốn tạo player mới ngay trong form Create Tournament để không bị gián đoạn workflow.
- **US-02:** Là người ghi trận, tôi muốn tạo player mới ngay trong dialog Record Match khi có tay mới tham gia.
- **US-03:** Sau khi tạo player mới, player đó xuất hiện ngay trong danh sách player của form hiện tại để tôi có thể tiếp tục thao tác mà không rời màn hình.

**Key workflow:**
1. User mở MatchForm / CreateTournamentView
2. Trong dropdown hoặc player selector, thấy option/nút "➕ Tạo player mới"
3. Click → `UserForm` dialog mở đè lên
4. Điền tên, tier, handicap → Submit
5. `UserForm` đóng, player mới xuất hiện ngay trong danh sách player của form hiện tại
6. User tiếp tục điền form gốc

## Success Criteria

- [ ] Có thể tạo player mới từ `MatchForm` mà không rời dialog
- [ ] Có thể tạo player mới từ `CreateTournamentView` mà không rời trang
- [ ] Player mới hiển thị ngay trong danh sách player sau khi tạo
- [ ] `UserForm` được reuse nguyên vẹn, không duplicate logic
- [ ] Danh sách player trong form được refresh sau khi tạo

## Constraints & Assumptions

- Frontend-only change: backend API tạo user đã có (`POST /api/v1/users`)
- `UserForm.vue` emit `submit` với `{ name, tier, handicap_rate }` — reuse không thay đổi
- `useUserStore` đã có action `createUser` — reuse trực tiếp
- Element Plus `el-select` hỗ trợ slot custom trong dropdown

## Questions & Open Items

- [x] Option A (quick-add trong dropdown) được chọn — confirmed
- [x] MatchForm không auto-select sau khi tạo; yêu cầu là player mới phải hiển thị ngay trong danh sách để user tự chọn tiếp
- [x] Tournament participant selector dùng `el-checkbox-group`; player mới chỉ được thêm vào danh sách, không auto-check mặc định
